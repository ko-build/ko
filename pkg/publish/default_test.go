// Copyright 2018 Google LLC All Rights Reserved.
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
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/random"
)

func TestDefault(t *testing.T) {
	img, err := random.Image(1024, 1)
	if err != nil {
		t.Fatalf("random.Image() = %v", err)
	}
	base := "blah"
	importpath := "github.com/Google/go-containerregistry/cmd/crane"
	expectedRepo := fmt.Sprintf("%s/%s", base, strings.ToLower(importpath))
	headPathPrefix := fmt.Sprintf("/v2/%s/blobs/", expectedRepo)
	initiatePath := fmt.Sprintf("/v2/%s/blobs/uploads/", expectedRepo)
	manifestPath := fmt.Sprintf("/v2/%s/manifests/latest", expectedRepo)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead && strings.HasPrefix(r.URL.Path, headPathPrefix) && r.URL.Path != initiatePath {
			http.Error(w, "NotFound", http.StatusNotFound)
			return
		}
		switch r.URL.Path {
		case "/v2/":
			w.WriteHeader(http.StatusOK)
		case initiatePath:
			if r.Method != http.MethodPost {
				t.Errorf("Method; got %v, want %v", r.Method, http.MethodPost)
			}
			http.Error(w, "Mounted", http.StatusCreated)
		case manifestPath:
			if r.Method != http.MethodPut {
				t.Errorf("Method; got %v, want %v", r.Method, http.MethodPut)
			}
			http.Error(w, "Created", http.StatusCreated)
		default:
			t.Fatalf("Unexpected path: %v", r.URL.Path)
		}
	}))
	defer server.Close()
	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse(%v) = %v", server.URL, err)
	}
	tag, err := name.NewTag(fmt.Sprintf("%s/%s:latest", u.Host, expectedRepo), name.WeakValidation)
	if err != nil {
		t.Fatalf("NewTag() = %v", err)
	}

	repoName := fmt.Sprintf("%s/%s", u.Host, base)
	def, err := NewDefault(repoName)
	if err != nil {
		t.Errorf("NewDefault() = %v", err)
	}
	if d, err := def.Publish(img, importpath); err != nil {
		t.Errorf("Publish() = %v", err)
	} else if !strings.HasPrefix(d.String(), tag.Repository.String()) {
		t.Errorf("Publish() = %v, wanted prefix %v", d, tag.Repository)
	}
}

func md5Hash(s string) string {
	// md5 as hex.
	hasher := md5.New()
	hasher.Write([]byte(s))
	return hex.EncodeToString(hasher.Sum(nil))
}

func TestDefaultWithCustomNamer(t *testing.T) {
	img, err := random.Image(1024, 1)
	if err != nil {
		t.Fatalf("random.Image() = %v", err)
	}
	base := "blah"
	importpath := "github.com/Google/go-containerregistry/cmd/crane"
	expectedRepo := fmt.Sprintf("%s/%s", base, md5Hash(strings.ToLower(importpath)))
	headPathPrefix := fmt.Sprintf("/v2/%s/blobs/", expectedRepo)
	initiatePath := fmt.Sprintf("/v2/%s/blobs/uploads/", expectedRepo)
	manifestPath := fmt.Sprintf("/v2/%s/manifests/latest", expectedRepo)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead && strings.HasPrefix(r.URL.Path, headPathPrefix) && r.URL.Path != initiatePath {
			http.Error(w, "NotFound", http.StatusNotFound)
			return
		}
		switch r.URL.Path {
		case "/v2/":
			w.WriteHeader(http.StatusOK)
		case initiatePath:
			if r.Method != http.MethodPost {
				t.Errorf("Method; got %v, want %v", r.Method, http.MethodPost)
			}
			http.Error(w, "Mounted", http.StatusCreated)
		case manifestPath:
			if r.Method != http.MethodPut {
				t.Errorf("Method; got %v, want %v", r.Method, http.MethodPut)
			}
			http.Error(w, "Created", http.StatusCreated)
		default:
			t.Fatalf("Unexpected path: %v", r.URL.Path)
		}
	}))
	defer server.Close()
	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse(%v) = %v", server.URL, err)
	}
	tag, err := name.NewTag(fmt.Sprintf("%s/%s:latest", u.Host, expectedRepo), name.WeakValidation)
	if err != nil {
		t.Fatalf("NewTag() = %v", err)
	}

	repoName := fmt.Sprintf("%s/%s", u.Host, base)

	def, err := NewDefault(repoName, WithNamer(md5Hash))
	if err != nil {
		t.Errorf("NewDefault() = %v", err)
	}
	if d, err := def.Publish(img, importpath); err != nil {
		t.Errorf("Publish() = %v", err)
	} else if !strings.HasPrefix(d.String(), repoName) {
		t.Errorf("Publish() = %v, wanted prefix %v", d, tag.Repository)
	} else if !strings.HasSuffix(d.Context().String(), md5Hash(strings.ToLower(importpath))) {
		t.Errorf("Publish() = %v, wanted suffix %v", d.Context(), md5Hash(importpath))
	}
}

