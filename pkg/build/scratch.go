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
	"fmt"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

// ScratchImage returns a scratch image manifest with scratch images for each of the specified platforms
func ScratchImage(platforms []string) (Result, error) {
	var manifests []mutate.IndexAddendum
	for _, pf := range platforms {
		if pf == "all" {
			return nil, fmt.Errorf("'all' is not supported for building a scratch image, the platform list must be provided")
		}
		p, err := v1.ParsePlatform(pf)
		if err != nil {
			return nil, err
		}
		img, err := mutate.ConfigFile(empty.Image, &v1.ConfigFile{
			RootFS: v1.RootFS{
				// Some clients check this.
				Type: "layers",
			},
			Architecture: p.Architecture,
			OS:           p.OS,
			Variant:      p.Variant,
			OSVersion:    p.OSVersion,
			OSFeatures:   p.OSFeatures,
		},
		)
		if err != nil {
			return nil, fmt.Errorf("setting config file on empty image, %w", err)
		}
		manifests = append(manifests, mutate.IndexAddendum{
			Add: img,
			Descriptor: v1.Descriptor{
				Platform: p,
			},
		})
	}
	idx := mutate.IndexMediaType(empty.Index, types.OCIImageIndex)
	idx = mutate.AppendManifests(idx, manifests...)
	return idx, nil
}
