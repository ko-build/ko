// Copyright 2026 ko Build Authors All Rights Reserved.
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
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/google/ko/pkg/build"
)

type podman struct {
	executable string
	base       string
	namer      Namer
	tags       []string
	tagOnly    bool
}

// NewPodman returns a publisher that loads images into Podman storage using the Podman CLI.
func NewPodman(base string, namer Namer, tags []string, tagOnly bool) (Interface, error) {
	if tagOnly {
		if err := validateTagOnly(tags); err != nil {
			return nil, err
		}
	}
	executable, err := exec.LookPath("podman")
	if err != nil {
		return nil, fmt.Errorf("--local-podman requires podman on PATH: %w", err)
	}
	return &podman{executable: executable, base: base, namer: namer, tags: tags, tagOnly: tagOnly}, nil
}

func (p *podman) Publish(ctx context.Context, br build.Result, s string) (name.Reference, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	img, ok := br.(v1.Image)
	if !ok {
		return nil, fmt.Errorf("--local-podman requires a single-platform image for %s; select one platform with --platform, defaultPlatforms in .ko.yaml, or KO_DEFAULTPLATFORMS", s)
	}
	s = strings.ToLower(strings.TrimPrefix(s, build.StrictScheme))
	repo := p.namer(p.base, s)
	h, err := img.Digest()
	if err != nil {
		return nil, err
	}
	// Podman can rebuild the manifest when loading a Docker archive, so return
	// an actual local tag rather than relying on the original manifest digest.
	digestTag, err := name.NewTag(fmt.Sprintf("%s:%s", repo, h.Hex))
	if err != nil {
		return nil, err
	}
	ref := digestTag
	refs := map[name.Tag]v1.Image{digestTag: img}
	for _, t := range p.tags {
		tag, err := name.NewTag(fmt.Sprintf("%s:%s", repo, t))
		if err != nil {
			return nil, err
		}
		refs[tag] = img
		if p.tagOnly {
			ref = tag
		}
	}

	f, err := os.CreateTemp("", "ko-podman-*.tar")
	if err != nil {
		return nil, fmt.Errorf("creating Podman image archive: %w", err)
	}
	defer os.Remove(f.Name())
	writeErr := tarball.MultiWrite(refs, f)
	if err := errors.Join(writeErr, f.Close()); err != nil {
		return nil, fmt.Errorf("writing Podman image archive: %w", err)
	}

	log.Printf("Loading %v into Podman", ref)
	//nolint:gosec // The executable comes from PATH and the only variable argument is a temporary file.
	cmd := exec.CommandContext(ctx, p.executable, "load", "--quiet", "--input", f.Name())
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		if msg := strings.TrimSpace(string(output)); msg != "" {
			return nil, fmt.Errorf("podman load for %v: %w: %s", ref, err, msg)
		}
		return nil, fmt.Errorf("podman load for %v: %w", ref, err)
	}
	log.Printf("Loaded %v into Podman", ref)
	return ref, nil
}

func (p *podman) Close() error { return nil }
