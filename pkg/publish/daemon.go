// Copyright 2018 ko Build Authors All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package publish

import (
	archivetar "archive/tar"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/google/ko/pkg/build"
)

const (
	// LocalDomain is a sentinel "registry" that represents side-loading images into the daemon.
	LocalDomain = "ko.local"
)

// demon is intentionally misspelled to avoid name collision (and drive Jon nuts).
// [Narrator: Jon wasn't the only one driven nuts.]
type demon struct {
	base   string
	client daemon.Client
	namer  Namer
	tags   []string
}

// DaemonOption is a functional option for NewDaemon.
type DaemonOption func(*demon) error

// WithLocalDomain is a functional option for overriding the domain used for images that are side-loaded into the daemon.
func WithLocalDomain(domain string) DaemonOption {
	return func(i *demon) error {
		if domain != "" {
			i.base = domain
		}
		return nil
	}
}

// WithDockerClient is a functional option for overriding the docker client.
func WithDockerClient(client daemon.Client) DaemonOption {
	return func(i *demon) error {
		if client != nil {
			i.client = client
		}
		return nil
	}
}

// NewDaemon returns a new publish.Interface that publishes images to a container daemon.
func NewDaemon(namer Namer, tags []string, opts ...DaemonOption) (Interface, error) {
	d := &demon{
		base:  LocalDomain,
		namer: namer,
		tags:  tags,
	}
	for _, option := range opts {
		if err := option(d); err != nil {
			return nil, err
		}
	}
	return d, nil
}

func (d *demon) getOpts(ctx context.Context) []daemon.Option {
	return []daemon.Option{
		daemon.WithContext(ctx),
		daemon.WithClient(d.client),
	}
}

// Publish implements publish.Interface
func (d *demon) Publish(ctx context.Context, br build.Result, s string) (name.Reference, error) {
	s = strings.TrimPrefix(s, build.StrictScheme)
	// https://github.com/google/go-containerregistry/issues/212
	s = strings.ToLower(s)

	// There's no way to write an index to a kind, so attempt to downcast it to an image.
	var img v1.Image
	switch i := br.(type) {
	case v1.Image:
		img = i
	case v1.ImageIndex:
		im, err := i.IndexManifest()
		if err != nil {
			return nil, err
		}
		goos, goarch := os.Getenv("GOOS"), os.Getenv("GOARCH")
		if goos == "" {
			goos = "linux"
		}
		if goarch == "" {
			goarch = "amd64"
		}
		for _, manifest := range im.Manifests {
			if manifest.Platform == nil {
				continue
			}
			if manifest.Platform.OS != goos {
				continue
			}
			if manifest.Platform.Architecture != goarch {
				continue
			}
			img, err = i.Image(manifest.Digest)
			if err != nil {
				return nil, err
			}
			break
		}
		if img == nil {
			return nil, fmt.Errorf("failed to find %s/%s image in index for image: %v", goos, goarch, s)
		}
	default:
		return nil, fmt.Errorf("failed to interpret %s result as image: %v", s, br)
	}

	h, err := img.Digest()
	if err != nil {
		return nil, err
	}

	digestTag, err := name.NewTag(fmt.Sprintf("%s:%s", d.namer(d.base, s), h.Hex))
	if err != nil {
		return nil, err
	}

	log.Printf("Loading %v", digestTag)
	// Tag through daemon.Write rather than daemon.Tag. Write applies the name
	// from the image ID when the image is already loaded. Docker 29.7 on
	// Windows can import an image without the repository tag from the load
	// tarball, and tagging from "repo:tag" then fails with "No such image".
	if err := loadDaemonImage(ctx, img, digestTag, d.getOpts(ctx)); err != nil {
		return nil, err
	}
	log.Printf("Loaded %v", digestTag)

	for _, tagName := range d.tags {
		log.Printf("Adding tag %v", tagName)
		tag, err := name.NewTag(fmt.Sprintf("%s:%s", d.namer(d.base, s), tagName))
		if err != nil {
			return nil, err
		}
		if err := loadDaemonImage(ctx, img, tag, d.getOpts(ctx)); err != nil {
			return nil, fmt.Errorf("adding tag %s: %w", tagName, err)
		}
		log.Printf("Added tag %v", tagName)
	}

	return &digestTag, nil
}

// loadDaemonImage loads img into the daemon as tag. A second write tags an
// image that Docker imported by ID only. That retry must itself retain the
// tag; an ID-only result is a failure. Docker 29.7 on Windows also rejects
// archive entries named "sha256:<hex>", so that error is retried from an
// archive whose config filename has no colon.
func loadDaemonImage(ctx context.Context, img v1.Image, tag name.Tag, opts []daemon.Option) error {
	resp, err := daemon.Write(tag, img, opts...)
	if err != nil {
		log.Println("daemon.Write response:", trimDaemonResponse(resp))
		return err
	}
	applied, streamErr := daemonTagStatus(resp, tag.String())
	if applied {
		return nil
	}
	if isInvalidTarEntryName(streamErr) {
		log.Printf("Docker rejected archive entry names for %s (%v); reloading without a colon in the config filename", tag, streamErr)
		return dockerLoadSanitized(ctx, tag, img)
	}
	if streamErr != nil {
		log.Printf("Docker load of %s reported %q; tagging by image ID", tag, streamErr)
	} else {
		log.Printf("Docker loaded %s without its tag; tagging by image ID", tag)
	}
	resp, err = daemon.Write(tag, img, opts...)
	if err != nil {
		if streamErr != nil {
			return fmt.Errorf("loading %s: %w", tag, streamErr)
		}
		return err
	}
	applied, streamErr = daemonTagStatus(resp, tag.String())
	if applied {
		return nil
	}
	if streamErr != nil {
		return fmt.Errorf("tagging %s: %w", tag, streamErr)
	}
	return fmt.Errorf("tagging %s: daemon imported the image without retaining the tag", tag)
}

