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
	"github.com/google/ko/pkg/commands/options"
	"log"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// addRun augments our CLI surface with run.
func addRun(topLevel *cobra.Command) {
	lo := &options.LocalOptions{}
	bo := &options.BinaryOptions{}
	no := &options.NameOptions{}
	ta := &options.TagsOptions{}
	do := &options.DebugOptions{}

	run := &cobra.Command{
		Use:   "run NAME --image=IMPORTPATH",
		Short: "A variant of `kubectl run` that containerizes IMPORTPATH first.",
		Long:  `This sub-command combines "ko publish" and "kubectl run" to support containerizing and running Go binaries on Kubernetes in a single command.`,
		Example: `
  # Publish the --image and run it on Kubernetes as:
  #   ${KO_DOCKER_REPO}/<package name>-<hash of import path>
  # When KO_DOCKER_REPO is ko.local, it is the same as if
  # --local and --preserve-import-paths were passed.
  ko run foo --image=github.com/foo/bar/cmd/baz

  # This supports relative import paths as well.
  ko run foo --image=./cmd/baz`,
		Run: func(cmd *cobra.Command, args []string) {
			imgs := publishImages([]string{bo.Path}, no, lo, ta, do)

			// There's only one, but this is the simple way to access the
			// reference since the import path may have been qualified.
			for k, v := range imgs {
				log.Printf("Running %q", k)
				// Issue a "kubectl run" command with our same arguments,
				// but supply a second --image to override the one we intercepted.
				argv := append(os.Args[1:], "--image", v.String())
				kubectlCmd := exec.Command("kubectl", argv...)

				// Pass through our environment
				kubectlCmd.Env = os.Environ()
				// Pass through our std*
				kubectlCmd.Stderr = os.Stderr
				kubectlCmd.Stdout = os.Stdout
				kubectlCmd.Stdin = os.Stdin

				// Run it.
				if err := kubectlCmd.Run(); err != nil {
					log.Fatalf("error executing \"kubectl run\": %v", err)
				}
			}
		},
		// We ignore unknown flags to avoid importing everything Go exposes
		// from our commands.
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
	}
	options.AddLocalArg(run, lo)
	options.AddNamingArgs(run, no)
	options.AddImageArg(run, bo)
	options.AddTagsArg(run, ta)
	options.AddDebugArg(run, do)

	topLevel.AddCommand(run)
}
