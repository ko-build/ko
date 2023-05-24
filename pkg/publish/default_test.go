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
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/ko/pkg/build"
	"github.com/google/ko/pkg/publish"
	ocimutate "github.com/sigstore/cosign/v2/pkg/oci/mutate"
	"github.com/sigstore/cosign/v2/pkg/oci/signed"
	"github.com/sigstore/cosign/v2/pkg/oci/static"
)

var (
	img, _ = random.Image(1024, 3)
	idx, _ = random.Index(1024, 3, 3)
)

func TestDefault(t *testing.T) {
	f, err := static.NewFile([]byte("da bom"))
	if err != nil {
		t.Fatalf("static.NewFile() = %v", err)
	}
	si, err := ocimutate.AttachFileToImage(signed.Image(img), "sbom", f)
	if err != nil {
		t.Fatalf("ocimutate.AttachFileToImage() = %v", err)
	}
	sii, err := ocimutate.AttachFileToImageIndex(signed.ImageIndex(idx), "sbom", f)
	if err != nil {
		t.Fatalf("ocimutate.AttachFileToImageIndex() = %v", err)
	}

	for _, br := range []build.Result{img, idx, si, sii} {
		base := "blah"
		importpath := "github.com/Google/go-containerregistry/cmd/crane"
		expectedRepo := fmt.Sprintf("%s/%s", base, strings.ToLower(importpath))

		server := httptest.NewServer(registry.New())
		defer server.Close()
		u, err := url.Parse(server.URL)
		if err != nil {
			t.Fatalf("url.Parse(%v) = %v", server.URL, err)
		}
		tag, err := name.NewTag(fmt.Sprintf("%s/%s:latest", u.Host, expectedRepo))
		if err != nil {
			t.Fatalf("NewTag() = %v", err)
		}

		repoName := fmt.Sprintf("%s/%s", u.Host, base)
		def, err := publish.NewDefault(repoName)
		if err != nil {
			t.Errorf("NewDefault() = %v", err)
		}
		if d, err := def.Publish(context.Background(), br, build.StrictScheme+importpath); err != nil {
			t.Errorf("Publish() = %v", err)
		} else if !strings.HasPrefix(d.String(), tag.Repository.String()) {
			t.Errorf("Publish() = %v, wanted prefix %v", d, tag.Repository)
		}
	}
}

func md5Hash(base, s string) string {
	// md5 as hex.
	hasher := md5.New()
	hasher.Write([]byte(s))
	return filepath.Join(base, hex.EncodeToString(hasher.Sum(nil)))
}

func TestDefaultWithCustomNamer(t *testing.T) {
	for _, br := range []build.Result{img, idx} {
		base := "blah"
		importpath := "github.com/Google/go-containerregistry/cmd/crane"
		expectedRepo := fmt.Sprintf("%s/%s", base, strings.ToLower(importpath))

		server := httptest.NewServer(registry.New())
		defer server.Close()
		u, err := url.Parse(server.URL)
		if err != nil {
			t.Fatalf("url.Parse(%v) = %v", server.URL, err)
		}
		tag, err := name.NewTag(fmt.Sprintf("%s/%s:latest", u.Host, expectedRepo))
		if err != nil {
			t.Fatalf("NewTag() = %v", err)
		}

		repoName := fmt.Sprintf("%s/%s", u.Host, base)

		def, err := publish.NewDefault(repoName, publish.WithNamer(md5Hash))
		if err != nil {
			t.Errorf("NewDefault() = %v", err)
		}
		if d, err := def.Publish(context.Background(), br, build.StrictScheme+importpath); err != nil {
			t.Errorf("Publish() = %v", err)
		} else if !strings.HasPrefix(d.String(), repoName) {
			t.Errorf("Publish() = %v, wanted prefix %v", d, tag.Repository)
		} else if !strings.HasSuffix(d.Context().String(), md5Hash("", strings.ToLower(importpath))) {
			t.Errorf("Publish() = %v, wanted suffix %v", d.Context(), md5Hash("", importpath))
		}
	}
}

