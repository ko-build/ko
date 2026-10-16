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
	"sort"
	"strings"
	"sync"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/ko/pkg/build"
	"github.com/sigstore/cosign/v3/pkg/oci"
	ociremote "github.com/sigstore/cosign/v3/pkg/oci/remote"
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
	// SBOMs lists the SBOMs attached to what was published: the top-level
	// entity first, then one per platform image of an index.
	SBOMs []SBOM
}

// SBOM describes an SBOM attached to a published image or image index.
type SBOM struct {
	// Platform is the os/arch[/variant] of the image the SBOM belongs to. It is
	// empty for the top-level entity (the index, or a single image).
	Platform string
	// MediaType is the media type of the SBOM document.
	MediaType string
	// Digest is the digest of the SBOM artifact, as pushed to the registry.
	Digest string
	// Size is the size of the SBOM document in bytes.
	Size int64
	// Ref is where the default publisher pushes the SBOM; it only exists in a
	// registry when the image was pushed.
	Ref name.Reference
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
	opts, err := sbomTagOptions()
	if err != nil {
		return nil, err
	}
	if p.Digest, p.Platforms, p.SBOMs, err = describe(br, result, opts); err != nil {
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

// describe returns the digest, platforms and SBOMs of what was actually
// published as ref. That is br itself, except when the publisher could not
// write an index (daemon, kind) and loaded a single image out of it: ref then
// identifies that child image, so report it rather than the whole index.
func describe(br build.Result, ref name.Reference, opts []ociremote.Option) (string, []string, []SBOM, error) {
	h, err := br.Digest()
	if err != nil {
		return "", nil, nil, err
	}
	idx, isIndex := br.(oci.SignedImageIndex)
	if !isIndex {
		var sboms []SBOM
		var err error
		if se, ok := br.(oci.SignedEntity); ok {
			sboms, err = sbomsOf(se, ref, h, "", opts)
		}
		platforms, perr := platformsOf(br)
		return h.String(), platforms, sboms, errors.Join(err, perr)
	}

	im, err := idx.IndexManifest()
	if err != nil {
		return "", nil, nil, err
	}
	id := strings.TrimPrefix(ref.Identifier(), "sha256:")
	if id != h.Hex {
		for _, d := range im.Manifests {
			if d.Digest.Hex == id && d.Platform != nil {
				child, err := idx.SignedImage(d.Digest)
				if err != nil {
					return "", nil, nil, err
				}
				sboms, err := sbomsOf(child, ref, d.Digest, d.Platform.String(), opts)
				return d.Digest.String(), []string{d.Platform.String()}, sboms, err
			}
		}
	}

	sboms, err := sbomsOf(idx, ref, h, "", opts)
	if err != nil {
		return "", nil, nil, err
	}
	for _, d := range im.Manifests {
		// Skip attestations, signatures and the like.
		if d.Platform == nil || d.Platform.OS == "unknown" {
			continue
		}
		child, err := idx.SignedImage(d.Digest)
		if err != nil {
			return "", nil, nil, err
		}
		cs, err := sbomsOf(child, ref, d.Digest, d.Platform.String(), opts)
		if err != nil {
			return "", nil, nil, err
		}
		sboms = append(sboms, cs...)
	}
	platforms, err := platformsOf(br)
	return h.String(), platforms, sboms, err
}

// sbomTagOptions returns the options the default publisher uses to name SBOMs,
// so that refs point where they were pushed: it respects COSIGN_REPOSITORY.
func sbomTagOptions() ([]ociremote.Option, error) {
	repo, err := ociremote.GetEnvTargetRepository()
	if err != nil {
		return nil, fmt.Errorf("reading COSIGN_REPOSITORY: %w", err)
	}
	if (repo == name.Repository{}) {
		return nil, nil
	}
	return []ociremote.Option{ociremote.WithTargetRepository(repo)}, nil
}

// sbomsOf returns the SBOM attached to se, which has digest h, if any.
func sbomsOf(se oci.SignedEntity, ref name.Reference, h v1.Hash, platform string, opts []ociremote.Option) ([]SBOM, error) {
	f, err := se.Attachment("sbom")
	if err != nil {
		// Not every entity has an SBOM, e.g. with --sbom=none.
		return nil, nil
	}
	mt, err := f.FileMediaType()
	if err != nil {
		return nil, err
	}
	payload, err := f.Payload()
	if err != nil {
		return nil, err
	}
	fh, err := f.Digest()
	if err != nil {
		return nil, err
	}
	sref, err := ociremote.SBOMTag(ref.Context().Digest(h.String()), opts...)
	if err != nil {
		return nil, err
	}
	return []SBOM{{Platform: platform, MediaType: string(mt), Digest: fh.String(), Size: int64(len(payload)), Ref: sref}}, nil
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
