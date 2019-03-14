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

package publish

import (
	"fmt"
	"log"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
)

const (
	// LocalDomain is a sentinel "registry" that represents side-loading images into the daemon.
	LocalDomain = "ko.local"
)

// demon is intentionally misspelled to avoid name collision (and drive Jon nuts).
type demon struct {
	namer Namer
	tags  []string
}

// NewDaemon returns a new publish.Interface that publishes images to a container daemon.
func NewDaemon(namer Namer, tags []string) Interface {
	return &demon{namer, tags}
}

// Publish implements publish.Interface
func (d *demon) Publish(img v1.Image, s string) (name.Reference, error) {
	// https://github.com/google/go-containerregistry/issues/212
	s = strings.ToLower(s)

	h, err := img.Digest()
	if err != nil {
		return nil, err
	}

	digestTag, err := name.NewTag(fmt.Sprintf("%s/%s:%s", LocalDomain, d.namer(s), h.Hex), name.WeakValidation)
	if err != nil {
		return nil, err
	}

	log.Printf("Loading %v", digestTag)
	if _, err := daemon.Write(digestTag, img); err != nil {
		return nil, err
	}
	log.Printf("Loaded %v", digestTag)

	for _, tagName := range d.tags {
		log.Printf("Adding tag %v", tagName)
		tag, err := name.NewTag(fmt.Sprintf("%s/%s:%s", LocalDomain, d.namer(s), tagName), name.WeakValidation)
		if err != nil {
			return nil, err
		}

		err = daemon.Tag(digestTag, tag)

		if err != nil {
			return nil, err
		}
		log.Printf("Added tag %v", tagName)
	}

	return &digestTag, nil
}
