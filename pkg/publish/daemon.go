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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
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
	if err := loadDaemonImage(img, digestTag, d.getOpts(ctx)); err != nil {
		return nil, err
	}
	log.Printf("Loaded %v", digestTag)

	for _, tagName := range d.tags {
		log.Printf("Adding tag %v", tagName)
		tag, err := name.NewTag(fmt.Sprintf("%s:%s", d.namer(d.base, s), tagName))
		if err != nil {
			return nil, err
		}
		if err := loadDaemonImage(img, tag, d.getOpts(ctx)); err != nil {
			return nil, fmt.Errorf("adding tag %s: %w", tagName, err)
		}
		log.Printf("Added tag %v", tagName)
	}

	return &digestTag, nil
}

// loadDaemonImage loads img into the daemon as tag. If the daemon reports that
// the image was imported by ID only, it loads again so the existing image is
// tagged from its ID.
func loadDaemonImage(img v1.Image, tag name.Tag, opts []daemon.Option) error {
	resp, err := daemon.Write(tag, img, opts...)
	if err != nil {
		log.Println("daemon.Write response:", trimDaemonResponse(resp))
		return err
	}
	streamErr := dockerStreamError(resp)
	if streamErr == nil && !daemonLoadDroppedTag(resp, tag.String()) {
		return nil
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
	if err := dockerStreamError(resp); err != nil {
		return fmt.Errorf("tagging %s: %w", tag, err)
	}
	return nil
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
