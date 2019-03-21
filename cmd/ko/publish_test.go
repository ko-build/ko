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

package main

import "testing"

func TestQualifyLocalImport(t *testing.T) {
	for _, c := range []struct {
		importpath, gopathsrc, pwd, want string
		wantErr                          bool
	}{{
		importpath: "./cmd/foo",
		gopathsrc:  "/home/go/src",
		pwd:        "/home/go/src/github.com/my/repo",
		want:       "github.com/my/repo/cmd/foo",
	}, {
		importpath: "./foo",
		gopathsrc:  "/home/go/src",
		pwd:        "/home/go/src/github.com/my/repo/cmd",
		want:       "github.com/my/repo/cmd/foo",
	}, {
		// $PWD not on $GOPATH/src
		importpath: "./cmd/foo",
		gopathsrc:  "/home/go/src",
		pwd:        "/",
		wantErr:    true,
	}} {
		got, err := qualifyLocalImport(c.importpath, c.gopathsrc, c.pwd)
		if gotErr := err != nil; gotErr != c.wantErr {
			t.Fatalf("qualifyLocalImport returned %v, wanted err? %t", err, c.wantErr)
		}
		if got != c.want {
			t.Fatalf("qualifyLocalImport(%q, %q, %q): got %q, want %q", c.importpath, c.gopathsrc, c.pwd, got, c.want)
		}
	}
}
