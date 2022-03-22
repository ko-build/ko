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
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/ko/pkg/build"
	"github.com/google/ko/pkg/publish/kind"
)

const (
	// KindDomain is a sentinel "registry" that represents side-loading images into kind nodes.
	KindDomain = "kind.local"
)

type kindPublisher struct {
	namer Namer
	tags  []string
}

// NewKindPublisher returns a new publish.Interface that loads images into kind nodes.
func NewKindPublisher(namer Namer, tags []string) Interface {
	return &kindPublisher{
		namer: namer,
		tags:  tags,
	}
}

// Publish implements publish.Interface.
func (t *kindPublisher) Publish(ctx context.Context, br build.Result, s string) (name.Reference, error) {
	s = strings.TrimPrefix(s, build.StrictScheme)
	// https://github.com/google/go-containerregistry/issues/212
	s = strings.ToLower(s)

	img, err := ToImage(br, s)
	if err != nil {
		return nil, err
	}

	h, err := img.Digest()
	if err != nil {
		return nil, err
	}

	digestTag, err := name.NewTag(fmt.Sprintf("%s:%s", t.namer(KindDomain, s), h.Hex))
	if err != nil {
		return nil, err
	}

	log.Printf("Loading %v", digestTag)
	if err := kind.Write(ctx, digestTag, img); err != nil {
		return nil, err
	}
	log.Printf("Loaded %v", digestTag)

	for _, tagName := range t.tags {
		log.Printf("Adding tag %v", tagName)
		tag, err := name.NewTag(fmt.Sprintf("%s:%s", t.namer(KindDomain, s), tagName))
		if err != nil {
			return nil, err
		}

		if err := kind.Tag(ctx, digestTag, tag); err != nil {
			return nil, err
		}
		log.Printf("Added tag %v", tagName)
	}

	return &digestTag, nil
}

func (t *kindPublisher) Close() error {
	return nil
}
