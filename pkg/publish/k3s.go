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
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/ko/pkg/build"
	"github.com/google/ko/pkg/publish/k3s"
	"log"
	"strings"
)

const (
	//K3sDomain   k3s local sentinel registry where the images get's loaded
	K3sDomain = "k3s.local"
)

type k3sPublisher struct {
	namer Namer
	tags  []string
}

//NewK3sPublisher returns a new publish.Interface that loads image into k3s clusters
func NewK3sPublisher(namer Namer, tags []string) Interface {
	return &k3sPublisher{
		namer: namer,
		tags:  tags,
	}
}

//Publish implements publish.Interface
func (k *k3sPublisher) Publish(ctx context.Context, br build.Result, s string) (name.Reference, error) {
	s = strings.TrimPrefix(s, build.StrictScheme)
	s = strings.ToLower(s)

	img, err := ToImage(br, s)
	if err != nil {
		return nil, err
	}

	h, err := img.Digest()
	if err != nil {
		return nil, err
	}

	digestTag, err := name.NewTag(fmt.Sprintf("%s:%s", k.namer(K3sDomain, s), h.Hex))
	if err != nil {
		return nil, err
	}

	log.Printf("Loading %v", digestTag)
	if err := k3s.Write(ctx, digestTag, img); err != nil {
		return nil, err
	}
	log.Printf("Loaded %v", digestTag)

	for _, tagName := range k.tags {
		tag, err := name.NewTag(fmt.Sprintf("%s:%s", k.namer(K3sDomain, s), tagName))
		if err != nil {
			return nil, err
		}

		if err := k3s.Tag(ctx, digestTag, tag); err != nil {
			return nil, err
		}
		log.Printf("Added tag %v", tagName)
	}

	return &digestTag, nil
}

//Close implements publish.Interface
func (k *k3sPublisher) Close() error {
	return nil
}
