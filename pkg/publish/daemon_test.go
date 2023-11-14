// Copyright 2018 ko Build Authors All Rights Reserved.
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
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/v1/random"
	kotesting "github.com/google/ko/pkg/internal/testing"
	"github.com/google/ko/pkg/publish"
)

func TestDaemon(t *testing.T) {
	importpath := "github.com/google/ko"
	img, err := random.Image(1024, 1)
	if err != nil {
		t.Fatalf("random.Image() = %v", err)
	}

	client := &kotesting.MockDaemon{}
	def, err := publish.NewDaemon(md5Hash, []string{}, publish.WithDockerClient(client))
	if err != nil {
		t.Fatalf("NewDaemon() = %v", err)
	}

	if d, err := def.Publish(context.Background(), img, importpath); err != nil {
		t.Errorf("Publish() = %v", err)
	} else if got, want := d.String(), md5Hash("ko.local", importpath); !strings.HasPrefix(got, want) {
		t.Errorf("Publish() = %v, wanted prefix %v", got, want)
	}
}

func TestDaemonTags(t *testing.T) {
	importpath := "github.com/google/ko"
	img, err := random.Image(1024, 1)
	if err != nil {
		t.Fatalf("random.Image() = %v", err)
	}

	client := &kotesting.MockDaemon{}
	def, err := publish.NewDaemon(md5Hash, []string{"v2.0.0", "v1.2.3", "production"}, publish.WithDockerClient(client))
	if err != nil {
		t.Fatalf("NewDaemon() = %v", err)
	}

	if d, err := def.Publish(context.Background(), img, importpath); err != nil {
		t.Errorf("Publish() = %v", err)
	} else if got, want := d.String(), md5Hash("ko.local", importpath); !strings.HasPrefix(got, want) {
		t.Errorf("Publish() = %v, wanted prefix %v", got, want)
	}

	imgDigest, err := img.Digest()
	if err != nil {
		t.Fatalf("img.Digest() = %v", err)
	}

	expected := []string{fmt.Sprintf("ko.local/98b8c7facdad74510a7cae0cd368eb4e:%s", strings.Replace(imgDigest.String(), "sha256:", "", 1)), "ko.local/98b8c7facdad74510a7cae0cd368eb4e:v2.0.0", "ko.local/98b8c7facdad74510a7cae0cd368eb4e:v1.2.3", "ko.local/98b8c7facdad74510a7cae0cd368eb4e:production"}

	for i, v := range expected {
		if client.Tags[i] != v {
			t.Errorf("Expected tag %v got %v", v, client.Tags[i])
		}
	}
}

func TestDaemonDomain(t *testing.T) {
	importpath := "github.com/google/ko"
	img, err := random.Image(1024, 1)
	if err != nil {
		t.Fatalf("random.Image() = %v", err)
	}

	localDomain := "registry.example.com/repository"
	client := &kotesting.MockDaemon{}
	def, err := publish.NewDaemon(md5Hash, []string{}, publish.WithLocalDomain(localDomain), publish.WithDockerClient(client))
	if err != nil {
		t.Fatalf("NewDaemon() = %v", err)
	}

	if d, err := def.Publish(context.Background(), img, importpath); err != nil {
		t.Errorf("Publish() = %v", err)
	} else if got, want := d.String(), md5Hash(localDomain, importpath); !strings.HasPrefix(got, want) {
		t.Errorf("Publish() = %v, wanted prefix %v", got, want)
	}
}
