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
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/google/ko/pkg/commands/options"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/sync/errgroup"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// addCreate augments our CLI surface with apply.
func addCreate(topLevel *cobra.Command) {
	koCreateFlags := []string{}
	po := &options.PublishOptions{}
	fo := &options.FilenameOptions{}
	so := &options.SelectorOptions{}
	bo := &options.BuildOptions{}
	create := &cobra.Command{
		Use:   "create -f FILENAME",
		Short: "Create the input files with image references resolved to built/pushed image digests.",
		Long:  `This sub-command finds import path references within the provided files, builds them into Go binaries, containerizes them, publishes them, and then feeds the resulting yaml into "kubectl create".`,
		Example: `
  # Build and publish import path references to a Docker
  # Registry as:
  #   ${KO_DOCKER_REPO}/<package name>-<hash of import path>
  # Then, feed the resulting yaml into "kubectl create".
  # When KO_DOCKER_REPO is ko.local, it is the same as if
  # --local was passed.
  ko create -f config/

  # Build and publish import path references to a Docker
  # Registry preserving import path names as:
  #   ${KO_DOCKER_REPO}/<import path>
  # Then, feed the resulting yaml into "kubectl create".
  ko create --preserve-import-paths -f config/

  # Build and publish import path references to a Docker
  # daemon as:
  #   ko.local/<import path>
  # Then, feed the resulting yaml into "kubectl create".
  ko create --local -f config/

  # Create from stdin:
  cat config.yaml | ko create -f -`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !isKubectlAvailable() {
				return errors.New("error: kubectl is not available. kubectl must be installed to use ko create")
			}

			// Cancel on signals.
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
			// Create a set of ko-specific flags to ignore when passing through
			// kubectl global flags.
			ignoreSet := make(map[string]struct{})
			for _, s := range koCreateFlags {
				ignoreSet[s] = struct{}{}
			}

			// Filter out ko flags from what we will pass through to kubectl.
			kubectlFlags := []string{}
			cmd.Flags().Visit(func(flag *pflag.Flag) {
				if _, ok := ignoreSet[flag.Name]; !ok {
					kubectlFlags = append(kubectlFlags, "--"+flag.Name, flag.Value.String())
				}
			})

			// Issue a "kubectl create" command reading from stdin,
			// to which we will pipe the resolved files.
			argv := []string{"create", "-f", "-"}
			argv = append(argv, kubectlFlags...)
			kubectlCmd := exec.CommandContext(ctx, "kubectl", argv...)

			// Pass through our environment
			kubectlCmd.Env = os.Environ()
			// Pass through our std{out,err} and make our resolved buffer stdin.
			kubectlCmd.Stderr = os.Stderr
			kubectlCmd.Stdout = os.Stdout

			// Wire up kubectl stdin to resolveFilesToWriter.
			stdin, err := kubectlCmd.StdinPipe()
			if err != nil {
				return fmt.Errorf("error piping to 'kubectl create': %v", err)
			}

			// Make sure builds are cancelled if kubectl create fails.
			g, ctx := errgroup.WithContext(ctx)
			g.Go(func() error {
				// kubectl buffers data before starting to create it, which
				// can lead to resources being created more slowly than desired.
				// In the case of --watch, it can lead to resources not being
				// applied at all until enough iteration has occurred.  To work
				// around this, we prime the stream with a bunch of empty objects
				// which kubectl will discard.
				// See https://github.com/google/go-containerregistry/pull/348
				for i := 0; i < 1000; i++ {
					stdin.Write([]byte("---\n"))
				}
				// Once primed kick things off.
				return resolveFilesToWriter(ctx, builder, publisher, fo, so, stdin)
			})

			g.Go(func() error {
				// Run it.
				if err := kubectlCmd.Run(); err != nil {
					return fmt.Errorf("error executing 'kubectl create': %v", err)
				}
				return nil
			})

			return g.Wait()
		},
	}
	options.AddPublishArg(create, po)
	options.AddFileArg(create, fo)
	options.AddSelectorArg(create, so)
	options.AddBuildOptions(create, bo)

	// Collect the ko-specific apply flags before registering the kubectl global
	// flags so that we can ignore them when passing kubectl global flags through
	// to kubectl.
	create.Flags().VisitAll(func(flag *pflag.Flag) {
		koCreateFlags = append(koCreateFlags, flag.Name)
	})

	// Register the kubectl global flags.
	kubeConfigFlags := genericclioptions.NewConfigFlags(false)
	kubeConfigFlags.AddFlags(create.Flags())

	topLevel.AddCommand(create)
}
