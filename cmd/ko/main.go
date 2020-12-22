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

package main

// HEY! YOU! This file has moved to the root of the project.
// !! PLEASE DO NOT ADD NEW FEATURES HERE !!
// Only sync with the root/main.go.

import (
	"log"
	"os"

	"github.com/google/go-containerregistry/pkg/logs"
	"github.com/google/ko/pkg/commands"

	cranecmd "github.com/google/go-containerregistry/cmd/crane/cmd"
	"github.com/spf13/cobra"
)

const Deprecation258 = `NOTICE!
-----------------------------------------------------------------
Please install ko from github.com/google/ko.

For more information see:
   https://github.com/google/ko/issues/258
-----------------------------------------------------------------
`

func main() {
	logs.Warn.SetOutput(os.Stderr)
	logs.Progress.SetOutput(os.Stderr)

	// Parent command to which all subcommands are added.
	cmds := &cobra.Command{
		Use:   "ko",
		Short: "Rapidly iterate with Go, Containers, and Kubernetes.",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
	commands.AddKubeCommands(cmds)

	// Also add the auth group from crane to facilitate logging into a
	// registry.
	authCmd := cranecmd.NewCmdAuth("ko", "auth")
	// That was a mistake, but just set it to Hidden so we don't break people.
	authCmd.Hidden = true
	cmds.AddCommand(authCmd)

	// Just add a `ko login` command:
	cmds.AddCommand(cranecmd.NewCmdAuthLogin())

	log.Print(Deprecation258)

	if err := cmds.Execute(); err != nil {
		log.Fatalf("error during command execution: %v", err)
	}
}
