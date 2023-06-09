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
	"testing"

	"github.com/google/go-containerregistry/pkg/crane"

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
