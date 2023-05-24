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

package resolve

import (
	"bytes"
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/ko/pkg/build"
	kotesting "github.com/google/ko/pkg/internal/testing"
	"gopkg.in/yaml.v3"
)

var (
	fooRef      = "github.com/awesomesauce/foo"
	foo         = mustRandom()
	fooHash     = mustDigest(foo)
	barRef      = "github.com/awesomesauce/bar"
	bar         = mustRandom()
	barHash     = mustDigest(bar)
	bazRef      = "github.com/awesomesauce/baz"
	baz         = mustRandom()
	bazHash     = mustDigest(baz)
	testBuilder = kotesting.NewFixedBuild(map[string]build.Result{
		fooRef: foo,
		barRef: bar,
		bazRef: baz,
	})
	testHashes = map[string]v1.Hash{
		fooRef: fooHash,
		barRef: barHash,
		bazRef: bazHash,
	}
)

func TestYAMLArrays(t *testing.T) {
	tests := []struct {
		desc   string
		refs   []string
		hashes []v1.Hash
		base   name.Repository
	}{{
		desc:   "singleton array",
		refs:   []string{fooRef},
		hashes: []v1.Hash{fooHash},
		base:   mustRepository("gcr.io/mattmoor"),
	}, {
		desc:   "singleton array (different base)",
		refs:   []string{fooRef},
		hashes: []v1.Hash{fooHash},
		base:   mustRepository("gcr.io/jasonhall"),
	}, {
		desc:   "two element array",
		refs:   []string{fooRef, barRef},
		hashes: []v1.Hash{fooHash, barHash},
		base:   mustRepository("gcr.io/jonjohnson"),
	}, {
		desc:   "empty array",
		refs:   []string{},
		hashes: []v1.Hash{},
		base:   mustRepository("gcr.io/blah"),
	}}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			inputStructured := []string{}
			for _, ref := range test.refs {
				inputStructured = append(inputStructured, build.StrictScheme+ref)
			}
			inputYAML, err := yaml.Marshal(inputStructured)
			if err != nil {
				t.Fatalf("yaml.Marshal(%v) = %v", inputStructured, err)
			}

			doc := strToYAML(t, string(inputYAML))
			err = ImageReferences(context.Background(), []*yaml.Node{doc}, testBuilder, kotesting.NewFixedPublish(test.base, testHashes))
			if err != nil {
				t.Fatalf("ImageReferences(%v) = %v", string(inputYAML), err)
			}
			var outStructured []string
			if err := doc.Decode(&outStructured); err != nil {
				t.Errorf("doc.Decode(%v) = %v", yamlToStr(t, doc), err)
			}

			if want, got := len(inputStructured), len(outStructured); want != got {
				t.Errorf("ImageReferences(%v) = %v, want %v", string(inputYAML), got, want)
			}

			var expectedStructured []string
			for i, ref := range test.refs {
				hash := test.hashes[i]
				expectedStructured = append(expectedStructured,
					kotesting.ComputeDigest(test.base, ref, hash))
			}

			if diff := cmp.Diff(expectedStructured, outStructured, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("ImageReferences(%v); (-want +got) = %v", string(inputYAML), diff)
			}
		})
	}
}

func TestYAMLMaps(t *testing.T) {
	base := mustRepository("gcr.io/mattmoor")
	tests := []struct {
		desc     string
		input    map[string]string
		expected map[string]string
	}{{
		desc:     "simple value",
		input:    map[string]string{"image": build.StrictScheme + fooRef},
		expected: map[string]string{"image": kotesting.ComputeDigest(base, fooRef, fooHash)},
	}, {
		desc:  "simple key",
		input: map[string]string{build.StrictScheme + bazRef: "blah"},
		expected: map[string]string{
			kotesting.ComputeDigest(base, bazRef, bazHash): "blah",
		},
	}, {
		desc:  "key and value",
		input: map[string]string{build.StrictScheme + fooRef: build.StrictScheme + barRef},
		expected: map[string]string{
			kotesting.ComputeDigest(base, fooRef, fooHash): kotesting.ComputeDigest(base, barRef, barHash),
		},
	}, {
		desc:     "empty map",
		input:    map[string]string{},
		expected: map[string]string{},
	}, {
		desc: "multiple values",
		input: map[string]string{
			"arg1": build.StrictScheme + fooRef,
			"arg2": build.StrictScheme + barRef,
		},
		expected: map[string]string{
			"arg1": kotesting.ComputeDigest(base, fooRef, fooHash),
			"arg2": kotesting.ComputeDigest(base, barRef, barHash),
		},
	}}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			inputStructured := test.input
			inputYAML, err := yaml.Marshal(inputStructured)
			if err != nil {
				t.Fatalf("yaml.Marshal(%v) = %v", inputStructured, err)
			}

			doc := strToYAML(t, string(inputYAML))
			err = ImageReferences(context.Background(), []*yaml.Node{doc}, testBuilder, kotesting.NewFixedPublish(base, testHashes))
			if err != nil {
				t.Fatalf("ImageReferences(%v) = %v", string(inputYAML), err)
			}
			var outStructured map[string]string
			if err := doc.Decode(&outStructured); err != nil {
				t.Errorf("doc.Decode(%v) = %v", yamlToStr(t, doc), err)
			}

			if want, got := len(inputStructured), len(outStructured); want != got {
				t.Errorf("ImageReferences(%v) = %v, want %v", string(inputYAML), got, want)
			}

			if diff := cmp.Diff(test.expected, outStructured, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("ImageReferences(%v); (-want +got) = %v", string(inputYAML), diff)
			}
		})
	}
}

