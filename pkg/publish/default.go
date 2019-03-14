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
	"net/http"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// defalt is intentionally misspelled to avoid keyword collision (and drive Jon nuts).
type defalt struct {
	base  string
	t     http.RoundTripper
	auth  authn.Authenticator
	namer Namer
	tags  []string
}

// Option is a functional option for NewDefault.
type Option func(*defaultOpener) error

type defaultOpener struct {
	base  string
	t     http.RoundTripper
	auth  authn.Authenticator
	namer Namer
	tags  []string
}

// Namer is a function from a supported import path to the portion of the resulting
// image name that follows the "base" repository name.
type Namer func(string) string

// identity is the default namer, so import paths are affixed as-is under the repository
// name for maximum clarity, e.g.
//   gcr.io/foo/github.com/bar/baz/cmd/blah
//   ^--base--^ ^-------import path-------^
func identity(in string) string { return in }

// As some registries do not support pushing an image by digest, the default tag for pushing
// is the 'latest' tag.
var defaultTags = []string{"latest"}

func (do *defaultOpener) Open() (Interface, error) {
	return &defalt{
		base:  do.base,
		t:     do.t,
		auth:  do.auth,
		namer: do.namer,
		tags:  do.tags,
	}, nil
}

// NewDefault returns a new publish.Interface that publishes references under the provided base
// repository using the default keychain to authenticate and the default naming scheme.
func NewDefault(base string, options ...Option) (Interface, error) {
	do := &defaultOpener{
		base:  base,
		t:     http.DefaultTransport,
		auth:  authn.Anonymous,
		namer: identity,
		tags:  defaultTags,
	}

	for _, option := range options {
		if err := option(do); err != nil {
			return nil, err
		}
	}
	return do.Open()
}

// Publish implements publish.Interface
func (d *defalt) Publish(img v1.Image, s string) (name.Reference, error) {
	// https://github.com/google/go-containerregistry/issues/212
	s = strings.ToLower(s)

	for _, tagName := range d.tags {
		tag, err := name.NewTag(fmt.Sprintf("%s/%s:%s", d.base, d.namer(s), tagName), name.WeakValidation)
		if err != nil {
			return nil, err
		}

		log.Printf("Publishing %v", tag)
		// TODO: This is slow because we have to load the image multiple times.
		// Figure out some way to publish the manifest with another tag.
		if err := remote.Write(tag, img, d.auth, d.t); err != nil {
			return nil, err
		}
	}

	h, err := img.Digest()
	if err != nil {
		return nil, err
	}
	dig, err := name.NewDigest(fmt.Sprintf("%s/%s@%s", d.base, d.namer(s), h), name.WeakValidation)
	if err != nil {
		return nil, err
	}
	log.Printf("Published %v", dig)
	return &dig, nil
}
