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

package resolve

import (
	"fmt"
	"testing"

	"github.com/google/go-containerregistry/pkg/ko/build"
	"github.com/google/go-containerregistry/pkg/ko/publish"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/random"
)

var (
	fixedBaseRepo, _ = name.NewRepository("gcr.io/asdf", name.WeakValidation)
	testImage, _     = random.Image(1024, 5)
)

func TestFixedPublish(t *testing.T) {
	hex1 := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	hex2 := "baadf00dbaadf00dbaadf00dbaadf00dbaadf00dbaadf00dbaadf00dbaadf00d"
	f := newFixedPublish(fixedBaseRepo, map[string]v1.Hash{
		"foo": {
			Algorithm: "sha256",
			Hex:       hex1,
		},
		"bar": {
			Algorithm: "sha256",
			Hex:       hex2,
		},
	})

	fooDigest, err := f.Publish(nil, "foo")
	if err != nil {
		t.Errorf("Publish(foo) = %v", err)
	}
	if got, want := fooDigest.String(), "gcr.io/asdf/foo@sha256:"+hex1; got != want {
		t.Errorf("Publish(foo) = %q, want %q", got, want)
	}

	barDigest, err := f.Publish(nil, "bar")
	if err != nil {
		t.Errorf("Publish(bar) = %v", err)
	}
	if got, want := barDigest.String(), "gcr.io/asdf/bar@sha256:"+hex2; got != want {
		t.Errorf("Publish(bar) = %q, want %q", got, want)
	}

	d, err := f.Publish(nil, "baz")
	if err == nil {
		t.Errorf("Publish(baz) = %v, want error", d)
	}
}

func TestFixedBuild(t *testing.T) {
	f := newFixedBuild(map[string]v1.Image{
		"asdf": testImage,
	})

	if got, want := f.IsSupportedReference("asdf"), true; got != want {
		t.Errorf("IsSupportedReference(asdf) = %v, want %v", got, want)
	}
	if got, err := f.Build("asdf"); err != nil {
		t.Errorf("Build(asdf) = %v, want %v", err, testImage)
	} else if got != testImage {
		t.Errorf("Build(asdf) = %v, want %v", got, testImage)
	}

	if got, want := f.IsSupportedReference("blah"), false; got != want {
		t.Errorf("IsSupportedReference(blah) = %v, want %v", got, want)
	}
	if got, err := f.Build("blah"); err == nil {
		t.Errorf("Build(blah) = %v, want error", got)
	}
}

type fixedBuild struct {
	entries map[string]v1.Image
}

// newFixedBuild returns a build.Interface implementation that simply resolves
// particular references to fixed v1.Image objects
func newFixedBuild(entries map[string]v1.Image) build.Interface {
	return &fixedBuild{entries}
}

// IsSupportedReference implements build.Interface
func (f *fixedBuild) IsSupportedReference(s string) bool {
	_, ok := f.entries[s]
	return ok
}

// Build implements build.Interface
func (f *fixedBuild) Build(s string) (v1.Image, error) {
	if img, ok := f.entries[s]; ok {
		return img, nil
	}
	return nil, fmt.Errorf("unsupported reference: %q", s)
}

type fixedPublish struct {
	base    name.Repository
	entries map[string]v1.Hash
}

// newFixedPublish returns a publish.Interface implementation that simply
// resolves particular references to fixed name.Digest references.
func newFixedPublish(base name.Repository, entries map[string]v1.Hash) publish.Interface {
	return &fixedPublish{base, entries}
}

// Publish implements publish.Interface
func (f *fixedPublish) Publish(_ v1.Image, s string) (name.Reference, error) {
	h, ok := f.entries[s]
	if !ok {
		return nil, fmt.Errorf("unsupported importpath: %q", s)
	}
	d, err := name.NewDigest(fmt.Sprintf("%s/%s@%s", f.base, s, h), name.WeakValidation)
	if err != nil {
		return nil, err
	}
	return &d, nil
}
