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

package build

import (
	"testing"

	"github.com/google/go-containerregistry/pkg/v1/random"
)

func makeImage() (Result, error) {
	return random.Index(256, 8, 1)
}

func digest(t *testing.T, img Result) string {
	d, err := img.Digest()
	if err != nil {
		t.Fatalf("Digest() = %v", err)
	}
	return d.String()
}

func TestSameFutureSameImage(t *testing.T) {
	f := newFuture(makeImage)

	i1, err := f.Get()
	if err != nil {
		t.Errorf("Get() = %v", err)
	}
	d1 := digest(t, i1)

	i2, err := f.Get()
	if err != nil {
		t.Errorf("Get() = %v", err)
	}
	d2 := digest(t, i2)

	if d1 != d2 {
		t.Errorf("Got different digests %s and %s", d1, d2)
	}
}

func TestDiffFutureDiffImage(t *testing.T) {
	f1 := newFuture(makeImage)
	f2 := newFuture(makeImage)

	i1, err := f1.Get()
	if err != nil {
		t.Errorf("Get() = %v", err)
	}
	d1 := digest(t, i1)

	i2, err := f2.Get()
	if err != nil {
		t.Errorf("Get() = %v", err)
	}
	d2 := digest(t, i2)

	if d1 == d2 {
		t.Errorf("Got same digest %s, wanted different", d1)
	}
}
