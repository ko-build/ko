// Copyright 2018 ko Build Authors All Rights Reserved.
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

package testing

import (
	"context"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/ko/pkg/build"
)

func TestFixedPublish(t *testing.T) {
	hex1 := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	hex2 := "baadf00dbaadf00dbaadf00dbaadf00dbaadf00dbaadf00dbaadf00dbaadf00d"
	fixedBaseRepo, _ := name.NewRepository("gcr.io/asdf")
	f := NewFixedPublish(fixedBaseRepo, map[string]v1.Hash{
		"foo": {
			Algorithm: "sha256",
			Hex:       hex1,
		},
		"bar": {
			Algorithm: "sha256",
			Hex:       hex2,
		},
	})

	fooDigest, err := f.Publish(context.Background(), nil, "foo")
	if err != nil {
		t.Errorf("Publish(foo) = %v", err)
	}
	if got, want := fooDigest.String(), "gcr.io/asdf/foo@sha256:"+hex1; got != want {
		t.Errorf("Publish(foo) = %q, want %q", got, want)
	}

	barDigest, err := f.Publish(context.Background(), nil, "bar")
	if err != nil {
		t.Errorf("Publish(bar) = %v", err)
	}
	if got, want := barDigest.String(), "gcr.io/asdf/bar@sha256:"+hex2; got != want {
		t.Errorf("Publish(bar) = %q, want %q", got, want)
	}

	d, err := f.Publish(context.Background(), nil, "baz")
	if err == nil {
		t.Errorf("Publish(baz) = %v, want error", d)
	}
}

func TestFixedBuild(t *testing.T) {
	testImage, _ := random.Image(1024, 5)
	f := NewFixedBuild(map[string]build.Result{
		"asdf": testImage,
	})

	if got := f.IsSupportedReference("asdf"); got != nil {
		t.Errorf("IsSupportedReference(asdf) = (%v), want nil", got)
	}
	if got, err := f.Build(context.Background(), "asdf"); err != nil {
		t.Errorf("Build(asdf) = %v, want %v", err, testImage)
	} else if got != testImage {
		t.Errorf("Build(asdf) = %v, want %v", got, testImage)
	}

	if got := f.IsSupportedReference("blah"); got == nil {
		t.Error("IsSupportedReference(blah) = nil, want error")
	}
	if got, err := f.Build(context.Background(), "blah"); err == nil {
		t.Errorf("Build(blah) = %v, want error", got)
	}
}
