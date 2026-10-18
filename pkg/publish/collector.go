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
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
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
	// ImportPath is the import path that was built, without the ko:// scheme.
	ImportPath string
	// Ref is the reference the image was published as.
	Ref name.Reference
	// Digest is the digest of the image or image index.
	Digest string
	// Platforms lists the os/arch[/variant] of each image, sorted.
	Platforms []string
	// SBOMs are the SBOMs attached to what was published.
	SBOMs SBOMs
}

// SBOMs are the SBOMs attached to a published image or image index.
type SBOMs struct {
	// Index is the SBOM of the image index, which relates its platform
	// images. It is nil when no index was published.
	Index *SBOM
	// Platforms are the SBOMs of the platform images, which list what each
	// binary was built from.
	Platforms []SBOM
}

// SBOM describes an SBOM attached to a published image or image index.
type SBOM struct {
	// Platform is the platform of the image the SBOM belongs to, nil for an
	// index.
	Platform *v1.Platform
	// MediaType is the media type of the SBOM document.
	MediaType string
	// Digest is the digest of the SBOM document: the sha256 of its content.
	Digest string
	// Size is the size of the SBOM document in bytes.
	Size int64
	// Subject is the image or image index the SBOM is attached to. When it is
	// pushed to a registry, the SBOM is pushed next to it (see SBOMRef).
	Subject name.Digest
}

// Collector is an Interface that remembers what its inner publisher
// published, so that it can be described afterwards with Results.
type Collector struct {
	inner Interface

	mu        sync.Mutex
	published []published
}

// published is what Publish records; it is only inspected by Results.
type published struct {
	importPath string
	ref        name.Reference
	br         build.Result
}

var _ Interface = (*Collector)(nil)

// NewCollector wraps inner in a publisher that records every successful
// Publish call.
func NewCollector(inner Interface) *Collector {
	return &Collector{inner: inner}
}

// Publish implements Interface. It never fails once the inner publisher
// succeeded: describing the result is left to Results.
func (c *Collector) Publish(ctx context.Context, br build.Result, ref string) (name.Reference, error) {
	result, err := c.inner.Publish(ctx, br, ref)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.published = append(c.published, published{
		importPath: strings.TrimPrefix(ref, build.StrictScheme),
		ref:        result,
		br:         br,
	})
	return result, nil
}

// Results describes everything published so far, sorted by import path.
func (c *Collector) Results() ([]Published, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	out := make([]Published, 0, len(c.published))
	for _, p := range c.published {
		d, err := describe(p)
		if err != nil {
			return nil, fmt.Errorf("describing %s: %w", p.importPath, err)
		}
		out = append(out, d)
	}
	slices.SortFunc(out, func(a, b Published) int { return cmp.Compare(a.ImportPath, b.ImportPath) })
	return out, nil
}

// Close implements Interface
func (c *Collector) Close() error {
	return c.inner.Close()
}

// entity is an image or image index that was published.
type entity struct {
	se       oci.SignedEntity // nil when the result carries no attachments
	digest   v1.Hash
	platform *v1.Platform // nil for an index
}

// describe reports what was actually published as p.ref.
func describe(p published) (Published, error) {
	es, err := entities(p.br, p.ref)
	if err != nil {
		return Published{}, err
	}
	d := Published{
		ImportPath: p.importPath,
		Ref:        p.ref,
		Digest:     es[0].digest.String(),
		Platforms:  []string{},
		SBOMs:      SBOMs{Platforms: []SBOM{}},
	}
	for _, e := range es {
		if e.platform != nil {
			d.Platforms = append(d.Platforms, e.platform.String())
		}
		sb, ok, err := sbomOf(e, p.ref.Context())
		if err != nil {
			return Published{}, fmt.Errorf("describing SBOM: %w", err)
		}
		switch {
		case !ok:
		case e.platform == nil:
			d.SBOMs.Index = &sb
		default:
			d.SBOMs.Platforms = append(d.SBOMs.Platforms, sb)
		}
	}
	slices.Sort(d.Platforms)
	return d, nil
}

// entities lists what was published as ref, top-level entity first: the
// image; or the index followed by its platform images; or, when the publisher
// could not write an index (daemon, kind) and loaded a single image out of it,
// that image, which ref then identifies.
func entities(br build.Result, ref name.Reference) ([]entity, error) {
	h, err := br.Digest()
	if err != nil {
		return nil, err
	}
	se, _ := br.(oci.SignedEntity)

	idx, ok := br.(oci.SignedImageIndex)
	if !ok {
		img, ok := br.(v1.Image)
		if !ok {
			return nil, fmt.Errorf("unexpected build result %T", br)
		}
		cf, err := img.ConfigFile()
		if err != nil {
			return nil, err
		}
		return []entity{{se: se, digest: h, platform: cf.Platform()}}, nil
	}

	im, err := idx.IndexManifest()
	if err != nil {
		return nil, err
	}
	var children []entity
	for _, d := range im.Manifests {
		// Skip attestations, signatures and the like.
		if d.Platform == nil || d.Platform.OS == "unknown" {
			continue
		}
		child, err := idx.SignedImage(d.Digest)
		if err != nil {
			return nil, err
		}
		children = append(children, entity{se: child, digest: d.Digest, platform: d.Platform})
	}

	loaded := strings.TrimPrefix(ref.Identifier(), "sha256:")
	if loaded != h.Hex {
		for _, c := range children {
			if c.digest.Hex == loaded {
				return []entity{c}, nil
			}
		}
	}
	return append([]entity{{se: se, digest: h}}, children...), nil
}

// SBOMRef returns where the default publisher pushes sb, as it does: next to
// its subject, or in COSIGN_REPOSITORY when set. Only call it for SBOMs that
// were pushed: it fails on an invalid COSIGN_REPOSITORY, which the default
// publisher rejects up front.
func SBOMRef(sb SBOM) (name.Tag, error) {
	repo, err := ociremote.GetEnvTargetRepository()
	if err != nil {
		return name.Tag{}, fmt.Errorf("reading COSIGN_REPOSITORY: %w", err)
	}
	var opts []ociremote.Option
	if (repo != name.Repository{}) {
		opts = append(opts, ociremote.WithTargetRepository(repo))
	}
	return ociremote.SBOMTag(sb.Subject, opts...)
}

// sbomOf describes the SBOM attached to e, if it has one.
func sbomOf(e entity, repo name.Repository) (SBOM, bool, error) {
	if e.se == nil {
		return SBOM{}, false, nil
	}
	f, err := e.se.Attachment("sbom")
	if err != nil {
		// Not every entity has an SBOM, e.g. with --sbom=none.
		return SBOM{}, false, nil
	}
	mt, err := f.FileMediaType()
	if err != nil {
		return SBOM{}, false, err
	}
	layers, err := f.Layers()
	if err != nil {
		return SBOM{}, false, err
	}
	if len(layers) != 1 {
		return SBOM{}, false, errors.New("SBOM attachment does not have exactly one layer")
	}
	size, err := layers[0].Size()
	if err != nil {
		return SBOM{}, false, err
	}
	// The layer holds the SBOM document as is, so its digest is the
	// document's: the same as the file in the SBOM directory, and the blob to
	// fetch from the registry.
	dh, err := layers[0].Digest()
	if err != nil {
		return SBOM{}, false, err
	}
	return SBOM{Platform: e.platform, MediaType: string(mt), Digest: dh.String(), Size: size, Subject: repo.Digest(e.digest.String())}, true, nil
}
