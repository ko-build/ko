// Copyright 2018 ko Build Authors All Rights Reserved.
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
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/ko/pkg/commands/options"
	"github.com/spf13/cobra"
)

// addRun augments our CLI surface with run.
func addRun(topLevel *cobra.Command) {
	po := &options.PublishOptions{}
	bo := &options.BuildOptions{}

	run := &cobra.Command{
		Use:   "run IMPORTPATH",
		Short: "A variant of `kubectl run` that containerizes IMPORTPATH first.",
		Long:  `This sub-command combines "ko build" and "kubectl run" to support containerizing and running Go binaries on Kubernetes in a single command.`,
		Example: `
  # Publish the image and run it on Kubernetes as:
  #   ${KO_DOCKER_REPO}/<package name>-<hash of import path>
  # When KO_DOCKER_REPO is ko.local, it is the same as if
  # --local and --preserve-import-paths were passed.
  ko run github.com/foo/bar/cmd/baz

  # This supports relative import paths as well.
  ko run ./cmd/baz

  # You can also supply args and flags to the command.
  ko run ./cmd/baz -- -v arg1 arg2 --yes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := options.Validate(po, bo); err != nil {
				return fmt.Errorf("validating options: %w", err)
			}

			ctx := cmd.Context()

			// Args after -- are for kubectl, so only consider importPaths before it.
			importPaths := args
			dashes := cmd.Flags().ArgsLenAtDash()
			if dashes != -1 {
				importPaths = args[:cmd.Flags().ArgsLenAtDash()]
			}
			if len(importPaths) == 0 {
				return errors.New("ko run: no importpaths listed")
			}

			kubectlArgs := []string{}
			dashes = unparsedDashes()
			if dashes != -1 && dashes != len(os.Args) {
				kubectlArgs = os.Args[dashes+1:]
			}

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

			if len(os.Args) < 3 {
				return fmt.Errorf("usage: %s run <package>", os.Args[0])
			}
			ip := os.Args[2]
			if strings.HasPrefix(ip, "-") {
				return fmt.Errorf("expected first arg to be positional, got %q", ip)
			}
			imgs, err := publishImages(ctx, importPaths, publisher, builder)
			if err != nil {
				return fmt.Errorf("failed to publish images: %w", err)
			}

			// Usually only one, but this is the simple way to access the
			// reference since the import path may have been qualified.
			for k, ref := range imgs {
				log.Printf("Running %q", k)
				pod := filepath.Base(ref.Context().String())

				// These are better defaults:
				defaults := []string{
					"--attach",                 // stream logs back
					"--rm",                     // clean up after ourselves
					"--restart=Never",          // we just want to run once
					"--log-flush-frequency=1s", // flush logs more often
				}

				// Replaced "<package>" with "--image=<published image>".
				argv := []string{"--image", ref.String()}

				// Add our default kubectl flags.
				// TODO: Add some way to override these.
				argv = append(argv, defaults...)

				// If present, adds -- arg1 arg2...
				argv = append(argv, kubectlArgs...)

				// "run <package> <defaults> --image <ref> <kubectlArgs>"
				argv = append([]string{"run", pod}, argv...)

				/* #nosec G706 -- argv is the kubectl invocation we just built; logging it is the intent. */
				log.Printf("$ kubectl %s", strings.Join(argv, " "))
				if err := runKubectl(ctx, argv); err != nil {
					return err
				}
			}
			return nil
		},
		// We ignore unknown flags to avoid importing everything Go exposes
		// from our commands.
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
	}
	options.AddPublishArg(run, po)
	options.AddBuildOptions(run, bo)

	topLevel.AddCommand(run)
}

func unparsedDashes() int {
	for i, s := range os.Args {
		if s == "--" {
			return i
		}
	}
	return -1
}

// runKubectl runs kubectl with argv. Unlike exec.CommandContext, a canceled
// context sends os.Interrupt so kubectl --rm can delete the pod instead of
// being SIGKILLed immediately.
func runKubectl(ctx context.Context, argv []string) error {
	/* #nosec G204 G702 -- ko run intentionally invokes kubectl with user-supplied arguments. */
	cmd := exec.Command("kubectl", argv...)
	cmd.Env = os.Environ()
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	return runInterruptible(ctx, cmd)
}

func runInterruptible(ctx context.Context, cmd *exec.Cmd) error {
	if err := cmd.Start(); err != nil {
		return err
	}

	waitErr := make(chan error, 1)
	go func() {
		waitErr <- cmd.Wait()
	}()

	select {
	case err := <-waitErr:
		return err
	case <-ctx.Done():
		if err := cmd.Process.Signal(os.Interrupt); err != nil {
			_ = cmd.Process.Kill()
		}
		return <-waitErr
	}
}
