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

package commands

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/ko/pkg/build"
	"github.com/google/ko/pkg/commands/options"
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
	testBuilder = kotesting.NewFixedBuild(map[string]build.Result{
		fooRef: foo,
		barRef: bar,
	})
	testHashes = map[string]v1.Hash{
		fooRef: fooHash,
		barRef: barHash,
	}
)

func TestResolveMultiDocumentYAMLs(t *testing.T) {
	refs := []string{fooRef, barRef}
	hashes := []v1.Hash{fooHash, barHash}
	base := mustRepository("gcr.io/multi-pass")

	buf := bytes.NewBuffer(nil)
	encoder := yaml.NewEncoder(buf)
	for _, input := range refs {
		if err := encoder.Encode(input); err != nil {
			t.Fatalf("Encode(%v) = %v", input, err)
		}
	}

	inputYAML := buf.Bytes()

	outYAML, err := resolveFile(
		context.Background(),
		yamlToTmpFile(t, buf.Bytes()),
		testBuilder,
		kotesting.NewFixedPublish(base, testHashes),
		&options.SelectorOptions{},
		&options.StrictOptions{})

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
			break
		} else {
			t.Errorf("yaml.Unmarshal(%v) = %v", string(outYAML), err)
		}
	}

	expectedStructured := []string{
		kotesting.ComputeDigest(base, refs[0], hashes[0]),
		kotesting.ComputeDigest(base, refs[1], hashes[1]),
	}

	if want, got := len(expectedStructured), len(outStructured); want != got {
		t.Errorf("resolveFile(%v) = %v, want %v", string(inputYAML), got, want)
	}

	if diff := cmp.Diff(expectedStructured, outStructured, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("resolveFile(%v); (-want +got) = %v", string(inputYAML), diff)
	}
}

func TestResolveMultiDocumentYAMLsWithSelector(t *testing.T) {
	passesSelector := `apiVersion: something/v1
kind: Foo
metadata:
  labels:
    qux: baz
`
	failsSelector := `apiVersion: other/v2
kind: Bar
`
	// Note that this ends in '---', so it in ends in a final null YAML document.
	inputYAML := []byte(fmt.Sprintf("%s---\n%s---", passesSelector, failsSelector))
	base := mustRepository("gcr.io/multi-pass")

	outputYAML, err := resolveFile(
		context.Background(),
		yamlToTmpFile(t, inputYAML),
		testBuilder,
		kotesting.NewFixedPublish(base, testHashes),
		&options.SelectorOptions{
			Selector: "qux=baz",
		},
		&options.StrictOptions{})
	if err != nil {
		t.Fatalf("ImageReferences(%v) = %v", string(inputYAML), err)
	}
	if diff := cmp.Diff(passesSelector, string(outputYAML)); diff != "" {
		t.Errorf("resolveFile (-want +got) = %v", diff)
	}
}

func mustRepository(s string) name.Repository {
	n, err := name.NewRepository(s)
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

func mustRandom() v1.Image {
	img, err := random.Image(1024, 5)
	if err != nil {
		panic(err)
	}
	return img
}

func yamlToTmpFile(t *testing.T, yaml []byte) string {
	t.Helper()

	tmpfile, err := ioutil.TempFile("", "doc")
	if err != nil {
		t.Fatalf("error creating temp file: %v", err)
	}

	if _, err := tmpfile.Write(yaml); err != nil {
		t.Fatalf("error writing temp file: %v", err)
	}

	if err := tmpfile.Close(); err != nil {
		t.Fatalf("error closing temp file: %v", err)
	}

	return tmpfile.Name()
}