func isInvalidTarEntryName(err error) bool {
	return err != nil && strings.Contains(err.Error(), "invalid entry name")
}

// daemonTagStatus reports whether resp shows that tag was retained.
// An empty response is the already-present fast path, which tags by image ID.
func daemonTagStatus(resp, tag string) (bool, error) {
	if resp == "" {
		return true, nil
	}
	if err := dockerStreamError(resp); err != nil {
		return false, err
	}
	if daemonLoadDroppedTag(resp, tag) {
		return false, nil
	}
	return true, nil
}

func dockerLoadSanitized(ctx context.Context, tag name.Tag, img v1.Image) error {
	pr, pw := io.Pipe()
	writeErr := make(chan error, 1)
	go func() {
		err := writeSanitizedDockerArchive(pw, tag, img)
		writeErr <- err
		_ = pw.CloseWithError(err)
	}()

	cmd := exec.CommandContext(ctx, "docker", "load")
	cmd.Stdin = pr
	out, err := cmd.CombinedOutput()
	if werr := <-writeErr; werr != nil && err == nil {
		err = werr
	}
	text := string(out)
	if err != nil {
		return fmt.Errorf("loading %s: %w: %s", tag, err, trimDaemonResponse(text))
	}
	if strings.Contains(text, tag.String()) && !strings.Contains(text, "Loaded image ID:") {
		return nil
	}
	if strings.Contains(text, "Loaded image ID:") || !strings.Contains(text, tag.String()) {
		return fmt.Errorf("loading %s: docker load did not retain the tag: %s", tag, trimDaemonResponse(text))
	}
	return nil
}

// writeSanitizedDockerArchive writes a docker save archive whose config blob
// is not named "sha256:<hex>". Docker's Windows loader rejects that entry.
func writeSanitizedDockerArchive(w io.Writer, ref name.Reference, img v1.Image) error {
	pr, pw := io.Pipe()
	writeErr := make(chan error, 1)
	go func() {
		err := tarball.Write(ref, img, pw)
		writeErr <- err
		_ = pw.CloseWithError(err)
	}()
	err := rewriteDockerArchive(pr, w)
	if werr := <-writeErr; werr != nil && err == nil {
		err = werr
	}
	return err
}

func rewriteDockerArchive(r io.Reader, w io.Writer) error {
	tr := archivetar.NewReader(r)
	tw := archivetar.NewWriter(w)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return tw.Close()
		}
		if err != nil {
			return err
		}
		hdr.PAXRecords = nil
		hdr.Format = archivetar.FormatUnknown
		if hdr.Name == "manifest.json" {
			body, err := io.ReadAll(tr)
			if err != nil {
				return err
			}
			body, err = rewriteManifestConfigNames(body)
			if err != nil {
				return err
			}
			hdr.Name = "manifest.json"
			hdr.Size = int64(len(body))
			if err := tw.WriteHeader(hdr); err != nil {
				return err
			}
			if _, err := tw.Write(body); err != nil {
				return err
			}
			continue
		}
		hdr.Name = strings.TrimPrefix(hdr.Name, "sha256:")
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		/* #nosec G110 -- this copies an archive entry we just wrote, bounded by the image itself. */
		if _, err := io.Copy(tw, tr); err != nil {
			return err
		}
	}
}

func rewriteManifestConfigNames(b []byte) ([]byte, error) {
	var manifests []map[string]json.RawMessage
	if err := json.Unmarshal(b, &manifests); err != nil {
		return nil, fmt.Errorf("parsing docker archive manifest: %w", err)
	}
	for _, m := range manifests {
		raw, ok := m["Config"]
		if !ok {
			continue
		}
		var cfg string
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return nil, fmt.Errorf("parsing docker archive config name: %w", err)
		}
		rewritten, err := json.Marshal(strings.TrimPrefix(cfg, "sha256:"))
		if err != nil {
			return nil, err
		}
		m["Config"] = rewritten
	}
	return json.Marshal(manifests)
}

func trimDaemonResponse(resp string) string {
	resp = strings.TrimSpace(resp)
	if len(resp) > 500 {
		return resp[:500] + "..."
	}
	return resp
}

// dockerStreamError returns the first error message in a Docker API stream.
// ImageLoad uses HTTP 200 even when the stream itself reports a failure.
func dockerStreamError(resp string) error {
	if resp == "" {
		return nil
	}
	dec := json.NewDecoder(strings.NewReader(resp))
	for {
		var msg struct {
			Error string `json:"error"`
		}
		if err := dec.Decode(&msg); err == io.EOF {
			return nil
		} else if err != nil {
			return nil
		}
		if msg.Error != "" {
			return errors.New(msg.Error)
		}
	}
}

// daemonLoadDroppedTag reports whether a docker load response imported an
// image by ID and did not retain tag. An empty response means the image was
// already present and tagged by ID.
func daemonLoadDroppedTag(resp, tag string) bool {
	if resp == "" || strings.Contains(resp, tag) {
		return false
	}
	return strings.Contains(resp, "Loaded image ID:")
}

func (d *demon) Close() error {
	return nil
}
