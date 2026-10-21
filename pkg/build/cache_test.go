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

package build

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/google/go-containerregistry/pkg/v1/random"
)

// copyBinary copies a Go binary (so it has a build ID) into dir.
func copyBinary(t *testing.T, src, dir string) string {
	t.Helper()
	in, err := os.Open(src)
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	dst := filepath.Join(dir, "app")
	out, err := os.Create(dst)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		t.Fatal(err)
	}
	return dst
}

func TestLayerCacheConcurrentPutAndGetMeta(t *testing.T) {
	t.Setenv("KOCACHE", t.TempDir())

	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	const n = 8
	files := make([]string, n)
	for i := range files {
		dir := t.TempDir()
		files[i] = copyBinary(t, self, dir)
		// Seed empty metadata so getMeta populates the in-memory maps.
		for _, name := range []string{"buildid-to-diffid", "diffid-to-descriptor"} {
			if err := os.WriteFile(filepath.Join(dir, name), []byte("{}"), 0644); err != nil {
				t.Fatal(err)
			}
		}
	}

	layer, err := random.Layer(64, "application/octet-stream")
	if err != nil {
		t.Fatal(err)
	}

	c := &layerCache{
		buildToDiff: map[string]buildIDToDiffID{},
		diffToDesc:  map[string]diffIDToDescriptor{},
	}
	ctx := context.Background()

	var wg sync.WaitGroup
	for _, file := range files {
		wg.Add(2)
		go func() {
			defer wg.Done()
			// The seeded metadata is empty, so this is always a miss.
			_, _, _ = c.getMeta(ctx, file)
		}()
		go func() {
			defer wg.Done()
			if err := c.put(ctx, file, layer); err != nil {
				t.Errorf("put(%q) = %v", file, err)
			}
		}()
	}
	wg.Wait()
}
