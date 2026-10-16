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
	"reflect"
	"testing"

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
)

func TestCollector(t *testing.T) {
	repo := name.MustParseReference("example.com/app:latest")
	inner := &cbPublish{cb: func(_ context.Context, b build.Result, _ string) (name.Reference, error) {
		h, err := b.Digest()
		if err != nil {
			return nil, err
		}
		return repo.Context().Digest(h.String()), nil
	}}

	img, err := random.Image(10, 1)
	if err != nil {
		t.Fatal(err)
	}
	cf, err := img.ConfigFile()
	if err != nil {
		t.Fatal(err)
	}
	cf.OS, cf.Architecture = "linux", "arm64"
	if img, err = mutate.ConfigFile(img, cf); err != nil {
		t.Fatal(err)
	}

	child := func(arch string) v1.Image {
		i, err := random.Image(10, 1)
		if err != nil {
			t.Fatal(err)
		}
		c, _ := i.ConfigFile()
		c.OS, c.Architecture = "linux", arch
		i, err = mutate.ConfigFile(i, c)
		if err != nil {
			t.Fatal(err)
		}
		return i
	}
	idx := mutate.AppendManifests(empty.Index,
		mutate.IndexAddendum{Add: child("arm64"), Descriptor: v1.Descriptor{Platform: &v1.Platform{OS: "linux", Architecture: "arm64"}}},
		mutate.IndexAddendum{Add: child("amd64"), Descriptor: v1.Descriptor{Platform: &v1.Platform{OS: "linux", Architecture: "amd64"}}},
	)

	col := NewCollector(inner)
	for _, tc := range []struct {
		importpath string
		res        build.Result
		want       []string
	}{
		{"example.com/b", signed.Image(img), []string{"linux/arm64"}},
		{"example.com/a", signed.ImageIndex(idx), []string{"linux/amd64", "linux/arm64"}},
	} {
		if _, err := col.Publish(context.Background(), tc.res, tc.importpath); err != nil {
			t.Fatalf("Publish(%s) = %v", tc.importpath, err)
		}
	}

	got := col.Results()
	if len(got) != 2 || got[0].ImportPath != "example.com/a" || got[1].ImportPath != "example.com/b" {
		t.Fatalf("Results() not sorted by import path: %+v", got)
	}
	if want := []string{"linux/amd64", "linux/arm64"}; !reflect.DeepEqual(got[0].Platforms, want) {
		t.Errorf("index platforms = %v, want %v", got[0].Platforms, want)
	}
	if want := []string{"linux/arm64"}; !reflect.DeepEqual(got[1].Platforms, want) {
		t.Errorf("image platforms = %v, want %v", got[1].Platforms, want)
	}
	for _, p := range got {
		if p.Ref.Identifier() != p.Digest {
			t.Errorf("%s: ref %s does not match digest %s", p.ImportPath, p.Ref, p.Digest)
		}
	}
}

func TestCollectorPublishError(t *testing.T) {
	want := errors.New("boom")
	col := NewCollector(&cbPublish{cb: func(context.Context, build.Result, string) (name.Reference, error) {
		return nil, want
	}})
	if _, err := col.Publish(context.Background(), signed.Image(empty.Image), "x"); !errors.Is(err, want) {
		t.Errorf("Publish() = %v, want %v", err, want)
	}
	if n := len(col.Results()); n != 0 {
		t.Errorf("Results() has %d entries after failure", n)
	}
}

func TestCollectorDowncastedIndex(t *testing.T) {
	repo := name.MustParseReference("ko.local/app:latest")
	var childDigest v1.Hash
	child := func(arch string) v1.Image {
		i, err := random.Image(10, 1)
		if err != nil {
			t.Fatal(err)
		}
		c, _ := i.ConfigFile()
		c.OS, c.Architecture = "linux", arch
		if i, err = mutate.ConfigFile(i, c); err != nil {
			t.Fatal(err)
		}
		if arch == "arm64" {
			childDigest, _ = i.Digest()
		}
		return i
	}
	idx := mutate.AppendManifests(empty.Index,
		mutate.IndexAddendum{Add: child("amd64"), Descriptor: v1.Descriptor{Platform: &v1.Platform{OS: "linux", Architecture: "amd64"}}},
		mutate.IndexAddendum{Add: child("arm64"), Descriptor: v1.Descriptor{Platform: &v1.Platform{OS: "linux", Architecture: "arm64"}}},
	)

	for _, tc := range []struct {
		name          string
		ref           func(index v1.Hash) name.Reference
		wantDigest    func(index v1.Hash) string
		wantPlatforms []string
	}{{
		name:          "daemon-style tag of a child image",
		ref:           func(v1.Hash) name.Reference { return repo.Context().Tag(childDigest.Hex) },
		wantDigest:    func(v1.Hash) string { return childDigest.String() },
		wantPlatforms: []string{"linux/arm64"},
	}, {
		name:          "index digest",
		ref:           func(h v1.Hash) name.Reference { return repo.Context().Digest(h.String()) },
		wantDigest:    func(h v1.Hash) string { return h.String() },
		wantPlatforms: []string{"linux/amd64", "linux/arm64"},
	}, {
		name:          "tag-only ref",
		ref:           func(v1.Hash) name.Reference { return repo },
		wantDigest:    func(h v1.Hash) string { return h.String() },
		wantPlatforms: []string{"linux/amd64", "linux/arm64"},
	}} {
		t.Run(tc.name, func(t *testing.T) {
			br := signed.ImageIndex(idx)
			ih, err := br.Digest()
			if err != nil {
				t.Fatal(err)
			}
			col := NewCollector(&cbPublish{cb: func(context.Context, build.Result, string) (name.Reference, error) {
				return tc.ref(ih), nil
			}})
			if _, err := col.Publish(context.Background(), br, "example.com/app"); err != nil {
				t.Fatalf("Publish() = %v", err)
			}
			got := col.Results()[0]
			if want := tc.wantDigest(ih); got.Digest != want {
				t.Errorf("Digest = %s, want %s", got.Digest, want)
			}
			if !reflect.DeepEqual(got.Platforms, tc.wantPlatforms) {
				t.Errorf("Platforms = %v, want %v", got.Platforms, tc.wantPlatforms)
			}
		})
	}
}

