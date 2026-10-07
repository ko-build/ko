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

package publish_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/google/ko/pkg/publish"
)

func TestTarball(t *testing.T) {
	img, err := random.Image(1024, 1)
	if err != nil {
		t.Fatalf("random.Image() = %v", err)
	}
	base := "blah"
	importpath := "github.com/Google/go-containerregistry/cmd/crane"

	fp, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer fp.Close()
	defer os.Remove(fp.Name())

	expectedRepo := md5Hash(base, strings.ToLower(importpath))

	tag, err := name.NewTag(fmt.Sprintf("%s/%s:latest", "example.com", expectedRepo))
	if err != nil {
		t.Fatalf("NewTag() = %v", err)
	}

	repoName := fmt.Sprintf("%s/%s", "example.com", base)
	tagss := [][]string{{
		// no tags
	}, {
		// one tag
		"v0.1.0",
	}, {
		// multiple tags
		"latest",
		"debug",
	}}
	for _, tags := range tagss {
		tp := publish.NewTarball(fp.Name(), repoName, md5Hash, tags)
		if d, err := tp.Publish(context.Background(), img, importpath); err != nil {
			t.Errorf("Publish() = %v", err)
		} else if !strings.HasPrefix(d.String(), tag.Repository.String()) {
			t.Errorf("Publish() = %v, wanted prefix %v", d, tag.Repository)
		}
	}
}

func TestTarballConcurrentPublish(t *testing.T) {
	for _, tags := range [][]string{nil, {"latest"}, {"latest", "debug"}} {
		t.Run(fmt.Sprintf("tags=%v", tags), func(t *testing.T) {
			file := t.TempDir() + "/images.tar"
			base := "example.com/test"
			namer := func(base, importpath string) string { return base + "/" + importpath }
			p := publish.NewTarball(file, base, namer, tags)
			img := empty.Image
			wantDigest, err := img.Digest()
			if err != nil {
				t.Fatal(err)
			}

			const count = 32
			start := make(chan struct{})
			results := make(chan error, count)
			for i := range count {
				go func() {
					<-start
					_, err := p.Publish(context.Background(), img, fmt.Sprintf("app-%d", i))
					results <- err
				}()
			}
			close(start)
			for range count {
				if err := <-results; err != nil {
					t.Errorf("Publish() = %v", err)
				}
			}
			if err := p.Close(); err != nil {
				t.Fatalf("Close() = %v", err)
			}

			if len(tags) == 0 {
				// An untagged archive contains the image without a repository tag.
				stored, err := tarball.ImageFromPath(file, nil)
				if err != nil {
					t.Fatalf("ImageFromPath() = %v", err)
				}
				if got, err := stored.Digest(); err != nil || got != wantDigest {
					t.Errorf("Digest() = %v, %v; want %v", got, err, wantDigest)
				}
				return
			}
			for i := range count {
				for _, tagName := range tags {
					tag, err := name.NewTag(fmt.Sprintf("%s/app-%d:%s", base, i, tagName))
					if err != nil {
						t.Fatal(err)
					}
					stored, err := tarball.ImageFromPath(file, &tag)
					if err != nil {
						t.Errorf("ImageFromPath(%s) = %v", tag, err)
						continue
					}
					if got, err := stored.Digest(); err != nil || got != wantDigest {
						t.Errorf("Digest(%s) = %v, %v; want %v", tag, got, err, wantDigest)
					}
				}
			}
		})
	}
}
