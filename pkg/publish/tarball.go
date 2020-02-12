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

package publish

import (
	"fmt"
	"log"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

type TarballPublisher struct {
	file  string
	base  string
	namer Namer
	tags  []string
}

// NewTarball returns a new publish.Interface that saves images to a tarball.
func NewTarball(file, base string, namer Namer, tags []string) *TarballPublisher {
	return &TarballPublisher{file, base, namer, tags}
}

// Publish implements publish.Interface.
func (t *TarballPublisher) Publish(img v1.Image, s string) (name.Reference, error) {
	// https://github.com/google/go-containerregistry/issues/212
	s = strings.ToLower(s)

	refs := make(map[name.Reference]v1.Image)
	for _, tagName := range t.tags {
		tag, err := name.ParseReference(fmt.Sprintf("%s/%s:%s", t.base, t.namer(s), tagName))
		if err != nil {
			return nil, err
		}
		refs[tag] = img
	}

	h, err := img.Digest()
	if err != nil {
		return nil, err
	}

	if len(t.tags) == 0 {
		ref, err := name.ParseReference(fmt.Sprintf("%s/%s@%s", t.base, t.namer(s), h))
		if err != nil {
			return nil, err
		}
		refs[ref] = img
	}

	ref := fmt.Sprintf("%s/%s@%s", t.base, t.namer(s), h)
	if len(t.tags) == 1 && t.tags[0] != defaultTags[0] {
		// If a single tag is explicitly set (not latest), then this
		// is probably a release, so include the tag in the reference.
		ref = fmt.Sprintf("%s/%s:%s@%s", t.base, t.namer(s), t.tags[0], h)
	}
	dig, err := name.NewDigest(ref)
	if err != nil {
		return nil, err
	}

	log.Printf("Saving %v", dig)
	if err := tarball.MultiRefWriteToFile(t.file, refs); err != nil {
		return nil, err
	}
	log.Printf("Saved %v", dig)

	return &dig, nil
}
