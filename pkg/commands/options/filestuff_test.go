// Copyright 2026 ko Build Authors All Rights Reserved.
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

package options

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestEnumerateFilesRecursiveFollowsSymlinkedDirectory(t *testing.T) {
	root := t.TempDir()
	target := t.TempDir()

	if err := os.WriteFile(filepath.Join(target, "deployment.yaml"), []byte("kind: Deployment\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "ignored.txt"), []byte("ignored\n"), 0644); err != nil {
		t.Fatal(err)
	}

	link := filepath.Join(root, "linked")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("skipping symlink test: %v", err)
	}

	got := collectFiles(&FilenameOptions{Filenames: []string{root}, Recursive: true})
	want := []string{filepath.Join(link, "deployment.yaml")}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("EnumerateFiles() = %v, want %v", got, want)
	}
}

func TestEnumerateFilesRecursiveFollowsSymlinkedRootDirectory(t *testing.T) {
	target := t.TempDir()
	link := filepath.Join(t.TempDir(), "linked")

	if err := os.WriteFile(filepath.Join(target, "service.yaml"), []byte("kind: Service\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("skipping symlink test: %v", err)
	}

	got := collectFiles(&FilenameOptions{Filenames: []string{link}, Recursive: true})
	want := []string{filepath.Join(link, "service.yaml")}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("EnumerateFiles() = %v, want %v", got, want)
	}
}

func TestEnumerateFilesNonRecursiveSkipsSymlinkedSubdirectory(t *testing.T) {
	root := t.TempDir()
	target := t.TempDir()

	if err := os.WriteFile(filepath.Join(target, "deployment.yaml"), []byte("kind: Deployment\n"), 0644); err != nil {
		t.Fatal(err)
	}

	link := filepath.Join(root, "linked")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("skipping symlink test: %v", err)
	}

	got := collectFiles(&FilenameOptions{Filenames: []string{root}})
	if len(got) != 0 {
		t.Fatalf("EnumerateFiles() = %v, want no files", got)
	}
}

func collectFiles(fo *FilenameOptions) []string {
	got := make([]string, 0)
	for file := range EnumerateFiles(fo) {
		got = append(got, file)
	}
	sort.Strings(got)
	return got
}
