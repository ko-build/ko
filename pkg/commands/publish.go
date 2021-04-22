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
	"fmt"

	"github.com/google/ko/pkg/commands/options"
	"github.com/spf13/cobra"
)

// addPublish augments our CLI surface with publish.
func addPublish(topLevel *cobra.Command) {
	po := &options.PublishOptions{}
	bo := &options.BuildOptions{}

	publish := &cobra.Command{
		Use:   "publish IMPORTPATH...",
		Short: "Build and publish container images from the given importpaths.",
		Long:  `This sub-command builds the provided import paths into Go binaries, containerizes them, and publishes them.`,
		Example: `
  # Build and publish import path references to a Docker
  # Registry as:
  #   ${KO_DOCKER_REPO}/<package name>-<hash of import path>
  # When KO_DOCKER_REPO is ko.local, it is the same as if
  # --local and --preserve-import-paths were passed.
  ko publish github.com/foo/bar/cmd/baz github.com/foo/bar/cmd/blah

  # Build and publish a relative import path as:
  #   ${KO_DOCKER_REPO}/<package name>-<hash of import path>
  # When KO_DOCKER_REPO is ko.local, it is the same as if
  # --local and --preserve-import-paths were passed.
  ko publish ./cmd/blah

  # Build and publish a relative import path as:
  #   ${KO_DOCKER_REPO}/<import path>
  # When KO_DOCKER_REPO is ko.local, it is the same as if
  # --local was passed.
  ko publish --preserve-import-paths ./cmd/blah

  # Build and publish import path references to a Docker
  # daemon as:
  #   ko.local/<import path>
  # This always preserves import paths.
  ko publish --local github.com/foo/bar/cmd/baz github.com/foo/bar/cmd/blah`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			ctx := createCancellableContext()
			builder, err := makeBuilder(ctx, bo)
			if err != nil {
				return fmt.Errorf("error creating builder: %v", err)
			}
			publisher, err := makePublisher(po)
			if err != nil {
				return fmt.Errorf("error creating publisher: %v", err)
			}
			defer publisher.Close()
			images, err := publishImages(ctx, args, publisher, builder)
			if err != nil {
				return fmt.Errorf("failed to publish images: %v", err)
			}
			for _, img := range images {
				fmt.Println(img)
			}
			return nil
		},
	}
	options.AddPublishArg(publish, po)
	options.AddBuildOptions(publish, bo)
	topLevel.AddCommand(publish)
}
