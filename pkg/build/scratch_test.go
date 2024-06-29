// Copyright 2023 ko Build Authors All Rights Reserved.
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

package build

import (
	"testing"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

func TestScratchImage(t *testing.T) {
	img, err := ScratchImage([]string{"linux/s390x", "plan9/386"})
	if err != nil {
		t.Fatalf("expected to create the image, %s", err)
	}
	mt, err := img.MediaType()
	if err != nil {
		t.Fatalf("expected to get a mediatype, %s", err)
	}
	expMT := types.OCIImageIndex
	if mt != expMT {
		t.Errorf("expected media type = %s, got %s", expMT, mt)
	}

	imgIdx, ok := img.(v1.ImageIndex)
	if !ok {
		t.Fatalf("expected to have an image index")
	}

	mf, err := imgIdx.IndexManifest()
	if mt != expMT {
		t.Errorf("expected a manifest, got %s", err)
	}
	if len(mf.Manifests) != 2 {
		t.Fatalf("expected two manifests, got %d", len(mf.Manifests))
	}
	for _, m := range mf.Manifests {
		img, err := imgIdx.Image(m.Digest)
		if err != nil {
			t.Fatalf("expected no error when getting image for digest %s, got %s", m.Digest, err)
		}
		ls, err := img.Layers()
		if len(ls) != 0 {
			t.Errorf("expected no layers, found %d", len(ls))
		}

		switch m.Platform.OS {
		case "linux":
			if m.Platform.Architecture != "s390x" {
				t.Errorf("expected arch = s390x, got %s", m.Platform.Architecture)
			}
		case "plan9":
			if m.Platform.Architecture != "386" {
				t.Errorf("expected arch = 386, got %s", m.Platform.Architecture)
			}
		default:
			t.Errorf("unexpected OS %s", m.Platform.OS)
		}
	}

}
