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

package commands

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/go-containerregistry/pkg/v1/empty"
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

	errImageLoad = fmt.Errorf("ImageLoad() error")
	errImageTag  = fmt.Errorf("ImageTag() error")
)

type erroringClient struct {
	daemon.Client

	inspectErr  error
	inspectResp types.ImageInspect
	inspectBody []byte
}

func (m *erroringClient) NegotiateAPIVersion(context.Context) {}
func (m *erroringClient) ImageLoad(context.Context, io.Reader, bool) (types.ImageLoadResponse, error) {
	return types.ImageLoadResponse{}, errImageLoad
}
func (m *erroringClient) ImageTag(_ context.Context, _ string, _ string) error {
	return errImageTag
}

func (m *erroringClient) ImageInspectWithRaw(_ context.Context, _ string) (types.ImageInspect, []byte, error) {
	return m.inspectResp, m.inspectBody, m.inspectErr
}

func TestResolveMultiDocumentYAMLs(t *testing.T) {
	refs := []string{fooRef, barRef}
	hashes := []v1.Hash{fooHash, barHash}
	base := mustRepository("gcr.io/multi-pass")

	buf := bytes.NewBuffer(nil)
	encoder := yaml.NewEncoder(buf)
	for _, input := range refs {
		if err := encoder.Encode(build.StrictScheme + input); err != nil {
			t.Fatalf("Encode(%v) = %v", input, err)
		}
	}

	inputYAML := buf.Bytes()

	outYAML, err := resolveFile(
		context.Background(),
		yamlToTmpFile(t, buf.Bytes()),
		testBuilder,
		kotesting.NewFixedPublish(base, testHashes),
		&options.SelectorOptions{})

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
		} else if errors.Is(err, io.EOF) {
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
		})
	if err != nil {
		t.Fatalf("ImageReferences(%v) = %v", string(inputYAML), err)
	}
	if diff := cmp.Diff(passesSelector, string(outputYAML)); diff != "" {
		t.Errorf("resolveFile (-want +got) = %v", diff)
	}
}

func TestNewBuilder(t *testing.T) {
	namespace := "base"
	s, err := registryServerWithImage(namespace)
	if err != nil {
		t.Fatalf("could not create test registry server: %v", err)
	}
	defer s.Close()
	baseImage := fmt.Sprintf("%s/%s", s.Listener.Addr().String(), namespace)

	tests := []struct {
		description             string
		importpath              string
		bo                      *options.BuildOptions
		wantQualifiedImportpath string
		shouldBuildError        bool
	}{
		{
			description: "test app with already qualified import path",
			importpath:  "ko://github.com/google/ko/test",
			bo: &options.BuildOptions{
				BaseImage:        baseImage,
				ConcurrentBuilds: 1,
				Platforms:        []string{"all"},
			},
			wantQualifiedImportpath: "ko://github.com/google/ko/test",
			shouldBuildError:        false,
		},
		{
			description: "programmatic build config",
			importpath:  "./test",
			bo: &options.BuildOptions{
				BaseImage: baseImage,
				BuildConfigs: map[string]build.Config{
					"github.com/google/ko/test": {
						ID: "id-can-be-anything",
						// no easy way to assert on the output, so trigger error to ensure config is picked up
						Flags: []string{"-invalid-flag-should-cause-error"},
					},
				},
				ConcurrentBuilds: 1,
				WorkingDirectory: "../..",
			},
			wantQualifiedImportpath: "ko://github.com/google/ko/test",
			shouldBuildError:        true,
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			ctx := context.Background()
			builder, err := NewBuilder(ctx, test.bo)
			if err != nil {
				t.Fatalf("NewBuilder(): %v", err)
			}
			qualifiedImportpath, err := builder.QualifyImport(test.importpath)
			if err != nil {
				t.Fatalf("builder.QualifyImport(%s): %v", test.importpath, err)
			}
			if qualifiedImportpath != test.wantQualifiedImportpath {
				t.Fatalf("incorrect qualified import path, got %s, wanted %s", qualifiedImportpath, test.wantQualifiedImportpath)
			}
			_, err = builder.Build(ctx, qualifiedImportpath)
			if err != nil && !test.shouldBuildError {
				t.Fatalf("builder.Build(): %v", err)
			}
			if err == nil && test.shouldBuildError {
				t.Fatalf("expected error got nil")
			}
		})
	}
}

