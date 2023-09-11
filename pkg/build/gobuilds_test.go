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

package build

import (
	"context"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/random"
)

func Test_gobuilds(t *testing.T) {
	base, err := random.Image(1024, 1)
	if err != nil {
		t.Fatalf("random.Image() = %v", err)
	}
	baseRef := name.MustParseReference("all.your/base")
	opts := []Option{
		WithBaseImages(func(context.Context, string) (name.Reference, Result, error) { return baseRef, base, nil }),
		withBuilder(writeTempFile),
		WithPlatforms("all"),
	}

	tests := []struct {
		description       string
		workingDirectory  string
		buildConfigs      map[string]Config
		opts              []Option
		nilDefaultBuilder bool // set to true if you want to test build config and don't want the test to fall back to the default builder
		importpath        string
	}{
		{
			description: "default builder used when no build configs provided",
			opts:        opts,
			importpath:  "github.com/google/ko",
		},
		{
			description:      "match build config using fully qualified import path",
			workingDirectory: "../..",
			buildConfigs: map[string]Config{
				"github.com/google/ko/test": {
					ID:  "build-config-0",
					Dir: "test",
				},
			},
			nilDefaultBuilder: true,
			opts:              opts,
			importpath:        "github.com/google/ko/test",
		},
		{
			description:      "match build config using ko scheme-prefixed fully qualified import path",
			workingDirectory: "../..",
			buildConfigs: map[string]Config{
				"github.com/google/ko/test": {
					ID:  "build-config-1",
					Dir: "test",
				},
			},
			nilDefaultBuilder: true,
			opts:              opts,
			importpath:        "ko://github.com/google/ko/test",
		},
		{
			description:      "find build config by resolving local import path to fully qualified import path",
			workingDirectory: "../../test",
			buildConfigs: map[string]Config{
				"github.com/google/ko/test": {
					ID: "build-config-2",
				},
			},
			nilDefaultBuilder: true,
			opts:              opts,
			importpath:        ".",
		},
		{
			description:      "find build config by matching local import path to build config directory",
			workingDirectory: "../..",
			buildConfigs: map[string]Config{
				"github.com/google/ko/tes12t": {
					ID:  "build-config-3",
					Dir: "test",
				},
			},
			nilDefaultBuilder: true,
			opts:              opts,
			importpath:        "./test",
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			ctx := context.Background()
			bi, err := NewGobuilds(ctx, test.workingDirectory, test.buildConfigs, test.opts...)
			if err != nil {
				t.Fatalf("NewGobuilds(): unexpected error: %v", err)
			}
			gbs := bi.(*gobuilds)
			if test.nilDefaultBuilder {
				gbs.defaultBuilder = nil
			}
			qualifiedImportpath, err := gbs.QualifyImport(test.importpath)
			if err != nil {
				t.Fatalf("gobuilds.QualifyImport(%s): unexpected error: %v", test.importpath, err)
			}
			if err = gbs.IsSupportedReference(qualifiedImportpath); err != nil {
				t.Fatalf("gobuilds.IsSupportedReference(%s): unexpected error: %v", qualifiedImportpath, err)
			}
			result, err := gbs.Build(ctx, qualifiedImportpath)
			if err != nil {
				t.Fatalf("gobuilds.Build(%s): unexpected error = %v", qualifiedImportpath, err)
			}
			if result == nil {
				t.Fatalf("gobuilds.Build(%s): expected non-nil result", qualifiedImportpath)
			}
		})
	}
}

func Test_relativePath(t *testing.T) {
	tests := []struct {
		description string
		baseDir     string
		importpath  string
		want        string
		wantErr     bool
	}{
		{
			description: "all empty string",
		},
		{
			description: "all current directory",
			baseDir:     ".",
			importpath:  ".",
			want:        ".",
		},
		{
			description: "fully qualified import path without ko prefix",
			baseDir:     "also-any-value-because-it-is-ignored",
			importpath:  "example.com/app/cmd/foo",
			want:        "example.com/app/cmd/foo",
		},
		{
			description: "fully qualified import path with ko prefix",
			baseDir:     "also-any-value-because-it-is-ignored",
			importpath:  "ko://example.com/app/cmd/foo",
			want:        "ko://example.com/app/cmd/foo",
		},
		{
			description: "importpath is local subdirectory",
			baseDir:     "foo",
			importpath:  "./foo/bar",
			want:        "./bar",
		},
		{
			description: "importpath is same local directory",
			baseDir:     "foo/bar",
			importpath:  "./foo/bar",
			want:        ".",
		},
		{
			description: "importpath is not subdirectory or same local directory",
			baseDir:     "foo",
			importpath:  "./bar",
			wantErr:     true,
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			got, err := relativePath(test.baseDir, test.importpath)
			if (err != nil) != test.wantErr {
				t.Errorf("relativePath() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if got != test.want {
				t.Errorf("relativePath() got = %v, want %v", got, test.want)
			}
		})
	}
}