func TestDefaultWithTags(t *testing.T) {
	for _, br := range []build.Result{img, idx} {
		base := "blah"
		importpath := "github.com/Google/go-containerregistry/cmd/crane"
		expectedRepo := fmt.Sprintf("%s/%s", base, strings.ToLower(importpath))

		server := httptest.NewServer(registry.New())
		defer server.Close()
		u, err := url.Parse(server.URL)
		if err != nil {
			t.Fatalf("url.Parse(%v) = %v", server.URL, err)
		}
		tag, err := name.NewTag(fmt.Sprintf("%s/%s:notLatest", u.Host, expectedRepo))
		if err != nil {
			t.Fatalf("NewTag() = %v", err)
		}

		repoName := fmt.Sprintf("%s/%s", u.Host, base)

		def, err := publish.NewDefault(repoName, publish.WithTags([]string{"notLatest", "v1.2.3"}))
		if err != nil {
			t.Errorf("NewDefault() = %v", err)
		}
		if d, err := def.Publish(context.Background(), br, build.StrictScheme+importpath); err != nil {
			t.Errorf("Publish() = %v", err)
		} else if !strings.HasPrefix(d.String(), repoName) {
			t.Errorf("Publish() = %v, wanted prefix %v", d, tag.Repository)
		} else if !strings.HasSuffix(d.Context().String(), strings.ToLower(importpath)) {
			t.Errorf("Publish() = %v, wanted suffix %v", d.Context(), md5Hash("", importpath))
		}

		otherTag := fmt.Sprintf("%s/%s:v1.2.3", u.Host, expectedRepo)

		first, err := crane.Digest(tag.String())
		if err != nil {
			t.Fatal(err)
		}
		second, err := crane.Digest(otherTag)
		if err != nil {
			t.Fatal(err)
		}

		if first != second {
			t.Errorf("tagging didn't work: %s != %s", second, first)
		}
	}
}

func TestDefaultWithReleaseTag(t *testing.T) {
	img, err := random.Image(1024, 1)
	if err != nil {
		t.Fatalf("random.Image() = %v", err)
	}
	base := "blah"
	releaseTag := "v1.2.3"
	importpath := "github.com/Google/go-containerregistry/cmd/crane"
	expectedRepo := fmt.Sprintf("%s/%s", base, strings.ToLower(importpath))

	server := httptest.NewServer(registry.New())
	defer server.Close()

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse(%v) = %v", server.URL, err)
	}
	tag, err := name.NewTag(fmt.Sprintf("%s/%s:notLatest", u.Host, expectedRepo))
	if err != nil {
		t.Fatalf("NewTag() = %v", err)
	}

	repoName := fmt.Sprintf("%s/%s", u.Host, base)

	def, err := publish.NewDefault(repoName, publish.WithTags([]string{releaseTag}))
	if err != nil {
		t.Errorf("NewDefault() = %v", err)
	}
	if d, err := def.Publish(context.Background(), img, build.StrictScheme+importpath); err != nil {
		t.Errorf("Publish() = %v", err)
	} else if !strings.HasPrefix(d.String(), repoName) {
		t.Errorf("Publish() = %v, wanted prefix %v", d, tag.Repository)
	} else if !strings.HasSuffix(d.Context().String(), strings.ToLower(importpath)) {
		t.Errorf("Publish() = %v, wanted suffix %v", d.Context(), md5Hash("", importpath))
	} else if !strings.Contains(d.String(), releaseTag) {
		t.Errorf("Publish() = %v, wanted tag included: %v", d.String(), releaseTag)
	}

	tags, err := remote.List(tag.Context())
	if err != nil {
		t.Fatalf("remote.List(): %v", err)
	}
	createdTags := make(map[string]struct{})
	for _, got := range tags {
		createdTags[got] = struct{}{}
	}
	if _, ok := createdTags["v1.2.3"]; !ok {
		t.Errorf("Tag v1.2.3 was not created.")
	}

	def, err = publish.NewDefault(repoName, publish.WithTags([]string{releaseTag}), publish.WithTagOnly(true))
	if err != nil {
		t.Errorf("NewDefault() = %v", err)
	}
	if d, err := def.Publish(context.Background(), img, build.StrictScheme+importpath); err != nil {
		t.Errorf("Publish() = %v", err)
	} else if !strings.HasPrefix(d.String(), repoName) {
		t.Errorf("Publish() = %v, wanted prefix %v", d, tag.Repository)
	} else if !strings.HasSuffix(d.Context().String(), strings.ToLower(importpath)) {
		t.Errorf("Publish() = %v, wanted suffix %v", d.Context(), md5Hash("", importpath))
	} else if !strings.Contains(d.String(), releaseTag) {
		t.Errorf("Publish() = %v, wanted tag included: %v", d.String(), releaseTag)
	} else if strings.Contains(d.String(), "@sha256:") {
		t.Errorf("Publish() = %v, wanted no digest", d.String())
	}
}
