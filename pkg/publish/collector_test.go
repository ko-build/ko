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
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/ko/pkg/build"
	"github.com/sigstore/cosign/v3/pkg/oci"
	ocimutate "github.com/sigstore/cosign/v3/pkg/oci/mutate"
	"github.com/sigstore/cosign/v3/pkg/oci/signed"
	"github.com/sigstore/cosign/v3/pkg/oci/static"
	ctypes "github.com/sigstore/cosign/v3/pkg/types"
)

// platformImage returns a linux/arch image, with sbom attached unless empty.
func platformImage(t *testing.T, arch, sbom string) oci.SignedImage {
	t.Helper()
	img, err := random.Image(10, 1)
	if err != nil {
		t.Fatal(err)
	}
	cf, err := img.ConfigFile()
	if err != nil {
		t.Fatal(err)
	}
	cf.OS, cf.Architecture = "linux", arch
	if img, err = mutate.ConfigFile(img, cf); err != nil {
		t.Fatal(err)
	}
	si := signed.Image(img)
	if sbom == "" {
		return si
	}
	if si, err = ocimutate.AttachFileToImage(si, "sbom", sbomFile(t, sbom)); err != nil {
		t.Fatal(err)
	}
	return si
}

// index returns an index of imgs, built like ko does so that the images keep
// their attachments, with sbom attached unless empty.
func index(t *testing.T, sbom string, imgs ...oci.SignedImage) oci.SignedImageIndex {
	t.Helper()
	adds := make([]ocimutate.IndexAddendum, 0, len(imgs))
	for _, img := range imgs {
		cf, err := img.ConfigFile()
		if err != nil {
			t.Fatal(err)
		}
		adds = append(adds, ocimutate.IndexAddendum{Add: img, Descriptor: v1.Descriptor{Platform: cf.Platform()}})
	}
	idx := ocimutate.AppendManifests(empty.Index, adds...)
	if sbom == "" {
		return idx
	}
	idx, err := ocimutate.AttachFileToImageIndex(idx, "sbom", sbomFile(t, sbom))
	if err != nil {
		t.Fatal(err)
	}
	return idx
}

func sbomFile(t *testing.T, content string) oci.File {
	t.Helper()
	f, err := static.NewFile([]byte(content), static.WithLayerMediaType(ctypes.SPDXJSONMediaType))
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func digest(t *testing.T, e interface{ Digest() (v1.Hash, error) }) v1.Hash {
	t.Helper()
	h, err := e.Digest()
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func parseRef(t *testing.T, s string) name.Reference {
	t.Helper()
	ref, err := name.ParseReference(s)
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

// returning is an inner publisher that publishes everything as ref.
func returning(ref name.Reference) Interface {
	return &cbPublish{cb: func(context.Context, build.Result, string) (name.Reference, error) {
		return ref, nil
	}}
}

// summary is the comparable part of a Published.
type summary struct {
	ImportPath, Ref, Digest string
	Platforms               []string
	IndexSBOM               *sbomSummary
	PlatformSBOMs           []sbomSummary
}

type sbomSummary struct {
	Platform, MediaType, Digest, Subject string
	Size                                 int64
}

func summarizeSBOM(sb SBOM) sbomSummary {
	s := sbomSummary{MediaType: sb.MediaType, Digest: sb.Digest, Subject: sb.Subject.String(), Size: sb.Size}
	if sb.Platform != nil {
		s.Platform = sb.Platform.String()
	}
	return s
}

func summarize(p Published) summary {
	s := summary{ImportPath: p.ImportPath, Ref: p.Ref.String(), Digest: p.Digest, Platforms: p.Platforms}
	if p.SBOMs.Index != nil {
		is := summarizeSBOM(*p.SBOMs.Index)
		s.IndexSBOM = &is
	}
	for _, sb := range p.SBOMs.Platforms {
		s.PlatformSBOMs = append(s.PlatformSBOMs, summarizeSBOM(sb))
	}
	return s
}

// sha256Of returns the digest of content, which an SBOM's Digest must be.
func sha256Of(content string) string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(content)))
}

