// Copyright 2018 Google LLC All Rights Reserved.
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
	"fmt"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/internal/legacy"
	"github.com/google/go-containerregistry/pkg/logs"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

// Copy copies a remote image or index from src to dst.
func Copy(src, dst string, opt ...Option) error {
	o := makeOptions(opt...)
	srcRef, err := name.ParseReference(src, o.name...)
	if err != nil {
		return fmt.Errorf("parsing reference %q: %v", src, err)
	}

	dstRef, err := name.ParseReference(dst, o.name...)
	if err != nil {
		return fmt.Errorf("parsing reference for %q: %v", dst, err)
	}

	logs.Progress.Printf("Copying from %v to %v", srcRef, dstRef)
	desc, err := remote.Get(srcRef, o.remote...)
	if err != nil {
		return fmt.Errorf("fetching %q: %v", src, err)
	}

	switch desc.MediaType {
	case types.OCIImageIndex, types.DockerManifestList:
		// Handle indexes separately.
		if o.platform != nil {
			// If platform is explicitly set, don't copy the whole index, just the appropriate image.
			if err := copyImage(desc, dstRef, o); err != nil {
				return fmt.Errorf("failed to copy image: %v", err)
			}
		} else {
			if err := copyIndex(desc, dstRef, o); err != nil {
				return fmt.Errorf("failed to copy index: %v", err)
			}
		}
	case types.DockerManifestSchema1, types.DockerManifestSchema1Signed:
		// Handle schema 1 images separately.
		if err := copySchema1(desc, srcRef, dstRef); err != nil {
			return fmt.Errorf("failed to copy schema 1 image: %v", err)
		}
	default:
		// Assume anything else is an image, since some registries don't set mediaTypes properly.
		if err := copyImage(desc, dstRef, o); err != nil {
			return fmt.Errorf("failed to copy image: %v", err)
		}
	}

	return nil
}

func copyImage(desc *remote.Descriptor, dstRef name.Reference, o options) error {
	img, err := desc.Image()
	if err != nil {
		return err
	}
	return remote.Write(dstRef, img, o.remote...)
}

func copyIndex(desc *remote.Descriptor, dstRef name.Reference, o options) error {
	idx, err := desc.ImageIndex()
	if err != nil {
		return err
	}
	return remote.WriteIndex(dstRef, idx, o.remote...)
}

func copySchema1(desc *remote.Descriptor, srcRef, dstRef name.Reference) error {
	srcAuth, err := authn.DefaultKeychain.Resolve(srcRef.Context().Registry)
	if err != nil {
		return err
	}
	dstAuth, err := authn.DefaultKeychain.Resolve(dstRef.Context().Registry)
	if err != nil {
		return err
	}

	return legacy.CopySchema1(desc, srcRef, dstRef, srcAuth, dstAuth)
}
