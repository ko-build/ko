// Copyright 2020 Google LLC All Rights Reserved.
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

package crane

import (
	"errors"
	"fmt"

	"github.com/containerd/stargz-snapshotter/estargz"
	"github.com/google/go-containerregistry/pkg/logs"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

// Optimize optimizes a remote image or index from src to dst.
// THIS API IS EXPERIMENTAL AND SUBJECT TO CHANGE WITHOUT WARNING.
func Optimize(src, dst string, prioritize []string, opt ...Option) error {
	o := makeOptions(opt...)
	srcRef, err := name.ParseReference(src, o.name...)
	if err != nil {
		return fmt.Errorf("parsing reference %q: %v", src, err)
	}

	dstRef, err := name.ParseReference(dst, o.name...)
	if err != nil {
		return fmt.Errorf("parsing reference for %q: %v", dst, err)
	}

	logs.Progress.Printf("Optimizing from %v to %v", srcRef, dstRef)
	desc, err := remote.Get(srcRef, o.remote...)
	if err != nil {
		return fmt.Errorf("fetching %q: %v", src, err)
	}

	switch desc.MediaType {
	case types.OCIImageIndex, types.DockerManifestList:
		// Handle indexes separately.
		if o.platform != nil {
			// If platform is explicitly set, don't optimize the whole index, just the appropriate image.
			if err := optimizeAndPushImage(desc, dstRef, prioritize, o); err != nil {
				return fmt.Errorf("failed to optimize image: %v", err)
			}
		} else {
			if err := optimizeAndPushIndex(desc, dstRef, prioritize, o); err != nil {
				return fmt.Errorf("failed to optimize index: %v", err)
			}
		}

	case types.DockerManifestSchema1, types.DockerManifestSchema1Signed:
		return errors.New("docker schema 1 images are not supported")

	default:
		// Assume anything else is an image, since some registries don't set mediaTypes properly.
		if err := optimizeAndPushImage(desc, dstRef, prioritize, o); err != nil {
			return fmt.Errorf("failed to optimize image: %v", err)
		}
	}

	return nil
}

func optimizeAndPushImage(desc *remote.Descriptor, dstRef name.Reference, prioritize []string, o options) error {
	img, err := desc.Image()
	if err != nil {
		return err
	}

	oimg, err := optimizeImage(img, prioritize)
	if err != nil {
		return err
	}

	return remote.Write(dstRef, oimg, o.remote...)
}

func optimizeImage(img v1.Image, prioritize []string) (v1.Image, error) {
	cfg, err := img.ConfigFile()
	if err != nil {
		return nil, err
	}
	ocfg := cfg.DeepCopy()
	ocfg.History = nil
	ocfg.RootFS.DiffIDs = nil

	oimg, err := mutate.ConfigFile(empty.Image, ocfg)
	if err != nil {
		return nil, err
	}

	layers, err := img.Layers()
	if err != nil {
		return nil, err
	}

	olayers := make([]mutate.Addendum, 0, len(layers))
	for _, layer := range layers {
		olayer, err := tarball.LayerFromOpener(layer.Uncompressed,
			tarball.WithEstargz,
			tarball.WithEstargzOptions(estargz.WithPrioritizedFiles(prioritize)))
		if err != nil {
			return nil, err
		}

		olayers = append(olayers, mutate.Addendum{
			Layer:     olayer,
			MediaType: types.DockerLayer,
		})
	}

	return mutate.Append(oimg, olayers...)
}

func optimizeAndPushIndex(desc *remote.Descriptor, dstRef name.Reference, prioritize []string, o options) error {
	idx, err := desc.ImageIndex()
	if err != nil {
		return err
	}

	oidx, err := optimizeIndex(idx, prioritize)
	if err != nil {
		return err
	}

	return remote.WriteIndex(dstRef, oidx, o.remote...)
}

func optimizeIndex(idx v1.ImageIndex, prioritize []string) (v1.ImageIndex, error) {
	im, err := idx.IndexManifest()
	if err != nil {
		return nil, err
	}

	// Build an image for each child from the base and append it to a new index to produce the result.
	adds := make([]mutate.IndexAddendum, 0, len(im.Manifests))
	for _, desc := range im.Manifests {
		img, err := idx.Image(desc.Digest)
		if err != nil {
			return nil, err
		}

		oimg, err := optimizeImage(img, prioritize)
		if err != nil {
			return nil, err
		}
		adds = append(adds, mutate.IndexAddendum{
			Add: oimg,
			Descriptor: v1.Descriptor{
				URLs:        desc.URLs,
				MediaType:   desc.MediaType,
				Annotations: desc.Annotations,
				Platform:    desc.Platform,
			},
		})
	}

	idxType, err := idx.MediaType()
	if err != nil {
		return nil, err
	}

	return mutate.IndexMediaType(mutate.AppendManifests(empty.Index, adds...), idxType), nil
}
