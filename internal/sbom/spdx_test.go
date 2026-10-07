// Copyright 2026 ko Build Authors All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sbom

import (
	"strings"
	"testing"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/sigstore/cosign/v3/pkg/oci/signed"
)

// A base index can carry a manifest with no platform, which the image index
// spec allows, and ko copies that descriptor through to the index it builds.
func TestGenerateIndexSPDXWithoutPlatform(t *testing.T) {
	img, err := random.Image(4, 1)
	if err != nil {
		t.Fatal(err)
	}

	idx := mutate.IndexMediaType(empty.Index, types.OCIImageIndex)
	idx = mutate.AppendManifests(idx,
		mutate.IndexAddendum{
			Add:        img,
			Descriptor: v1.Descriptor{Platform: &v1.Platform{OS: "linux", Architecture: "amd64"}},
		},
		mutate.IndexAddendum{
			Add: img,
		},
	)

	im, err := idx.IndexManifest()
	if err != nil {
		t.Fatal(err)
	}
	if im.Manifests[1].Platform != nil {
		t.Fatalf("expected the second manifest to have no platform, got %v", im.Manifests[1].Platform)
	}

	b, err := GenerateIndexSPDX("v0.0.0-test", signed.ImageIndex(idx))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(b) == 0 {
		t.Fatal("expected an SBOM, got nothing")
	}
	if !strings.Contains(string(b), "linux") {
		t.Error("expected the platform of the described manifest to survive")
	}
}
