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
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/ko/pkg/publish"
)

// buildOutput is the document written by `ko build -O json`. The tags, labels
// and annotations keys follow the json output of docker/metadata-action; the
// remaining keys are ko extras.
type buildOutput struct {
	Version     string            `json:"version"`
	Tags        []string          `json:"tags"` // repo:tag, without digest
	Labels      map[string]string `json:"labels"`
	Annotations []string          `json:"annotations"`
	Images      []imageOutput     `json:"images"`
}

type imageOutput struct {
	ImportPath string   `json:"importpath"`
	Ref        string   `json:"ref"`
	Digest     string   `json:"digest"`
	Platforms  []string `json:"platforms"`
}

// newBuildOutput describes results. tags are the tags requested with --tags,
// which the publishers apply to every image; the ref a publisher returns may
// not carry any of them (it is a bare digest ref for the default publisher).
func newBuildOutput(results []publish.Published, tags, labels, annotations []string) buildOutput {
	out := buildOutput{
		Version:     version(),
		Tags:        []string{},
		Labels:      map[string]string{},
		Annotations: append([]string{}, annotations...),
		Images:      []imageOutput{},
	}
	for _, l := range labels {
		k, v, _ := strings.Cut(l, "=")
		out.Labels[k] = v
	}
	for _, r := range results {
		for _, t := range tags {
			out.Tags = append(out.Tags, r.Ref.Context().Tag(t).String())
		}
		out.Images = append(out.Images, imageOutput{
			ImportPath: strings.TrimPrefix(r.ImportPath, "ko://"),
			Ref:        r.Ref.String(),
			Digest:     r.Digest,
			Platforms:  r.Platforms,
		})
	}
	sort.Strings(out.Tags)
	return out
}

func writeJSONBuildOutput(w io.Writer, out buildOutput) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		return fmt.Errorf("encoding json output: %w", err)
	}
	return nil
}

func writeBuildOutput(w io.Writer, images map[string]name.Reference) error {
	for _, img := range images {
		fmt.Fprintln(w, img)
	}
	return nil
}