func TestCollectorSBOMs(t *testing.T) {
	repo := name.MustParseReference("example.com/app")
	withSBOM := func(arch string, content string) oci.SignedImage {
		i, err := random.Image(10, 1)
		if err != nil {
			t.Fatal(err)
		}
		c, _ := i.ConfigFile()
		c.OS, c.Architecture = "linux", arch
		if i, err = mutate.ConfigFile(i, c); err != nil {
			t.Fatal(err)
		}
		si := signed.Image(i)
		if content == "" {
			return si
		}
		f, err := static.NewFile([]byte(content), static.WithLayerMediaType("text/spdx+json"))
		if err != nil {
			t.Fatal(err)
		}
		if si, err = ocimutate.AttachFileToImage(si, "sbom", f); err != nil {
			t.Fatal(err)
		}
		return si
	}
	amd, arm := withSBOM("amd64", "amd-sbom"), withSBOM("arm64", "")
	// Like ko, use the signed AppendManifests so children keep their attachments.
	idx := ocimutate.AppendManifests(empty.Index,
		ocimutate.IndexAddendum{Add: amd, Descriptor: v1.Descriptor{Platform: &v1.Platform{OS: "linux", Architecture: "amd64"}}},
		ocimutate.IndexAddendum{Add: arm, Descriptor: v1.Descriptor{Platform: &v1.Platform{OS: "linux", Architecture: "arm64"}}},
	)
	idxFile, err := static.NewFile([]byte("index-sbom"), static.WithLayerMediaType("text/spdx+json"))
	if err != nil {
		t.Fatal(err)
	}
	idx, err = ocimutate.AttachFileToImageIndex(idx, "sbom", idxFile)
	if err != nil {
		t.Fatal(err)
	}
	amdDigest, _ := amd.Digest()
	idxDigest, _ := idx.Digest()

	type want struct {
		platform string
		size     int64
		ref      string
	}
	for _, tc := range []struct {
		name string
		br   build.Result
		ref  func() name.Reference
		want []want
	}{{
		name: "index: top-level then per-platform, platforms without SBOM skipped",
		br:   idx,
		ref:  func() name.Reference { return repo.Context().Digest(idxDigest.String()) },
		want: []want{
			{"", int64(len("index-sbom")), "example.com/app:sha256-" + idxDigest.Hex + ".sbom"},
			{"linux/amd64", int64(len("amd-sbom")), "example.com/app:sha256-" + amdDigest.Hex + ".sbom"},
		},
	}, {
		name: "downcast to a child image reports only that image",
		br:   idx,
		ref:  func() name.Reference { return repo.Context().Tag(amdDigest.Hex) },
		want: []want{{"linux/amd64", int64(len("amd-sbom")), "example.com/app:sha256-" + amdDigest.Hex + ".sbom"}},
	}, {
		name: "single image without SBOM",
		br:   arm,
		ref:  func() name.Reference { return repo },
		want: nil,
	}} {
		t.Run(tc.name, func(t *testing.T) {
			col := NewCollector(&cbPublish{cb: func(context.Context, build.Result, string) (name.Reference, error) {
				return tc.ref(), nil
			}})
			if _, err := col.Publish(context.Background(), tc.br, "example.com/app"); err != nil {
				t.Fatalf("Publish() = %v", err)
			}
			var got []want
			sboms := col.Results()[0].SBOMs
			if len(sboms) > 0 {
				got = make([]want, 0, len(sboms))
			}
			for _, sb := range sboms {
				if sb.MediaType != "text/spdx+json" || sb.Digest == "" {
					t.Errorf("incomplete SBOM entry: %+v", sb)
				}
				got = append(got, want{sb.Platform, sb.Size, sb.Ref.String()})
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("SBOMs = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestCollectorSBOMRefRespectsCosignRepository(t *testing.T) {
	t.Setenv("COSIGN_REPOSITORY", "sboms.example.com/store")

	i, err := random.Image(10, 1)
	if err != nil {
		t.Fatal(err)
	}
	f, err := static.NewFile([]byte("sbom"), static.WithLayerMediaType("text/spdx+json"))
	if err != nil {
		t.Fatal(err)
	}
	si, err := ocimutate.AttachFileToImage(signed.Image(i), "sbom", f)
	if err != nil {
		t.Fatal(err)
	}
	h, err := si.Digest()
	if err != nil {
		t.Fatal(err)
	}

	repo := name.MustParseReference("example.com/app")
	col := NewCollector(&cbPublish{cb: func(context.Context, build.Result, string) (name.Reference, error) {
		return repo.Context().Digest(h.String()), nil
	}})
	if _, err := col.Publish(context.Background(), si, "example.com/app"); err != nil {
		t.Fatalf("Publish() = %v", err)
	}
	got := col.Results()[0].SBOMs
	if len(got) != 1 {
		t.Fatalf("got %d SBOMs, want 1", len(got))
	}
	if want := "sboms.example.com/store:sha256-" + h.Hex + ".sbom"; got[0].Ref.String() != want {
		t.Errorf("Ref = %s, want %s", got[0].Ref, want)
	}
}
