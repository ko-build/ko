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

package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/build"
	"io"
	"os/exec"
	"strings"

	"github.com/google/ko/pkg/commands/options"
	"github.com/spf13/cobra"
)

// addBuild augments our CLI surface with build.
func addBuild(topLevel *cobra.Command) {
	po := &options.PublishOptions{}
	bo := &options.BuildOptions{}

	build := &cobra.Command{
		Use:     "build IMPORTPATH...",
		Short:   "Build and publish container images from the given importpaths.",
		Long:    `This sub-command builds the provided import paths into Go binaries, containerizes them, and publishes them.`,
		Aliases: []string{"publish"},
		Example: `
  # Build and publish import path references to a Docker
  # Registry as:
  #   ${KO_DOCKER_REPO}/<package name>-<hash of import path>
  # When KO_DOCKER_REPO is ko.local, it is the same as if
  # --local and --preserve-import-paths were passed.
  ko build github.com/foo/bar/cmd/baz github.com/foo/bar/cmd/blah

  # Build and publish a relative import path as:
  #   ${KO_DOCKER_REPO}/<package name>-<hash of import path>
  # When KO_DOCKER_REPO is ko.local, it is the same as if
  # --local and --preserve-import-paths were passed.
  ko build ./cmd/blah

  # Build and publish a relative import path as:
  #   ${KO_DOCKER_REPO}/<import path>
  # When KO_DOCKER_REPO is ko.local, it is the same as if
  # --local was passed.
  ko build --preserve-import-paths ./cmd/blah

  # Build and publish import path references to a Docker
  # daemon as:
  #   ko.local/<import path>
  # This always preserves import paths.
  ko build --local github.com/foo/bar/cmd/baz github.com/foo/bar/cmd/blah`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := options.Validate(po, bo); err != nil {
				return fmt.Errorf("validating options: %w", err)
			}
			ctx := cmd.Context()

			uniq := map[string]struct{}{}
			for _, a := range args {
				if strings.HasSuffix(a, "/...") {
					matches, err := importPaths(ctx, a)
					if err != nil {
						return fmt.Errorf("resolving import path %q: %w", a, err)
					}
					for _, m := range matches {
						uniq[m] = struct{}{}
					}
				} else {
					uniq[a] = struct{}{}
				}
			}
			importpaths := make([]string, 0, len(uniq))
			for k := range uniq {
				importpaths = append(importpaths, k)
			}
			if len(importpaths) == 0 {
				return errors.New("no package main packages matched")
			}
			// TODO: sort?

			bo.InsecureRegistry = po.InsecureRegistry
			builder, err := makeBuilder(ctx, bo)
			if err != nil {
				return fmt.Errorf("error creating builder: %w", err)
			}
			publisher, err := makePublisher(po)
			if err != nil {
				return fmt.Errorf("error creating publisher: %w", err)
			}
			defer publisher.Close()
			images, err := publishImages(ctx, importpaths, publisher, builder)
			if err != nil {
				return fmt.Errorf("failed to publish images: %w", err)
			}
			for _, img := range images {
				fmt.Println(img)
			}
			return nil
		},
	}
	options.AddPublishArg(build, po)
	options.AddBuildOptions(build, bo)
	topLevel.AddCommand(build)
}

// importPaths resolves a wildcard importpath string like "./cmd/..." or
// "./..." and returns the 'package main' packages matched by that wildcard,
// using `go list`.
func importPaths(ctx context.Context, s string) ([]string, error) {
	var buf bytes.Buffer
	cmd := exec.CommandContext(ctx, "go", "list", "-json", s)
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	var out []string
	dec := json.NewDecoder(&buf)
	for {
		var p build.Package
		if err := dec.Decode(&p); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, err
		}
		if p.Name == "main" {
			out = append(out, p.ImportPath)
		}
	}
	return out, nil
}
