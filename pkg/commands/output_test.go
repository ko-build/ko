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

package commands

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/ko/pkg/build"
	"github.com/google/ko/pkg/commands/options"
	"github.com/google/ko/pkg/publish"
)

const (
	digestA = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	hexB    = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	digestB = "sha256:" + hexB
)

func TestNewBuildOutput(t *testing.T) {
	pushed := &options.PublishOptions{Push: true, DockerRepo: "example.com", Tags: []string{"latest"}}
	for _, tc := range []struct {
		name    string
		results []publish.Published
		po      *options.PublishOptions
		bo      *options.BuildOptions
		want    buildOutput
	}{{
		name: "empty",
		po:   pushed,
		bo:   &options.BuildOptions{},
		want: buildOutput{Tags: []string{}, Labels: map[string]string{}, Annotations: []string{}, Images: []imageOutput{}},
	}, {
		name: "digest-only refs get every requested tag; labels and annotations",
		results: []publish.Published{
			{ImportPath: "example.com/b", Ref: name.MustParseReference("example.com/b@" + digestB), Digest: digestB, Platforms: []string{"linux/arm64"}},
			{ImportPath: "example.com/a", Ref: name.MustParseReference("localhost:5000/a@" + digestA), Digest: digestA, Platforms: []string{}},
		},
		po: &options.PublishOptions{Push: true, Tags: []string{"latest", "v1"}},
		bo: &options.BuildOptions{Labels: []string{"org.opencontainers.image.title=app", "empty="}, Annotations: []string{"a=b"}},
		want: buildOutput{
			Tags:        []string{"example.com/b:latest", "example.com/b:v1", "localhost:5000/a:latest", "localhost:5000/a:v1"},
			Labels:      map[string]string{"org.opencontainers.image.title": "app", "empty": ""},
			Annotations: []string{"a=b"},
			Images: []imageOutput{
				{ImportPath: "example.com/b", Ref: "example.com/b@" + digestB, Digest: digestB, Platforms: []string{"linux/arm64"}},
				{ImportPath: "example.com/a", Ref: "localhost:5000/a@" + digestA, Digest: digestA, Platforms: []string{}},
			},
		},
	}, {
		name: "tags drop the digest and the daemon's digest tag",
		results: []publish.Published{
			{ImportPath: "example.com/c", Ref: name.MustParseReference("example.com/c:v1@" + digestA), Digest: digestA},
			{ImportPath: "example.com/d", Ref: name.MustParseReference("ko.local/d:" + hexB), Digest: digestB},
		},
		po: &options.PublishOptions{Tags: []string{"v1"}},
		bo: &options.BuildOptions{},
		want: buildOutput{
			Tags: []string{"example.com/c:v1", "ko.local/d:v1"}, Labels: map[string]string{}, Annotations: []string{},
			Images: []imageOutput{
				{ImportPath: "example.com/c", Ref: "example.com/c:v1@" + digestA, Digest: digestA},
				{ImportPath: "example.com/d", Ref: "ko.local/d:" + hexB, Digest: digestB},
			},
		},
	}, {
		name: "images sharing a name are not deduplicated",
		results: []publish.Published{
			{ImportPath: "example.com/f", Ref: name.MustParseReference("example.com/shared@" + digestA), Digest: digestA},
			{ImportPath: "example.com/g", Ref: name.MustParseReference("example.com/shared@" + digestB), Digest: digestB},
		},
		po: pushed,
		bo: &options.BuildOptions{},
		want: buildOutput{
			Tags: []string{"example.com/shared:latest", "example.com/shared:latest"}, Labels: map[string]string{}, Annotations: []string{},
			Images: []imageOutput{
				{ImportPath: "example.com/f", Ref: "example.com/shared@" + digestA, Digest: digestA},
				{ImportPath: "example.com/g", Ref: "example.com/shared@" + digestB, Digest: digestB},
			},
		},
	}, {
		name:    "SBOMs get a ref when pushed",
		results: []publish.Published{withSBOMs()},
		po:      pushed,
		bo:      &options.BuildOptions{},
		want: buildOutput{
			Tags: []string{"example.com/a:latest"}, Labels: map[string]string{}, Annotations: []string{},
			Images: []imageOutput{{
				ImportPath: "example.com/a", Ref: "example.com/a@" + digestA, Digest: digestA, Platforms: []string{"linux/amd64"},
				SBOM: &sbomsOutput{
					Index:     &sbomOutput{Format: "spdx", MediaType: "text/spdx+json", Digest: digestB, Size: 10, Ref: "example.com/a:sha256-" + hexB + ".sbom"},
					Platforms: []sbomOutput{{Platform: "linux/amd64", Format: "spdx", MediaType: "text/spdx+json", Digest: digestB, Size: 20, Ref: "example.com/a:sha256-" + hexB + ".sbom"}},
				},
			}},
		},
	}, {
		name:    "SBOMs get no ref when not pushed",
		results: []publish.Published{withSBOMs()},
		po:      &options.PublishOptions{Push: false, Tags: []string{"latest"}},
		bo:      &options.BuildOptions{},
		want: buildOutput{
			Tags: []string{"example.com/a:latest"}, Labels: map[string]string{}, Annotations: []string{},
			Images: []imageOutput{{
				ImportPath: "example.com/a", Ref: "example.com/a@" + digestA, Digest: digestA, Platforms: []string{"linux/amd64"},
				SBOM: &sbomsOutput{
					Index:     &sbomOutput{Format: "spdx", MediaType: "text/spdx+json", Digest: digestB, Size: 10},
					Platforms: []sbomOutput{{Platform: "linux/amd64", Format: "spdx", MediaType: "text/spdx+json", Digest: digestB, Size: 20}},
				},
			}},
		},
	}} {
		t.Run(tc.name, func(t *testing.T) {
			tc.want.Version = version()
			got, err := newBuildOutput(tc.results, tc.po, tc.bo)
			if err != nil {
				t.Fatalf("newBuildOutput() = %v", err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("newBuildOutput() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func withSBOMs() publish.Published {
	subject := name.MustParseReference("example.com/a@" + digestB).(name.Digest)
	return publish.Published{
		ImportPath: "example.com/a", Ref: name.MustParseReference("example.com/a@" + digestA), Digest: digestA, Platforms: []string{"linux/amd64"},
		SBOMs: publish.SBOMs{
			Index:     &publish.SBOM{MediaType: "text/spdx+json", Digest: digestB, Size: 10, Subject: subject},
			Platforms: []publish.SBOM{{Platform: &v1.Platform{OS: "linux", Architecture: "amd64"}, MediaType: "text/spdx+json", Digest: digestB, Size: 20, Subject: subject}},
		},
	}
}

func TestSBOMOutputPath(t *testing.T) {
	const importpath = "example.com/app/cmd/server"
	amd64 := &v1.Platform{OS: "linux", Architecture: "amd64"}
	arm64v8 := &v1.Platform{OS: "linux", Architecture: "arm64", Variant: "v8"}
	content := []byte("sbom")
	digest := fmt.Sprintf("sha256:%x", sha256.Sum256(content))

	for _, tc := range []struct {
		name     string
		written  *v1.Platform // the platform the file is named for; nil means the index
		sbom     publish.SBOM
		noDir    bool
		wantPath bool
	}{
		{name: "platform image", written: amd64, sbom: publish.SBOM{Platform: amd64, Digest: digest}, wantPath: true},
		{name: "index", written: nil, sbom: publish.SBOM{Digest: digest}, wantPath: true},
		{name: "file named without the variant", written: &v1.Platform{OS: "linux", Architecture: "arm64"}, sbom: publish.SBOM{Platform: arm64v8, Digest: digest}, wantPath: true},
		{name: "file named with the variant", written: arm64v8, sbom: publish.SBOM{Platform: &v1.Platform{OS: "linux", Architecture: "arm64"}, Digest: digest}, wantPath: true},
		{name: "file named with another architecture's variant", written: &v1.Platform{OS: "linux", Architecture: "arm", Variant: "v7"}, sbom: publish.SBOM{Platform: &v1.Platform{OS: "linux", Architecture: "arm64"}, Digest: digest}},
		{name: "file overwritten by another image", written: amd64, sbom: publish.SBOM{Platform: amd64, Digest: digestB}},
		{name: "no such file", written: amd64, sbom: publish.SBOM{Platform: arm64v8, Digest: digest}},
		{name: "no SBOM directory", written: amd64, sbom: publish.SBOM{Platform: amd64, Digest: digest}, noDir: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			want := build.SBOMPath(dir, importpath, tc.written)
			if err := os.WriteFile(want, content, 0o600); err != nil {
				t.Fatal(err)
			}
			bo := &options.BuildOptions{SBOMDir: dir}
			if tc.noDir {
				bo.SBOMDir = ""
			}
			res := publish.Published{ImportPath: importpath, Ref: name.MustParseReference("example.com/server@" + digestA)}
			if tc.sbom.Platform == nil {
				res.SBOMs.Index = &tc.sbom
			} else {
				res.SBOMs.Platforms = []publish.SBOM{tc.sbom}
			}

			out, err := newBuildOutput([]publish.Published{res}, &options.PublishOptions{}, bo)
			if err != nil {
				t.Fatalf("newBuildOutput() = %v", err)
			}
			sb := out.Images[0].SBOM.Index
			if tc.sbom.Platform != nil {
				sb = &out.Images[0].SBOM.Platforms[0]
			}
			if !tc.wantPath {
				want = ""
			}
			if sb.Path != want {
				t.Errorf("path = %q, want %q", sb.Path, want)
			}
		})
	}
}

func TestNewBuildOutputCosignRepository(t *testing.T) {
	for _, tc := range []struct {
		name, cosignRepository string
		push                   bool
		wantRef                string
		wantErr                bool
	}{
		{name: "pushed, SBOMs in COSIGN_REPOSITORY", cosignRepository: "sboms.example.com/store", push: true, wantRef: "sboms.example.com/store:sha256-" + hexB + ".sbom"},
		{name: "not pushed, invalid COSIGN_REPOSITORY is ignored", cosignRepository: "not a valid repository!"},
		{name: "pushed, invalid COSIGN_REPOSITORY fails", cosignRepository: "not a valid repository!", push: true, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("COSIGN_REPOSITORY", tc.cosignRepository)
			out, err := newBuildOutput([]publish.Published{withSBOMs()}, &options.PublishOptions{Push: tc.push}, &options.BuildOptions{})
			if (err != nil) != tc.wantErr {
				t.Fatalf("newBuildOutput() = %v, wantErr %v", err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}
			if got := out.Images[0].SBOM.Platforms[0].Ref; got != tc.wantRef {
				t.Errorf("ref = %q, want %q", got, tc.wantRef)
			}
		})
	}
}

// TestWriteJSONBuildOutput pins the JSON key names, that empty collections are
// [] or {} rather than null, that an image without SBOM has a null sbom, and
// that optional SBOM fields are omitted.
func TestWriteJSONBuildOutput(t *testing.T) {
	out := buildOutput{
		Version: "v1", Tags: []string{}, Labels: map[string]string{}, Annotations: []string{},
		Images: []imageOutput{{
			ImportPath: "example.com/a", Ref: "example.com/a@" + digestA, Digest: digestA, Platforms: []string{},
			SBOM: &sbomsOutput{Platforms: []sbomOutput{{Platform: "linux/amd64", Format: "spdx", MediaType: "text/spdx+json", Digest: digestB, Size: 1}}},
		}, {
			ImportPath: "example.com/b", Ref: "example.com/b@" + digestA, Digest: digestA, Platforms: []string{},
		}},
	}
	want := `{
  "version": "v1",
  "tags": [],
  "labels": {},
  "annotations": [],
  "images": [
    {
      "importpath": "example.com/a",
      "ref": "example.com/a@` + digestA + `",
      "digest": "` + digestA + `",
      "platforms": [],
      "sbom": {
        "index": null,
        "platforms": [
          {
            "platform": "linux/amd64",
            "format": "spdx",
            "mediaType": "text/spdx+json",
            "digest": "` + digestB + `",
            "size": 1
          }
        ]
      }
    },
    {
      "importpath": "example.com/b",
      "ref": "example.com/b@` + digestA + `",
      "digest": "` + digestA + `",
      "platforms": [],
      "sbom": null
    }
  ]
}
`
	var buf bytes.Buffer
	if err := writeJSONBuildOutput(&buf, out); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, buf.String()); diff != "" {
		t.Errorf("writeJSONBuildOutput() mismatch (-want +got):\n%s", diff)
	}
}

func TestWriteBuildOutput(t *testing.T) {
	var buf bytes.Buffer
	images := map[string]name.Reference{"ko://example.com/a": name.MustParseReference("example.com/a@" + digestA)}
	if err := writeBuildOutput(&buf, images); err != nil {
		t.Fatal(err)
	}
	if want := "example.com/a@" + digestA + "\n"; buf.String() != want {
		t.Errorf("writeBuildOutput() = %q, want %q", buf.String(), want)
	}
}

func TestSBOMPushed(t *testing.T) {
	for _, tc := range []struct {
		name string
		po   options.PublishOptions
		want bool
	}{
		{"registry push", options.PublishOptions{Push: true, DockerRepo: "example.com/app"}, true},
		{"--push=false", options.PublishOptions{Push: false, DockerRepo: "example.com/app"}, false},
		{"docker daemon", options.PublishOptions{Push: true, Local: true, DockerRepo: "ko.local"}, false},
		{"kind", options.PublishOptions{Push: true, DockerRepo: "kind.local"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := sbomPushed(&tc.po); got != tc.want {
				t.Errorf("sbomPushed() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSBOMFormat(t *testing.T) {
	for mt, want := range map[string]string{
		"text/spdx+json":                 "spdx",
		"text/spdx":                      "spdx",
		"application/vnd.cyclonedx+json": "cyclonedx",
		"application/vnd.syft+json":      "syft",
		"application/x-unknown":          "application/x-unknown",
	} {
		if got := sbomFormat(mt); got != want {
			t.Errorf("sbomFormat(%q) = %q, want %q", mt, got, want)
		}
	}
}