func TestNewPublisherCanPublish(t *testing.T) {
	dockerRepo := "registry.example.com/repo"
	localDomain := "localdomain.example.com/repo"
	importpath := "github.com/google/ko/test"
	tests := []struct {
		description   string
		wantImageName string
		po            *options.PublishOptions
		shouldError   bool
		wantError     error
	}{
		{
			description:   "base import path",
			wantImageName: fmt.Sprintf("%s/%s", dockerRepo, path.Base(importpath)),
			po: &options.PublishOptions{
				BaseImportPaths: true,
				DockerRepo:      dockerRepo,
			},
		},
		{
			description:   "preserve import path",
			wantImageName: fmt.Sprintf("%s/%s", dockerRepo, importpath),
			po: &options.PublishOptions{
				DockerRepo:          dockerRepo,
				PreserveImportPaths: true,
			},
		},
		{
			description:   "override LocalDomain",
			wantImageName: fmt.Sprintf("%s/%s", localDomain, importpath),
			po: &options.PublishOptions{
				Local:               true,
				LocalDomain:         localDomain,
				PreserveImportPaths: true,
				DockerClient:        &kotesting.MockDaemon{},
			},
		},
		{
			description:   "override DockerClient",
			wantImageName: strings.ToLower(fmt.Sprintf("%s/%s", localDomain, importpath)),
			po: &options.PublishOptions{
				DockerClient: &erroringClient{},
				Local:        true,
			},
			shouldError: true,
			wantError:   errImageTag,
		},
		{
			description:   "bare with local domain and repo",
			wantImageName: strings.ToLower(fmt.Sprintf("%s/foo", dockerRepo)),
			po: &options.PublishOptions{
				DockerRepo: dockerRepo + "/foo",
				Local:      true,
				Bare:       true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			publisher, err := NewPublisher(test.po)
			if err != nil {
				t.Fatalf("NewPublisher(): %v", err)
			}
			defer publisher.Close()
			ref, err := publisher.Publish(context.Background(), empty.Image, build.StrictScheme+importpath)
			if test.shouldError {
				if err == nil || !strings.HasSuffix(err.Error(), test.wantError.Error()) {
					t.Errorf("%s: got error %v, wanted %v", test.description, err, test.wantError)
				}
				return
			}
			if err != nil {
				t.Fatalf("publisher.Publish(): %v", err)
			}
			gotImageName := ref.Context().Name()
			if gotImageName != test.wantImageName {
				t.Errorf("got %s, wanted %s", gotImageName, test.wantImageName)
			}
		})
	}
}

// registryServerWithImage starts a local registry and pushes a random image.
// Use this to speed up tests, by not having to reach out to gcr.io for the default base image.
// The registry uses a NOP logger to avoid spamming test logs.
// Remember to call `defer Close()` on the returned `httptest.Server`.
func registryServerWithImage(namespace string) (*httptest.Server, error) {
	nopLog := log.New(io.Discard, "", 0)
	r := registry.New(registry.Logger(nopLog))
	s := httptest.NewServer(r)
	imageName := fmt.Sprintf("%s/%s", s.Listener.Addr().String(), namespace)
	image, err := random.Image(1024, 1)
	if err != nil {
		return nil, fmt.Errorf("random.Image(): %w", err)
	}
	crane.Push(image, imageName)
	return s, nil
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

	tmpfile, err := os.CreateTemp("", "doc")
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