func TestCollectorResults(t *testing.T) {
	const repo = "example.com/app"
	amd := platformImage(t, "amd64", "amd-sbom")
	arm := platformImage(t, "arm64", "")
	idx := index(t, "index-sbom", amd, arm)
	amdHex, armHex, idxHex := digest(t, amd).Hex, digest(t, arm).Hex, digest(t, idx).Hex
	spdx := string(ctypes.SPDXJSONMediaType)

	for _, tc := range []struct {
		name string
		br   build.Result
		ref  string // as returned by the inner publisher
		want summary
	}{{
		name: "image",
		br:   amd,
		ref:  repo + "@sha256:" + amdHex,
		want: summary{
			ImportPath: repo, Ref: repo + "@sha256:" + amdHex, Digest: "sha256:" + amdHex,
			Platforms:     []string{"linux/amd64"},
			PlatformSBOMs: []sbomSummary{{Platform: "linux/amd64", MediaType: spdx, Digest: sha256Of("amd-sbom"), Subject: repo + "@sha256:" + amdHex, Size: int64(len("amd-sbom"))}},
		},
	}, {
		name: "image without SBOM",
		br:   arm,
		ref:  repo + "@sha256:" + armHex,
		want: summary{ImportPath: repo, Ref: repo + "@sha256:" + armHex, Digest: "sha256:" + armHex, Platforms: []string{"linux/arm64"}},
	}, {
		name: "index: index SBOM, and one per platform image that has one",
		br:   idx,
		ref:  repo + "@sha256:" + idxHex,
		want: summary{
			ImportPath: repo, Ref: repo + "@sha256:" + idxHex, Digest: "sha256:" + idxHex,
			Platforms:     []string{"linux/amd64", "linux/arm64"},
			IndexSBOM:     &sbomSummary{MediaType: spdx, Digest: sha256Of("index-sbom"), Subject: repo + "@sha256:" + idxHex, Size: int64(len("index-sbom"))},
			PlatformSBOMs: []sbomSummary{{Platform: "linux/amd64", MediaType: spdx, Digest: sha256Of("amd-sbom"), Subject: repo + "@sha256:" + amdHex, Size: int64(len("amd-sbom"))}},
		},
	}, {
		name: "index with a tag-only ref",
		br:   idx,
		ref:  repo + ":latest",
		want: summary{
			ImportPath: repo, Ref: repo + ":latest", Digest: "sha256:" + idxHex,
			Platforms:     []string{"linux/amd64", "linux/arm64"},
			IndexSBOM:     &sbomSummary{MediaType: spdx, Digest: sha256Of("index-sbom"), Subject: repo + "@sha256:" + idxHex, Size: int64(len("index-sbom"))},
			PlatformSBOMs: []sbomSummary{{Platform: "linux/amd64", MediaType: spdx, Digest: sha256Of("amd-sbom"), Subject: repo + "@sha256:" + amdHex, Size: int64(len("amd-sbom"))}},
		},
	}, {
		name: "index loaded as a single image (daemon, kind)",
		br:   idx,
		ref:  "ko.local/app:" + amdHex,
		want: summary{
			ImportPath: repo, Ref: "ko.local/app:" + amdHex, Digest: "sha256:" + amdHex,
			Platforms:     []string{"linux/amd64"},
			PlatformSBOMs: []sbomSummary{{Platform: "linux/amd64", MediaType: spdx, Digest: sha256Of("amd-sbom"), Subject: "ko.local/app@sha256:" + amdHex, Size: int64(len("amd-sbom"))}},
		},
	}} {
		t.Run(tc.name, func(t *testing.T) {
			col := NewCollector(returning(parseRef(t, tc.ref)))
			if _, err := col.Publish(context.Background(), tc.br, build.StrictScheme+repo); err != nil {
				t.Fatalf("Publish() = %v", err)
			}
			got, err := col.Results()
			if err != nil {
				t.Fatalf("Results() = %v", err)
			}
			if len(got) != 1 {
				t.Fatalf("Results() has %d entries, want 1", len(got))
			}
			if diff := cmp.Diff(tc.want, summarize(got[0])); diff != "" {
				t.Errorf("Results() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCollectorResultsSorted(t *testing.T) {
	img := platformImage(t, "amd64", "")
	col := NewCollector(returning(parseRef(t, "example.com/app@sha256:"+digest(t, img).Hex)))
	for _, ip := range []string{"ko://example.com/b", "example.com/a"} {
		if _, err := col.Publish(context.Background(), img, ip); err != nil {
			t.Fatalf("Publish(%s) = %v", ip, err)
		}
	}
	got, err := col.Results()
	if err != nil {
		t.Fatalf("Results() = %v", err)
	}
	paths := make([]string, 0, len(got))
	for _, p := range got {
		paths = append(paths, p.ImportPath)
	}
	if diff := cmp.Diff([]string{"example.com/a", "example.com/b"}, paths); diff != "" {
		t.Errorf("import paths mismatch (-want +got):\n%s", diff)
	}
}

func TestCollectorErrors(t *testing.T) {
	img := platformImage(t, "amd64", "sbom")
	ref := parseRef(t, "example.com/app@sha256:"+digest(t, img).Hex)

	t.Run("inner publish error is returned and nothing is recorded", func(t *testing.T) {
		want := errors.New("boom")
		col := NewCollector(&cbPublish{cb: func(context.Context, build.Result, string) (name.Reference, error) {
			return nil, want
		}})
		if _, err := col.Publish(context.Background(), img, "example.com/app"); !errors.Is(err, want) {
			t.Errorf("Publish() = %v, want %v", err, want)
		}
		got, err := col.Results()
		if err != nil || len(got) != 0 {
			t.Errorf("Results() = %v, %v; want no entries", got, err)
		}
	})

	t.Run("an invalid COSIGN_REPOSITORY does not matter to the collector", func(t *testing.T) {
		t.Setenv("COSIGN_REPOSITORY", "not a valid repository!")
		col := NewCollector(returning(ref))
		if _, err := col.Publish(context.Background(), img, "example.com/app"); err != nil {
			t.Fatalf("Publish() = %v", err)
		}
		if _, err := col.Results(); err != nil {
			t.Errorf("Results() = %v, want success", err)
		}
	})
}

func TestSBOMRef(t *testing.T) {
	hex := strings.Repeat("a", 64)
	subject, err := name.NewDigest("example.com/app@sha256:" + hex)
	if err != nil {
		t.Fatal(err)
	}
	sb := SBOM{Subject: subject}
	for _, tc := range []struct {
		name, cosignRepository, want string
		wantErr                      bool
	}{
		{name: "next to the subject", want: "example.com/app:sha256-" + hex + ".sbom"},
		{name: "COSIGN_REPOSITORY", cosignRepository: "sboms.example.com/store", want: "sboms.example.com/store:sha256-" + hex + ".sbom"},
		{name: "invalid COSIGN_REPOSITORY", cosignRepository: "not a valid repository!", wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("COSIGN_REPOSITORY", tc.cosignRepository)
			got, err := SBOMRef(sb)
			if (err != nil) != tc.wantErr {
				t.Fatalf("SBOMRef() = %v, wantErr %v", err, tc.wantErr)
			}
			if !tc.wantErr && got.String() != tc.want {
				t.Errorf("SBOMRef() = %s, want %s", got, tc.want)
			}
		})
	}
}
