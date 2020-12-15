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
	"log"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// runCmd is suitable for use with cobra.Command's Run field.
type runCmd func(*cobra.Command, []string)

// passthru returns a runCmd that simply passes our CLI arguments
// through to a binary named command.
func passthru(command string) runCmd {
	return func(_ *cobra.Command, _ []string) {
		if !isKubectlAvailable() {
			log.Print("error: kubectl is not available. kubectl must be installed to use ko delete.")
			return
		}

		// Cancel on signals.
		ctx := createCancellableContext()

		// Start building a command line invocation by passing
		// through our arguments to command's CLI.
		cmd := exec.CommandContext(ctx, command, os.Args[1:]...)

		// Pass through our environment
		cmd.Env = os.Environ()
		// Pass through our stdfoo
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin

		// Run it.
		if err := cmd.Run(); err != nil {
			log.Fatalf("error executing %q command with args: %v; %v", command, os.Args[1:], err)
		}
	}
}

// addDelete augments our CLI surface with publish.
func addDelete(topLevel *cobra.Command) {
	topLevel.AddCommand(&cobra.Command{
		Use:   "delete",
		Short: `See "kubectl help delete" for detailed usage.`,
		Run:   passthru("kubectl"),
		// We ignore unknown flags to avoid importing everything Go exposes
		// from our commands.
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
	})
}
