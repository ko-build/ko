/*
Copyright 2021 Google LLC All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package commands

import (
	"context"
	"testing"

	"github.com/google/ko/pkg/build"
	"github.com/google/ko/pkg/commands/options"
)

func TestOverrideDefaultBaseImageUsingBuildOption(t *testing.T) {
	wantDigest := "sha256:76c39a6f76890f8f8b026f89e081084bc8c64167d74e6c93da7a053cb4ccb5dd"
	wantImage := "gcr.io/distroless/static-debian9@" + wantDigest
	bo := &options.BuildOptions{
		BaseImage: wantImage,
	}

	baseFn := getBaseImage("all", bo)
	_, res, err := baseFn(context.Background(), "ko://example.com/helloworld")
	if err != nil {
		t.Fatalf("getBaseImage(): %v", err)
	}

	digest, err := res.Digest()
	if err != nil {
		t.Fatalf("res.Digest(): %v", err)
	}
	gotDigest := digest.String()
	if gotDigest != wantDigest {
		t.Errorf("got digest %s, wanted %s", gotDigest, wantDigest)
	}
}

// TestDefaultBaseImage is a canary-type test for ensuring that config has been read when creating a builder.
func TestDefaultBaseImage(t *testing.T) {
	_, err := NewBuilder(context.Background(), &options.BuildOptions{
		WorkingDirectory: "testdata/config",
	})
	if err != nil {
		t.Fatal(err)
	}
	wantDefaultBaseImage := "gcr.io/distroless/base:nonroot" // matches value in ./testdata/.ko.yaml
	if defaultBaseImage != wantDefaultBaseImage {
		t.Fatalf("wanted defaultBaseImage %s, got %s", wantDefaultBaseImage, defaultBaseImage)
	}
}

func TestBuildConfigWithWorkingDirectoryAndDirAndMain(t *testing.T) {
	_, err := NewBuilder(context.Background(), &options.BuildOptions{
		WorkingDirectory: "testdata/paths",
	})
	if err != nil {
		t.Fatalf("NewBuilder(): %+v", err)
	}

	if len(buildConfigs) != 1 {
		t.Fatalf("expected 1 build config, got %d", len(buildConfigs))
	}
	expectedImportPath := "example.com/testapp/cmd/foo" // module from app/go.mod + `main` from .ko.yaml
	if _, exists := buildConfigs[expectedImportPath]; !exists {
		t.Fatalf("expected build config for import path [%s], got %+v", expectedImportPath, buildConfigs)
	}
}

func TestCreateBuildConfigs(t *testing.T) {
	compare := func(expected string, actual string) {
		if expected != actual {
			t.Errorf("test case failed: expected '%#v', but actual value is '%#v'", expected, actual)
		}
	}

	buildConfigs := []build.Config{
		{ID: "defaults"},
		{ID: "OnlyMain", Main: "test"},
		{ID: "OnlyMainWithFile", Main: "test/main.go"},
		{ID: "OnlyDir", Dir: "test"},
		{ID: "DirAndMain", Dir: "test", Main: "main.go"},
	}

	for _, b := range buildConfigs {
		for importPath, buildCfg := range createBuildConfigs("../..", []build.Config{b}) {
			switch buildCfg.ID {
			case "defaults":
				compare("github.com/google/ko", importPath)

			case "OnlyMain", "OnlyMainWithFile", "OnlyDir", "DirAndMain":
				compare("github.com/google/ko/test", importPath)

			default:
				t.Fatalf("unknown test case: %s", buildCfg.ID)
			}
		}
	}
}
