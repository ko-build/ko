// Copyright 2021 ko Build Authors All Rights Reserved.
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
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/google/ko/pkg/build"
	"github.com/google/ko/pkg/commands/options"
)

func TestPublishImages(t *testing.T) {
	namespace := "base"
	s, err := registryServerWithImage(namespace)
	if err != nil {
		t.Fatalf("could not create test registry server: %v", err)
	}
	repo := s.Listener.Addr().String()
	baseImage := fmt.Sprintf("%s/%s", repo, namespace)
	sampleAppDir, err := sampleAppRelDir()
	if err != nil {
		t.Fatalf("sampleAppRelDir(): %v", err)
	}
	tests := []struct {
		description string
		publishArg  string
		importpath  string
	}{
		{
			description: "import path with ko scheme",
			publishArg:  "ko://github.com/google/ko/test",
			importpath:  "github.com/google/ko/test",
		},
		{
			description: "import path without ko scheme",
			publishArg:  "github.com/google/ko/test",
			importpath:  "github.com/google/ko/test",
		},
		{
			description: "file path",
			publishArg:  sampleAppDir,
			importpath:  "github.com/google/ko/test",
		},
	}
	for _, test := range tests {
		ctx := context.Background()
		bo := &options.BuildOptions{
			BaseImage:        baseImage,
			ConcurrentBuilds: 1,
			Platforms:        []string{"all"},
		}
		builder, err := NewBuilder(ctx, bo)
		if err != nil {
			t.Fatalf("%s: MakeBuilder(): %v", test.description, err)
		}
		po := &options.PublishOptions{
			DockerRepo:          repo,
			PreserveImportPaths: true,
		}
		publisher, err := NewPublisher(po)
		if err != nil {
			t.Fatalf("%s: MakePublisher(): %v", test.description, err)
		}
		importpathWithScheme := build.StrictScheme + test.importpath
		refs, err := PublishImages(ctx, []string{test.publishArg}, publisher, builder)
		if err != nil {
			t.Fatalf("%s: PublishImages(): %v", test.description, err)
		}
		ref, exists := refs[importpathWithScheme]
		if !exists {
			t.Errorf("%s: could not find image for importpath %s", test.description, importpathWithScheme)
		}
		gotImageName := ref.Context().Name()
		wantImageName := strings.ToLower(fmt.Sprintf("%s/%s", repo, test.importpath))
		if gotImageName != wantImageName {
			t.Errorf("%s: got %s, wanted %s", test.description, gotImageName, wantImageName)
		}
	}
}

func sampleAppRelDir() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("could not get current filename")
	}
	basepath := filepath.Dir(filename)
	testAppDir := filepath.Join(basepath, "..", "..", "test")
	return filepath.Rel(basepath, testAppDir)
}
