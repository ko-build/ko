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
	"github.com/google/go-containerregistry/pkg/v1/random"
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
