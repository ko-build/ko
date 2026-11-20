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

package publish

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

func TestPodmanPublish(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake Podman requires /bin/sh")
	}
	dir := t.TempDir()
	script := `#!/bin/sh
set -eu
test "$#" -eq 4
test "$1 $2 $3" = "load --quiet --input"
printf '%s' "$4" > "$ARCHIVE_PATH"
if [ "$FAIL_LOAD" = yes ]; then
    echo "storage unavailable" >&2
    exit 1
fi
/bin/cp "$4" "$CAPTURE"
`
	if err := os.WriteFile(filepath.Join(dir, "podman"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	capture := filepath.Join(dir, "captured.tar")
	pathFile := filepath.Join(dir, "archive-path")
	t.Setenv("CAPTURE", capture)
	t.Setenv("ARCHIVE_PATH", pathFile)

	img, err := random.Image(1024, 1)
	if err != nil {
		t.Fatal(err)
	}
	h, err := img.Digest()
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := img.ConfigName()
	if err != nil {
		t.Fatal(err)
	}
	namer := func(base, s string) string { return base + "/" + s }
	base, imagePath := "ko.local", "example.com/APP"
	repo := "ko.local/example.com/app"

	for _, tt := range []struct {
		name    string
		tags    []string
		tagOnly bool
		wantTag string
		fail    bool
	}{
		{name: "digest tag", tags: []string{"latest", "dev"}, wantTag: h.Hex},
		{name: "explicit tag only", tags: []string{"dev"}, tagOnly: true, wantTag: "dev"},
		{name: "no user tags", wantTag: h.Hex},
		{name: "load error", tags: []string{"latest"}, fail: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if tt.fail {
				t.Setenv("FAIL_LOAD", "yes")
			} else {
				t.Setenv("FAIL_LOAD", "no")
			}
			p, err := NewPodman(base, namer, tt.tags, tt.tagOnly)
			if err != nil {
				t.Fatal(err)
			}
			ref, err := p.Publish(t.Context(), img, "ko://"+imagePath)
			if tt.fail {
				if err == nil || !strings.Contains(err.Error(), "storage unavailable") {
					t.Errorf("Publish() error = %v, want the Podman error", err)
				}
			} else if err != nil {
				t.Fatalf("Publish() = %v", err)
			} else if got, want := ref.Name(), repo+":"+tt.wantTag; got != want {
				t.Errorf("Publish() = %s, want %s", got, want)
			}
			archivePath, err := os.ReadFile(pathFile)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(string(archivePath)); !errors.Is(err, os.ErrNotExist) {
				t.Errorf("os.Stat(%q) = %v, want archive removed", archivePath, err)
			}
			if tt.fail {
				return
			}
			manifest, err := tarball.LoadManifest(func() (io.ReadCloser, error) {
				return os.Open(capture)
			})
			if err != nil {
				t.Fatal(err)
			}
			wantTags := []string{repo + ":" + h.Hex}
			for _, tag := range tt.tags {
				wantTags = append(wantTags, repo+":"+tag)
			}
			if len(manifest) != 1 {
				t.Fatalf("archive contains %d images, want one", len(manifest))
			}
			desc := manifest[0]
			slices.Sort(desc.RepoTags)
			slices.Sort(wantTags)
			if !slices.Equal(desc.RepoTags, wantTags) || desc.Config != cfg.String() || len(desc.Layers) != 1 {
				t.Errorf("archive contains %+v, want tags %v, config %s, and one layer", desc, wantTags, cfg)
			}
		})
	}
}

func TestPodmanRejectsIndex(t *testing.T) {
	p := &podman{}
	_, err := p.Publish(t.Context(), empty.Index, "ko://example.com/app")
	if err == nil {
		t.Fatal("Publish(index) succeeded, want single-platform error")
	}
	for _, hint := range []string{"single-platform", "--platform", "defaultPlatforms", "KO_DEFAULTPLATFORMS"} {
		if !strings.Contains(err.Error(), hint) {
			t.Errorf("Publish(index) error %q does not mention %s", err, hint)
		}
	}
}

func TestPodmanRejectsInvalidTagOnly(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	for _, tt := range []struct {
		name string
		tags []string
		want string
	}{
		{name: "none", want: "must specify exactly one tag"},
		{name: "empty", tags: []string{""}, want: "must specify exactly one tag"},
		{name: "latest", tags: []string{"latest"}, want: "latest tag cannot be used"},
		{name: "multiple", tags: []string{"dev", "test"}, want: "must specify exactly one tag"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewPodman("ko.local", identity, tt.tags, true); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("NewPodman(%v) = %v, want %q", tt.tags, err, tt.want)
			}
		})
	}
}

func TestPodmanCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	p := &podman{}
	if _, err := p.Publish(ctx, empty.Image, "ko://example.com/app"); !errors.Is(err, context.Canceled) {
		t.Errorf("Publish() = %v, want context.Canceled", err)
	}
}
