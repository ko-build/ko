// Copyright 2020 ko Build Authors All Rights Reserved.
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

package publish

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/ko/pkg/build"
)

func TestLayout(t *testing.T) {
	img, err := random.Image(1024, 1)
	if err != nil {
		t.Fatalf("random.Image() = %v", err)
	}
	importpath := "github.com/Google/go-containerregistry/cmd/crane"

	tmp, err := os.MkdirTemp("/tmp", "ko")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	lp := NewLayout(tmp)
	if d, err := lp.Publish(context.Background(), img, importpath); err != nil {
		t.Errorf("Publish() = %v", err)
	} else if !strings.HasPrefix(d.String(), tmp) {
		t.Errorf("Publish() = %v, wanted prefix %v", d, tmp)
	}
}

func TestLayoutConcurrentPublish(t *testing.T) {
	for _, existing := range []bool{false, true} {
		t.Run(fmt.Sprintf("existing=%t", existing), func(t *testing.T) {
			// The layout path is also used as an image reference, so it must
			// not contain the uppercase characters in macOS's default temp path.
			dir, err := os.MkdirTemp("/tmp", "ko-layout")
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { os.RemoveAll(dir) })
			want := make(map[v1.Hash]bool)
			if existing {
				p, err := layout.Write(dir, empty.Index)
				if err != nil {
					t.Fatal(err)
				}
				if err := p.AppendImage(empty.Image); err != nil {
					t.Fatal(err)
				}
				h, err := empty.Image.Digest()
				if err != nil {
					t.Fatal(err)
				}
				want[h] = false
			}

			const count = 32
			images := make([]build.Result, count)
			for i := range count {
				var br build.Result
				var err error
				isIndex := i%2 == 0
				if isIndex {
					br, err = random.Index(1024, 1, 2)
				} else {
					br, err = random.Image(1024, 1)
				}
				if err != nil {
					t.Fatal(err)
				}
				images[i] = br
				h, err := br.Digest()
				if err != nil {
					t.Fatal(err)
				}
				want[h] = isIndex
			}

			p := NewLayout(dir)
			start := make(chan struct{})
			results := make(chan error, count)
			for i, br := range images {
				go func() {
					<-start
					_, err := p.Publish(context.Background(), br, fmt.Sprintf("app-%d", i))
					results <- err
				}()
			}
			close(start)
			for range count {
				if err := <-results; err != nil {
					t.Errorf("Publish() = %v", err)
				}
			}

			stored, err := layout.ImageIndexFromPath(dir)
			if err != nil {
				t.Fatal(err)
			}
			manifest, err := stored.IndexManifest()
			if err != nil {
				t.Fatal(err)
			}
			if len(manifest.Manifests) != len(want) {
				t.Errorf("layout contains %d descriptors, want %d", len(manifest.Manifests), len(want))
			}
			for h, isIndex := range want {
				if isIndex {
					if _, err := stored.ImageIndex(h); err != nil {
						t.Errorf("ImageIndex(%s) = %v", h, err)
					}
				} else if _, err := stored.Image(h); err != nil {
					t.Errorf("Image(%s) = %v", h, err)
				}
			}
		})
	}
}
