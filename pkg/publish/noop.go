// Copyright 2022 Google LLC All Rights Reserved.
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
	"errors"
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/ko/pkg/build"
)

// Original spelling was preserved when this was refactored out of pkg/commands
type nopPublisher struct {
	repoName string
	namer    Namer
	tag      string
	tagOnly  bool
}

type noOpOpener struct {
	repoName string
	namer    Namer
	tags     []string
	tagOnly  bool
}

// NoOpOption provides functional options to the NoOp publisher.
type NoOpOption func(*noOpOpener)

func (o *noOpOpener) Open() (Interface, error) {
	tag := defaultTags[0]
	if o.tagOnly {
		// Replicate the tag-only validations in the default publisher
		if len(o.tags) != 1 {
			return nil, errors.New("must specify exactly one tag to resolve images into tag-only references")
		}
		if o.tags[0] == defaultTags[0] {
			return nil, errors.New("latest tag cannot be used in tag-only references")
		}
	}
	// If one or more tags are specified, use the first tag in the list
	if len(o.tags) >= 1 {
		tag = o.tags[0]
	}
	return &nopPublisher{
		repoName: o.repoName,
		namer:    o.namer,
		tag:      tag,
		tagOnly:  o.tagOnly,
	}, nil
}

// NewNoOp returns a publisher.Interface that simulates publishing without actually publishing
// anything, to provide fallback behavior when the user configures no push destinations.
func NewNoOp(baseName string, options ...NoOpOption) (Interface, error) {
	nop := &noOpOpener{
		repoName: baseName,
		namer:    identity,
	}
	for _, option := range options {
		option(nop)
	}
	return nop.Open()
}

// Publish implements publish.Interface
func (n *nopPublisher) Publish(_ context.Context, br build.Result, s string) (name.Reference, error) {
	s = strings.TrimPrefix(s, build.StrictScheme)
	h, err := br.Digest()
	if err != nil {
		return nil, err
	}
	// If the tag is not empty or is not "latest", use the :tag@sha suffix
	if n.tag != "" || n.tag != defaultTags[0] {
		// If tag only, just return the tag
		if n.tagOnly {
			return name.NewTag(fmt.Sprintf("%s:%s", n.namer(n.repoName, s), n.tag))
		}
		return name.NewDigest(fmt.Sprintf("%s:%s@%s", n.namer(n.repoName, s), n.tag, h))
	}
	return name.NewDigest(fmt.Sprintf("%s@%s", n.namer(n.repoName, s), h))
}

func (n *nopPublisher) Close() error { return nil }
