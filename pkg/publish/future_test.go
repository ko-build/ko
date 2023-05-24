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

package publish

import (
	"fmt"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/random"
)

func makeRef() (name.Reference, error) {
	img, err := random.Image(256, 8)
	if err != nil {
		return nil, err
	}
	d, err := img.Digest()
	if err != nil {
		return nil, err
	}
	return name.NewDigest(fmt.Sprintf("gcr.io/foo/bar@%s", d))
}

func TestSameFutureSameReference(t *testing.T) {
	f := newFuture(makeRef)

	ref1, err := f.Get()
	if err != nil {
		t.Errorf("Get() = %v", err)
	}
	d1 := ref1.String()

	ref2, err := f.Get()
	if err != nil {
		t.Errorf("Get() = %v", err)
	}
	d2 := ref2.String()

	if d1 != d2 {
		t.Errorf("Got different digests %s and %s", d1, d2)
	}
}

func TestDiffFutureDiffReference(t *testing.T) {
	f1 := newFuture(makeRef)
	f2 := newFuture(makeRef)

	ref1, err := f1.Get()
	if err != nil {
		t.Errorf("Get() = %v", err)
	}
	d1 := ref1.String()

	ref2, err := f2.Get()
	if err != nil {
		t.Errorf("Get() = %v", err)
	}
	d2 := ref2.String()

	if d1 == d2 {
		t.Errorf("Got same digest %s, wanted different", d1)
	}
}
