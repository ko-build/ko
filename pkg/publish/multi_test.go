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
	"testing"

	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/ko/pkg/publish"
)

func TestMulti(t *testing.T) {
	img, err := random.Image(1024, 1)
	if err != nil {
		t.Fatalf("random.Image() = %v", err)
	}
	base := "blah"
	repoName := fmt.Sprintf("%s/%s", "example.com", base)
	importpath := "github.com/Google/go-containerregistry/cmd/crane"

	fp, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer fp.Close()
	defer os.Remove(fp.Name())

	tp := publish.NewTarball(fp.Name(), repoName, md5Hash, []string{})

	tmp, err := os.MkdirTemp("/tmp", "ko")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	lp, err := publish.NewLayout(tmp)
	if err != nil {
		t.Errorf("NewLayout() = %v", err)
	}

	p := publish.MultiPublisher(lp, tp)
	if _, err := p.Publish(context.Background(), img, importpath); err != nil {
		t.Errorf("Publish() = %v", err)
	}

	if err := p.Close(); err != nil {
		t.Errorf("Close() = %v", err)
	}
}

func TestMulti_Zero(t *testing.T) {
	img, err := random.Image(1024, 1)
	if err != nil {
		t.Fatalf("random.Image() = %v", err)
	}

	p := publish.MultiPublisher() // No publishers.
	if _, err := p.Publish(context.Background(), img, "foo"); err == nil {
		t.Errorf("Publish() got nil error")
	}
}
