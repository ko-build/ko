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
	"errors"
	"fmt"
	"io"
	"log"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/google/go-containerregistry/pkg/v1/random"

	"github.com/google/ko/pkg/commands/options"
)

func TestOverrideDefaultBaseImageUsingBuildOption(t *testing.T) {
	namespace := "base"
	s, err := registryServerWithImage(namespace)
	if err != nil {
		t.Fatalf("could not create test registry server: %v", err)
	}
	defer s.Close()
	baseImage := fmt.Sprintf("%s/%s", s.Listener.Addr().String(), namespace)
	wantDigest, err := crane.Digest(baseImage)
	if err != nil {
		t.Fatalf("crane.Digest(%s): %v", baseImage, err)
	}
	wantImage := fmt.Sprintf("%s@%s", baseImage, wantDigest)
	bo := &options.BuildOptions{
		BaseImage: wantImage,
		Platforms: []string{"all"},
	}

	baseFn := getBaseImage(bo)
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

func TestGetBaseImageLocalHonorsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	bo := &options.BuildOptions{
		BaseImage: "ko.local/does-not-exist",
	}
	_, _, err := getBaseImage(bo)(ctx, "ko://example.com/helloworld")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("getBaseImage() = %v, want context.Canceled", err)
	}
}

func TestGetBaseImageConcurrentFetch(t *testing.T) {
	t.Setenv("KOCACHE", "")

	s := httptest.NewServer(registry.New(registry.Logger(log.New(io.Discard, "", 0))))
	defer s.Close()

	const n = 8
	bo := &options.BuildOptions{BaseImageOverrides: map[string]string{}}
	want := map[string]string{}
	for i := range n {
		img, err := random.Image(1024, 1)
		if err != nil {
			t.Fatalf("random.Image(): %v", err)
		}
		ref := fmt.Sprintf("%s/base%d", s.Listener.Addr().String(), i)
		if err := crane.Push(img, ref); err != nil {
			t.Fatalf("crane.Push(%s): %v", ref, err)
		}
		importpath := fmt.Sprintf("example.com/cmd%d", i)
		bo.BaseImageOverrides[importpath] = ref
		want[importpath] = mustDigest(img).String()
	}

	baseFn := getBaseImage(bo)
	var wg sync.WaitGroup
	for importpath, wantDigest := range want {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, res, err := baseFn(context.Background(), "ko://"+importpath)
			if err != nil {
				t.Errorf("getBaseImage(%s): %v", importpath, err)
				return
			}
			digest, err := res.Digest()
			if err != nil {
				t.Errorf("res.Digest(): %v", err)
				return
			}
			if got := digest.String(); got != wantDigest {
				t.Errorf("getBaseImage(%s) digest = %s, wanted %s", importpath, got, wantDigest)
			}
		}()
	}
	wg.Wait()
}
