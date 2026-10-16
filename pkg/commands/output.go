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
	"github.com/google/ko/pkg/commands/options"
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
	// SBOM lists the SBOMs of the image: the top-level entity first, then one
	// per platform image of an index. It is empty with --sbom=none.
	SBOM []sbomOutput `json:"sbom"`
}

type sbomOutput struct {
	// Platform is empty for the top-level entity (the index, or a single image).
	Platform  string `json:"platform,omitempty"`
	Format    string `json:"format"`
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
	// Ref is only set when the SBOM was pushed to a registry.
	Ref string `json:"ref,omitempty"`
}

// newBuildOutput describes results. tags are the tags requested with --tags,
// which the publishers apply to every image; the ref a publisher returns may
// not carry any of them (it is a bare digest ref for the default publisher).
//
// sbomPushed tells whether the publisher pushed the SBOMs to a registry; only
// then do the SBOMs get a ref.
func newBuildOutput(results []publish.Published, tags, labels, annotations []string, sbomPushed bool) buildOutput {
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
		sboms := make([]sbomOutput, 0, len(r.SBOMs))
		for _, sb := range r.SBOMs {
			so := sbomOutput{Platform: sb.Platform, Format: sbomFormat(sb.MediaType), MediaType: sb.MediaType, Digest: sb.Digest, Size: sb.Size}
			if sbomPushed && sb.Ref != nil {
				so.Ref = sb.Ref.String()
			}
			sboms = append(sboms, so)
		}
		out.Images = append(out.Images, imageOutput{
			ImportPath: strings.TrimPrefix(r.ImportPath, "ko://"),
			Ref:        r.Ref.String(),
			Digest:     r.Digest,
			Platforms:  r.Platforms,
			SBOM:       sboms,
		})
	}
	sort.Strings(out.Tags)
	return out
}

// sbomFormat names the SBOM format of a media type, e.g. "spdx".
func sbomFormat(mediaType string) string {
	switch mt := strings.ToLower(mediaType); {
	case strings.Contains(mt, "spdx"):
		return "spdx"
	case strings.Contains(mt, "cyclonedx"):
		return "cyclonedx"
	default:
		return mediaType
	}
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

// sbomPushed tells whether the publisher built from po pushes SBOMs to a
// registry: only the default registry publisher does, not the docker daemon,
// kind, or --push=false. It must be called after makePublisher, which settles
// po.Local.
func sbomPushed(po *options.PublishOptions) bool {
	return po.Push && !po.Local && !strings.HasPrefix(po.DockerRepo, publish.KindDomain)
}
