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

package options_test

import (
	"path"
	"testing"

	"github.com/google/ko/pkg/commands/options"
)

func TestMakeNamer(t *testing.T) {
	foreachTestCaseMakeNamer(func(tc testMakeNamerCase) {
		t.Run(tc.name, func(t *testing.T) {
			namer := options.MakeNamer(&tc.opts)
			got := namer("registry.example.org/foo/bar", "example.org/sample/cmd/example")

			if got != tc.want {
				t.Errorf("got image name %s, wanted %s", got, tc.want)
			}
		})
	})
}

func foreachTestCaseMakeNamer(fn func(tc testMakeNamerCase)) {
	for _, namerCase := range testMakeNamerCases() {
		fn(namerCase)
	}
}

func testMakeNamerCases() []testMakeNamerCase {
	return []testMakeNamerCase{{
		name: "defaults",
		want: "registry.example.org/foo/bar/example-51d74b7127c5f7495a338df33ecdeb19",
	}, {
		name: "with preserve import paths",
		want: "registry.example.org/foo/bar/example.org/sample/cmd/example",
		opts: options.PublishOptions{PreserveImportPaths: true},
	}, {
		name: "with base import paths",
		want: "registry.example.org/foo/bar/example",
		opts: options.PublishOptions{BaseImportPaths: true},
	}, {
		name: "with bare",
		want: "registry.example.org/foo/bar",
		opts: options.PublishOptions{Bare: true},
	}, {
		name: "with custom namer",
		want: "registry.example.org/foo/bar-example",
		opts: options.PublishOptions{ImageNamer: func(base string, importpath string) string {
			return base + "-" + path.Base(importpath)
		}},
	}}
}

type testMakeNamerCase struct {
	name string
	opts options.PublishOptions
	want string
}