// object has public fields to avoid `yaml:"foo"` annotations.
type object struct {
	S string
	M map[string]object
	A []object
	P *object
}

func TestYAMLObject(t *testing.T) {
	base := mustRepository("gcr.io/bazinga")
	tests := []struct {
		desc     string
		input    *object
		expected *object
	}{{
		desc:     "empty object",
		input:    &object{},
		expected: &object{},
	}, {
		desc:     "string field",
		input:    &object{S: build.StrictScheme + fooRef},
		expected: &object{S: kotesting.ComputeDigest(base, fooRef, fooHash)},
	}, {
		desc:     "map field",
		input:    &object{M: map[string]object{"blah": {S: build.StrictScheme + fooRef}}},
		expected: &object{M: map[string]object{"blah": {S: kotesting.ComputeDigest(base, fooRef, fooHash)}}},
	}, {
		desc:     "array field",
		input:    &object{A: []object{{S: build.StrictScheme + fooRef}}},
		expected: &object{A: []object{{S: kotesting.ComputeDigest(base, fooRef, fooHash)}}},
	}, {
		desc:     "pointer field",
		input:    &object{P: &object{S: build.StrictScheme + fooRef}},
		expected: &object{P: &object{S: kotesting.ComputeDigest(base, fooRef, fooHash)}},
	}, {
		desc:     "deep field",
		input:    &object{M: map[string]object{"blah": {A: []object{{P: &object{S: build.StrictScheme + fooRef}}}}}},
		expected: &object{M: map[string]object{"blah": {A: []object{{P: &object{S: kotesting.ComputeDigest(base, fooRef, fooHash)}}}}}},
	}}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			inputStructured := test.input
			inputYAML, err := yaml.Marshal(inputStructured)
			if err != nil {
				t.Fatalf("yaml.Marshal(%v) = %v", inputStructured, err)
			}

			doc := strToYAML(t, string(inputYAML))
			err = ImageReferences(context.Background(), []*yaml.Node{doc}, testBuilder, kotesting.NewFixedPublish(base, testHashes))
			if err != nil {
				t.Fatalf("ImageReferences(%v) = %v", string(inputYAML), err)
			}
			var outStructured *object
			if err := doc.Decode(&outStructured); err != nil {
				t.Errorf("doc.Decode(%v) = %v", yamlToStr(t, doc), err)
			}

			if diff := cmp.Diff(test.expected, outStructured, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("ImageReferences(%v); (-want +got) = %v", string(inputYAML), diff)
			}
		})
	}
}

func TestStrict(t *testing.T) {
	refs := []string{
		fooRef,
		barRef,
	}
	buf := bytes.NewBuffer(nil)
	encoder := yaml.NewEncoder(buf)
	for _, input := range refs {
		if err := encoder.Encode(build.StrictScheme + input); err != nil {
			t.Fatalf("Encode(%v) = %v", input, err)
		}
	}
	base := mustRepository("gcr.io/multi-pass")
	doc := strToYAML(t, buf.String())

	err := ImageReferences(context.Background(), []*yaml.Node{doc}, testBuilder, kotesting.NewFixedPublish(base, testHashes))
	if err != nil {
		t.Fatalf("ImageReferences: %v", err)
	}
	t.Log(yamlToStr(t, doc))
}

func TestIsSupportedReferenceError(t *testing.T) {
	ref := build.StrictScheme + fooRef

	buf := bytes.NewBuffer(nil)
	encoder := yaml.NewEncoder(buf)
	if err := encoder.Encode(ref); err != nil {
		t.Fatalf("Encode(%v) = %v", ref, err)
	}

	base := mustRepository("gcr.io/multi-pass")
	doc := strToYAML(t, buf.String())

	noMatchBuilder := kotesting.NewFixedBuild(nil)

	err := ImageReferences(context.Background(), []*yaml.Node{doc}, noMatchBuilder, kotesting.NewFixedPublish(base, testHashes))
	if err == nil {
		t.Fatal("ImageReferences should err, got nil")
	}
}

func mustRandom() build.Result {
	img, err := random.Index(1024, 5, 1)
	if err != nil {
		panic(err)
	}
	return img
}

func mustRepository(s string) name.Repository {
	n, err := name.NewRepository(s)
	if err != nil {
		panic(err)
	}
	return n
}

func mustDigest(img build.Result) v1.Hash {
	d, err := img.Digest()
	if err != nil {
		panic(err)
	}
	return d
}