func TestDefaultWithTags(t *testing.T) {
	img, err := random.Image(1024, 1)
	if err != nil {
		t.Fatalf("random.Image() = %v", err)
	}
	base := "blah"
	importpath := "github.com/Google/go-containerregistry/cmd/crane"
	expectedRepo := fmt.Sprintf("%s/%s", base, strings.ToLower(importpath))
	headPathPrefix := fmt.Sprintf("/v2/%s/blobs/", expectedRepo)
	initiatePath := fmt.Sprintf("/v2/%s/blobs/uploads/", expectedRepo)
	manifestPath := fmt.Sprintf("/v2/%s/manifests/", expectedRepo)

	createdTags := make(map[string]struct{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead && strings.HasPrefix(r.URL.Path, headPathPrefix) && r.URL.Path != initiatePath {
			http.Error(w, "NotFound", http.StatusNotFound)
			return
		}
		switch {
		case r.URL.Path == "/v2/":
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == initiatePath:
			if r.Method != http.MethodPost {
				t.Errorf("Method; got %v, want %v", r.Method, http.MethodPost)
			}
			http.Error(w, "Mounted", http.StatusCreated)
		case strings.HasPrefix(r.URL.Path, manifestPath):
			if r.Method != http.MethodPut {
				t.Errorf("Method; got %v, want %v", r.Method, http.MethodPut)
			}

			createdTags[strings.TrimPrefix(r.URL.Path, manifestPath)] = struct{}{}

			http.Error(w, "Created", http.StatusCreated)
		default:
			t.Fatalf("Unexpected path: %v", r.URL.Path)
		}
	}))
	defer server.Close()
	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse(%v) = %v", server.URL, err)
	}
	tag, err := name.NewTag(fmt.Sprintf("%s/%s:notLatest", u.Host, expectedRepo), name.WeakValidation)
	if err != nil {
		t.Fatalf("NewTag() = %v", err)
	}

	repoName := fmt.Sprintf("%s/%s", u.Host, base)

	def, err := NewDefault(repoName, WithTags([]string{"notLatest", "v1.2.3"}))
	if err != nil {
		t.Errorf("NewDefault() = %v", err)
	}
	if d, err := def.Publish(img, importpath); err != nil {
		t.Errorf("Publish() = %v", err)
	} else if !strings.HasPrefix(d.String(), repoName) {
		t.Errorf("Publish() = %v, wanted prefix %v", d, tag.Repository)
	} else if !strings.HasSuffix(d.Context().String(), strings.ToLower(importpath)) {
		t.Errorf("Publish() = %v, wanted suffix %v", d.Context(), md5Hash(importpath))
	}

	if _, ok := createdTags["notLatest"]; !ok {
		t.Errorf("Tag notLatest was not created.")
	}

	if _, ok := createdTags["v1.2.3"]; !ok {
		t.Errorf("Tag v1.2.3 was not created.")
	}
}
