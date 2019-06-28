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
	"os"

	"github.com/spf13/cobra"
)

type CompletionFlags struct {
	Zsh bool
}

func addCompletion(topLevel *cobra.Command) {
	var completionFlags CompletionFlags

	completionCmd := &cobra.Command{
		Use:   "completion",
		Short: "Output shell completion code (default Bash)",
		Run: func(cmd *cobra.Command, args []string) {
			if completionFlags.Zsh {
				cmd.Root().GenZshCompletion(os.Stdout)
			} else {
				cmd.Root().GenBashCompletion(os.Stdout)
			}
		},
	}

	completionCmd.Flags().BoolVar(&completionFlags.Zsh, "zsh", false, "Generates completion code for Zsh shell.")
	topLevel.AddCommand(completionCmd)
}
