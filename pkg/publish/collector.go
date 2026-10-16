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
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/ko/pkg/build"
	"github.com/sigstore/cosign/v3/pkg/oci"
)

// Published describes the outcome of publishing a single import path.
type Published struct {
	// ImportPath is the import path that was built, as passed to Publish.
	ImportPath string
	// Ref is the reference the image was published as.
	Ref name.Reference
	// Digest is the digest of the image or image index.
	Digest string
	// Platforms lists the os/arch[/variant] of each image, sorted.
	Platforms []string
}

// Collector is an Interface that remembers what its inner publisher published.
type Collector struct {
	inner Interface

	mu      sync.Mutex
	results []Published
}

var _ Interface = (*Collector)(nil)

// NewCollector wraps inner in a publisher that records a Published entry for
// every successful Publish call.
func NewCollector(inner Interface) *Collector {
	return &Collector{inner: inner}
}

// Publish implements Interface
func (c *Collector) Publish(ctx context.Context, br build.Result, ref string) (name.Reference, error) {
	result, err := c.inner.Publish(ctx, br, ref)
	if err != nil {
		return nil, err
	}
	p := Published{ImportPath: ref, Ref: result}
	if p.Digest, p.Platforms, err = describe(br, result); err != nil {
		return nil, fmt.Errorf("describing %s: %w", ref, err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.results = append(c.results, p)
	return result, nil
}

// Results returns the collected entries, sorted by import path.
func (c *Collector) Results() []Published {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := append([]Published(nil), c.results...)
	sort.Slice(out, func(i, j int) bool { return out[i].ImportPath < out[j].ImportPath })
	return out
}

// Close implements Interface
func (c *Collector) Close() error {
	return c.inner.Close()
}

// describe returns the digest and platforms of what was actually published as
// ref. That is br itself, except when the publisher could not write an index
// (daemon, kind) and loaded a single image out of it: ref then identifies that
// child image, so report it rather than the whole index.
func describe(br build.Result, ref name.Reference) (string, []string, error) {
	h, err := br.Digest()
	if err != nil {
		return "", nil, err
	}
	if idx, ok := br.(oci.SignedImageIndex); ok {
		if id := strings.TrimPrefix(ref.Identifier(), "sha256:"); id != h.Hex {
			im, err := idx.IndexManifest()
			if err != nil {
				return "", nil, err
			}
			for _, d := range im.Manifests {
				if d.Digest.Hex == id && d.Platform != nil {
					return d.Digest.String(), []string{d.Platform.String()}, nil
				}
			}
		}
	}
	platforms, err := platformsOf(br)
	return h.String(), platforms, err
}

func platformsOf(br build.Result) ([]string, error) {
	platforms := []string{}
	switch t := br.(type) {
	case oci.SignedImageIndex:
		im, err := t.IndexManifest()
		if err != nil {
			return nil, err
		}
		for _, d := range im.Manifests {
			// Skip attestations, signatures and the like.
			if d.Platform == nil || d.Platform.OS == "unknown" {
				continue
			}
			platforms = append(platforms, d.Platform.String())
		}
	case v1.Image:
		cf, err := t.ConfigFile()
		if err != nil {
			return nil, err
		}
		p := v1.Platform{OS: cf.OS, Architecture: cf.Architecture, Variant: cf.Variant}
		platforms = append(platforms, p.String())
	}
	sort.Strings(platforms)
	return platforms, nil
}
