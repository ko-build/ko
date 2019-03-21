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
	"bytes"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/random"
	yaml "gopkg.in/yaml.v2"
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
	testBuilder = newFixedBuild(map[string]v1.Image{
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
			inputStructured := test.refs
			inputYAML, err := yaml.Marshal(inputStructured)
			if err != nil {
				t.Fatalf("yaml.Marshal(%v) = %v", inputStructured, err)
			}

			outYAML, err := ImageReferences(inputYAML, testBuilder, newFixedPublish(test.base, testHashes))
			if err != nil {
				t.Fatalf("ImageReferences(%v) = %v", string(inputYAML), err)
			}
			var outStructured []string
			if err := yaml.Unmarshal(outYAML, &outStructured); err != nil {
				t.Errorf("yaml.Unmarshal(%v) = %v", string(outYAML), err)
			}

			if want, got := len(inputStructured), len(outStructured); want != got {
				t.Errorf("ImageReferences(%v) = %v, want %v", string(inputYAML), got, want)
			}

			var expectedStructured []string
			for i, ref := range test.refs {
				hash := test.hashes[i]
				expectedStructured = append(expectedStructured,
					computeDigest(test.base, ref, hash))
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
		input:    map[string]string{"image": fooRef},
		expected: map[string]string{"image": computeDigest(base, fooRef, fooHash)},
	}, {
		desc:  "simple key",
		input: map[string]string{bazRef: "blah"},
		expected: map[string]string{
			computeDigest(base, bazRef, bazHash): "blah",
		},
	}, {
		desc:  "key and value",
		input: map[string]string{fooRef: barRef},
		expected: map[string]string{
			computeDigest(base, fooRef, fooHash): computeDigest(base, barRef, barHash),
		},
	}, {
		desc:     "empty map",
		input:    map[string]string{},
		expected: map[string]string{},
	}, {
		desc: "multiple values",
		input: map[string]string{
			"arg1": fooRef,
			"arg2": barRef,
		},
		expected: map[string]string{
			"arg1": computeDigest(base, fooRef, fooHash),
			"arg2": computeDigest(base, barRef, barHash),
		},
	}}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			inputStructured := test.input
			inputYAML, err := yaml.Marshal(inputStructured)
			if err != nil {
				t.Fatalf("yaml.Marshal(%v) = %v", inputStructured, err)
			}

			outYAML, err := ImageReferences(inputYAML, testBuilder, newFixedPublish(base, testHashes))
			if err != nil {
				t.Fatalf("ImageReferences(%v) = %v", string(inputYAML), err)
			}
			var outStructured map[string]string
			if err := yaml.Unmarshal(outYAML, &outStructured); err != nil {
				t.Errorf("yaml.Unmarshal(%v) = %v", string(outYAML), err)
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
		input:    &object{S: fooRef},
		expected: &object{S: computeDigest(base, fooRef, fooHash)},
	}, {
		desc:     "map field",
		input:    &object{M: map[string]object{"blah": {S: fooRef}}},
		expected: &object{M: map[string]object{"blah": {S: computeDigest(base, fooRef, fooHash)}}},
	}, {
		desc:     "array field",
		input:    &object{A: []object{{S: fooRef}}},
		expected: &object{A: []object{{S: computeDigest(base, fooRef, fooHash)}}},
	}, {
		desc:     "pointer field",
		input:    &object{P: &object{S: fooRef}},
		expected: &object{P: &object{S: computeDigest(base, fooRef, fooHash)}},
	}, {
		desc:     "deep field",
		input:    &object{M: map[string]object{"blah": {A: []object{{P: &object{S: fooRef}}}}}},
		expected: &object{M: map[string]object{"blah": {A: []object{{P: &object{S: computeDigest(base, fooRef, fooHash)}}}}}},
	}}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			inputStructured := test.input
			inputYAML, err := yaml.Marshal(inputStructured)
			if err != nil {
				t.Fatalf("yaml.Marshal(%v) = %v", inputStructured, err)
			}

			outYAML, err := ImageReferences(inputYAML, testBuilder, newFixedPublish(base, testHashes))
			if err != nil {
				t.Fatalf("ImageReferences(%v) = %v", string(inputYAML), err)
			}
			var outStructured *object
			if err := yaml.Unmarshal(outYAML, &outStructured); err != nil {
				t.Errorf("yaml.Unmarshal(%v) = %v", string(outYAML), err)
			}

			if diff := cmp.Diff(test.expected, outStructured, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("ImageReferences(%v); (-want +got) = %v", string(inputYAML), diff)
			}
		})
	}
}

func TestMultiDocumentYAMLs(t *testing.T) {
	tests := []struct {
		desc   string
		refs   []string
		hashes []v1.Hash
		base   name.Repository
	}{{
		desc:   "two string documents",
		refs:   []string{fooRef, barRef},
		hashes: []v1.Hash{fooHash, barHash},
		base:   mustRepository("gcr.io/multi-pass"),
	}}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			encoder := yaml.NewEncoder(buf)
			for _, input := range test.refs {
				if err := encoder.Encode(input); err != nil {
					t.Fatalf("Encode(%v) = %v", input, err)
				}
			}
			inputYAML := buf.Bytes()

			outYAML, err := ImageReferences(inputYAML, testBuilder, newFixedPublish(test.base, testHashes))
			if err != nil {
				t.Fatalf("ImageReferences(%v) = %v", string(inputYAML), err)
			}

			buf = bytes.NewBuffer(outYAML)
			decoder := yaml.NewDecoder(buf)
			var outStructured []string
			for {
				var output string
				if err := decoder.Decode(&output); err == nil {
					outStructured = append(outStructured, output)
				} else if err == io.EOF {
					outStructured = append(outStructured, output)
					break
				} else {
					t.Errorf("yaml.Unmarshal(%v) = %v", string(outYAML), err)
				}
			}

			var expectedStructured []string
			for i, ref := range test.refs {
				hash := test.hashes[i]
				expectedStructured = append(expectedStructured,
					computeDigest(test.base, ref, hash))
			}
			// The multi-document output always seems to leave a trailing --- so we end up with
			// an extra empty element.
			expectedStructured = append(expectedStructured, "")

			if want, got := len(expectedStructured), len(outStructured); want != got {
				t.Errorf("ImageReferences(%v) = %v, want %v", string(inputYAML), got, want)
			}

			if diff := cmp.Diff(expectedStructured, outStructured, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("ImageReferences(%v); (-want +got) = %v", string(inputYAML), diff)
			}
		})
	}
}

func mustRandom() v1.Image {
	img, err := random.Image(1024, 5)
	if err != nil {
		panic(err)
	}
	return img
}

func mustRepository(s string) name.Repository {
	n, err := name.NewRepository(s, name.WeakValidation)
	if err != nil {
		panic(err)
	}
	return n
}

func mustDigest(img v1.Image) v1.Hash {
	d, err := img.Digest()
	if err != nil {
		panic(err)
	}
	return d
}

func computeDigest(base name.Repository, ref string, h v1.Hash) string {
	d, err := newFixedPublish(base, map[string]v1.Hash{ref: h}).Publish(nil, ref)
	if err != nil {
		panic(err)
	}
	return d.String()
}
