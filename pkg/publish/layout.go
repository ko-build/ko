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

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/layout"
)

type LayoutPublisher struct {
	p layout.Path
}

// NewLayout returns a new publish.Interface that saves images to an OCI Image Layout.
func NewLayout(path string) (Interface, error) {
	p, err := layout.FromPath(path)
	if err != nil {
		p, err = layout.Write(path, empty.Index)
		if err != nil {
			return nil, err
		}
	}
	return &LayoutPublisher{p}, nil
}

// Publish implements publish.Interface.
func (l *LayoutPublisher) Publish(img v1.Image, s string) (name.Reference, error) {
	log.Printf("Saving %v", s)
	if err := l.p.AppendImage(img); err != nil {
		return nil, err
	}
	log.Printf("Saved %v", s)

	h, err := img.Digest()
	if err != nil {
		return nil, err
	}

	dig, err := name.NewDigest(fmt.Sprintf("%s@%s", l.p, h))
	if err != nil {
		return nil, err
	}

	return dig, nil
}

func (l *LayoutPublisher) Close() error {
	return nil
}
