// Copyright 2019 Google LLC All Rights Reserved.
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
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// provided by govvv in compile-time
var Version string

// addVersion augments our CLI surface with version.
func addVersion(topLevel *cobra.Command) {
	topLevel.AddCommand(&cobra.Command{
		Use:   "version",
		Short: `Print ko version.`,
		Run: func(cmd *cobra.Command, args []string) {
			version()
		},
	})
}

func version() {
	if Version == "" {
		gitDir := fmt.Sprintf("--git-dir=%v/src/github.com/google/ko/.git", os.Getenv("GOPATH"))
		hash, err := exec.Command("git", gitDir, "rev-parse", "HEAD").Output()
		if err != nil {
			log.Fatalf("error during command execution: %v", err)
		}
		Version = strings.TrimSpace(string(hash))
	}
	fmt.Printf("version: %v\n", Version)
}
