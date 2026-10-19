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
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/ko/pkg/build"
	"github.com/google/ko/pkg/commands/options"
	"github.com/google/ko/pkg/publish"
	ctypes "github.com/sigstore/cosign/v3/pkg/types"
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
	// SBOM is null when the image has no SBOM, e.g. with --sbom=none.
	SBOM *sbomsOutput `json:"sbom"`
}

type sbomsOutput struct {
	// Index is the SBOM of the image index, null when no index was published.
	Index *sbomOutput `json:"index"`
	// Platforms are the SBOMs of the platform images.
	Platforms []sbomOutput `json:"platforms"`
}

type sbomOutput struct {
	Platform  string `json:"platform,omitempty"`
	Format    string `json:"format"`
	MediaType string `json:"mediaType"`
	// Digest is the sha256 of the SBOM document.
	Digest string `json:"digest"`
	Size   int64  `json:"size"`
	// Ref is only set when the SBOM was pushed to a registry.
	Ref string `json:"ref,omitempty"`
	// Path is only set when the SBOM directory holds this very document.
	Path string `json:"path,omitempty"`
}

// newBuildOutput describes results, published with po and built with bo.
//
// Tags come from po.Tags, which the publishers apply to every image: the ref
// a publisher returns may carry none of them (the default publisher returns a
// bare digest ref). SBOMs only get a ref when they were pushed to a registry,
// and a path when they were written to bo.SBOMDir.
func newBuildOutput(results []publish.Published, po *options.PublishOptions, bo *options.BuildOptions) (buildOutput, error) {
	out := buildOutput{
		Version:     version(),
		Tags:        []string{},
		Labels:      map[string]string{},
		Annotations: append([]string{}, bo.Annotations...),
		Images:      []imageOutput{},
	}
	for _, l := range bo.Labels {
		k, v, _ := strings.Cut(l, "=")
		out.Labels[k] = v
	}
	sboms := sbomDescriber{pushed: sbomPushed(po)}
	if bo.SBOMDir != "" {
		dir, err := filepath.Abs(bo.SBOMDir)
		if err != nil {
			return buildOutput{}, fmt.Errorf("resolving the SBOM directory: %w", err)
		}
		sboms.dir = dir
	}
	for _, r := range results {
		for _, t := range po.Tags {
			out.Tags = append(out.Tags, r.Ref.Context().Tag(t).String())
		}
		so, err := sboms.describeAll(r.ImportPath, r.SBOMs)
		if err != nil {
			return buildOutput{}, err
		}
		out.Images = append(out.Images, imageOutput{
			ImportPath: r.ImportPath,
			Ref:        r.Ref.String(),
			Digest:     r.Digest,
			Platforms:  r.Platforms,
			SBOM:       so,
		})
	}
	sort.Strings(out.Tags)
	return out, nil
}

// sbomDescriber turns SBOMs into their output.
type sbomDescriber struct {
	pushed bool   // whether the SBOMs were pushed to a registry
	dir    string // the absolute SBOM directory, if any
}

// describeAll describes the SBOMs of an image, or returns nil if it has none.
func (d sbomDescriber) describeAll(importpath string, sboms publish.SBOMs) (*sbomsOutput, error) {
	if sboms.Index == nil && len(sboms.Platforms) == 0 {
		return nil, nil
	}
	out := &sbomsOutput{Platforms: make([]sbomOutput, 0, len(sboms.Platforms))}
	if sboms.Index != nil {
		so, err := d.describe(importpath, *sboms.Index)
		if err != nil {
			return nil, err
		}
		out.Index = &so
	}
	for _, sb := range sboms.Platforms {
		so, err := d.describe(importpath, sb)
		if err != nil {
			return nil, err
		}
		out.Platforms = append(out.Platforms, so)
	}
	return out, nil
}

func (d sbomDescriber) describe(importpath string, sb publish.SBOM) (sbomOutput, error) {
	so := sbomOutput{Format: sbomFormat(sb.MediaType), MediaType: sb.MediaType, Digest: sb.Digest, Size: sb.Size}
	if sb.Platform != nil {
		so.Platform = sb.Platform.String()
	}
	if d.pushed {
		ref, err := publish.SBOMRef(sb)
		if err != nil {
			return sbomOutput{}, err
		}
		so.Ref = ref.String()
	}
	if d.dir != "" {
		path, err := sbomPath(d.dir, importpath, sb)
		if err != nil {
			return sbomOutput{}, err
		}
		so.Path = path
	}
	return so, nil
}

// sbomPath returns the file in dir holding sb, or "" if there is none: a file
// is only reported when its content matches sb's digest, as import paths that
// share a base name overwrite each other's file. A single-platform image's
// file is named for the platform it was built for, which may differ from its
// config's in the variant only: either may lack it.
func sbomPath(dir, importpath string, sb publish.SBOM) (string, error) {
	candidates := []string{build.SBOMPath(dir, importpath, sb.Platform)}
	if p := sb.Platform; p != nil {
		other := *p
		if p.Variant != "" {
			other.Variant = ""
			candidates = append(candidates, build.SBOMPath(dir, importpath, &other))
		} else {
			other.Variant = "*"
			matches, err := matchFiles(build.SBOMPath(dir, importpath, &other))
			if err != nil {
				return "", err
			}
			candidates = append(candidates, matches...)
		}
	}
	for _, path := range candidates {
		b, err := os.ReadFile(path)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return "", fmt.Errorf("reading SBOM: %w", err)
		}
		if fmt.Sprintf("sha256:%x", sha256.Sum256(b)) == sb.Digest {
			return path, nil
		}
	}
	return "", nil
}

// matchFiles returns the files in pattern's directory whose name matches
// pattern's base name. Only the base name is a pattern, so that the directory
// may contain glob metacharacters.
func matchFiles(pattern string) ([]string, error) {
	dir := filepath.Dir(pattern)
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading SBOM directory: %w", err)
	}
	var out []string
	for _, e := range entries {
		ok, err := filepath.Match(filepath.Base(pattern), e.Name())
		if err != nil {
			return nil, fmt.Errorf("matching SBOM files: %w", err)
		}
		if ok && !e.IsDir() {
			out = append(out, filepath.Join(dir, e.Name()))
		}
	}
	return out, nil
}

// sbomFormat names the SBOM format of a media type, e.g. "spdx". Unknown media
// types are returned as is.
func sbomFormat(mediaType string) string {
	switch mediaType {
	case ctypes.SPDXMediaType, ctypes.SPDXJSONMediaType:
		return "spdx"
	case ctypes.CycloneDXJSONMediaType, ctypes.CycloneDXXMLMediaType:
		return "cyclonedx"
	case ctypes.SyftMediaType:
		return "syft"
	default:
		return mediaType
	}
}

// writeJSONBuildOutput writes out as indented JSON.
func writeJSONBuildOutput(w io.Writer, out buildOutput) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		return fmt.Errorf("encoding json output: %w", err)
	}
	return nil
}

// writeBuildOutput prints the published image references, one per line.
func writeBuildOutput(w io.Writer, images map[string]name.Reference) error {
	for _, img := range images {
		if _, err := fmt.Fprintln(w, img); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
	}
	return nil
}
