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
	"encoding/json"
	"reflect"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/ko/pkg/publish"
)

func TestBuildOutputJSON(t *testing.T) {
	const (
		digestA = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		hexB    = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
		digestB = "sha256:" + hexB
	)
	for _, tc := range []struct {
		name        string
		tags        []string
		results     []publish.Published
		labels      []string
		annotations []string
		want        map[string]any
	}{{
		name: "empty",
		want: map[string]any{
			"tags": []any{}, "labels": map[string]any{}, "annotations": []any{}, "images": []any{},
		},
	}, {
		name: "digest-only refs get the requested tags",
		tags: []string{"latest", "v1"},
		results: []publish.Published{
			{ImportPath: "ko://example.com/b", Ref: name.MustParseReference("example.com/b@" + digestB), Digest: digestB, Platforms: []string{"linux/arm64"}},
			{ImportPath: "example.com/a", Ref: name.MustParseReference("localhost:5000/a@" + digestA), Digest: digestA, Platforms: []string{}},
		},
		labels:      []string{"org.opencontainers.image.title=app", "empty="},
		annotations: []string{"a=b"},
		want: map[string]any{
			"tags":        []any{"example.com/b:latest", "example.com/b:v1", "localhost:5000/a:latest", "localhost:5000/a:v1"},
			"labels":      map[string]any{"org.opencontainers.image.title": "app", "empty": ""},
			"annotations": []any{"a=b"},
			"images": []any{
				map[string]any{"importpath": "example.com/b", "ref": "example.com/b@" + digestB, "digest": digestB, "platforms": []any{"linux/arm64"}},
				map[string]any{"importpath": "example.com/a", "ref": "localhost:5000/a@" + digestA, "digest": digestA, "platforms": []any{}},
			},
		},
	}, {
		name: "tagged digest ref and tag ref lose their digest and hex tag",
		tags: []string{"v1"},
		results: []publish.Published{
			{ImportPath: "example.com/c", Ref: name.MustParseReference("example.com/c:v1@" + digestA), Digest: digestA, Platforms: []string{}},
			{ImportPath: "example.com/d", Ref: name.MustParseReference("ko.local/d:" + hexB), Digest: digestB, Platforms: []string{}},
		},
		want: map[string]any{
			"tags":        []any{"example.com/c:v1", "ko.local/d:v1"},
			"labels":      map[string]any{},
			"annotations": []any{},
			"images": []any{
				map[string]any{"importpath": "example.com/c", "ref": "example.com/c:v1@" + digestA, "digest": digestA, "platforms": []any{}},
				map[string]any{"importpath": "example.com/d", "ref": "ko.local/d:" + hexB, "digest": digestB, "platforms": []any{}},
			},
		},
	}, {
		name: "images sharing a name are not deduplicated",
		tags: []string{"latest"},
		results: []publish.Published{
			{ImportPath: "example.com/f", Ref: name.MustParseReference("example.com/shared@" + digestA), Digest: digestA, Platforms: []string{}},
			{ImportPath: "example.com/g", Ref: name.MustParseReference("example.com/shared@" + digestB), Digest: digestB, Platforms: []string{}},
		},
		want: map[string]any{
			"tags":        []any{"example.com/shared:latest", "example.com/shared:latest"},
			"labels":      map[string]any{},
			"annotations": []any{},
			"images": []any{
				map[string]any{"importpath": "example.com/f", "ref": "example.com/shared@" + digestA, "digest": digestA, "platforms": []any{}},
				map[string]any{"importpath": "example.com/g", "ref": "example.com/shared@" + digestB, "digest": digestB, "platforms": []any{}},
			},
		},
	}} {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := writeJSONBuildOutput(&buf, newBuildOutput(tc.results, tc.tags, tc.labels, tc.annotations)); err != nil {
				t.Fatal(err)
			}
			var got map[string]any
			if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
				t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
			}
			if _, ok := got["version"]; !ok {
				t.Error("missing version key")
			}
			delete(got, "version")
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}
