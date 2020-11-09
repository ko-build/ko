// Copyright 2020 Google LLC All Rights Reserved.
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

package cmd

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/types"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/spf13/cobra"
)

func init() { Root.AddCommand(NewCmdAuth()) }

// NewCmdAuth creates a new cobra.Command for the auth subcommand.
func NewCmdAuth() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Log in or access credentials",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}
	cmd.AddCommand(NewCmdAuthGet(), NewCmdAuthLogin())
	return cmd
}

// NewCmdAuthGet creates a new `crane auth get` command.
func NewCmdAuthGet() *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "Implements a credential helper",
		Example: `  # Read configured credentials for reg.example.com
  echo "reg.example.com" | crane auth get
  {"username":"AzureDiamond","password":"hunter2"}`,
		Args: cobra.NoArgs,
		Run: func(_ *cobra.Command, args []string) {
			b, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				log.Fatal(err)
			}
			reg, err := name.NewRegistry(strings.TrimSpace(string(b)))
			if err != nil {
				log.Fatal(err)
			}
			auther, err := authn.DefaultKeychain.Resolve(reg)
			if err != nil {
				log.Fatal(err)
			}
			auth, err := auther.Authorization()
			if err != nil {
				log.Fatal(err)
			}
			if err := json.NewEncoder(os.Stdout).Encode(auth); err != nil {
				log.Fatal(err)
			}
		},
	}
}

// NewCmdAuthLogin creates a new `crane auth login` command.
func NewCmdAuthLogin() *cobra.Command {
	var opts loginOptions

	cmd := &cobra.Command{
		Use:   "login [OPTIONS] [SERVER]",
		Short: "Log in to a registry",
		Example: `  # Log in to reg.example.com
  crane auth login reg.example.com -u AzureDiamond -p hunter2`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			reg, err := name.NewRegistry(args[0])
			if err != nil {
				log.Fatal(err)
			}

			opts.serverAddress = reg.Name()

			if err := login(opts); err != nil {
				log.Fatal(err)
			}
		},
	}

	flags := cmd.Flags()

	flags.StringVarP(&opts.user, "username", "u", "", "Username")
	flags.StringVarP(&opts.password, "password", "p", "", "Password")
	flags.BoolVarP(&opts.passwordStdin, "password-stdin", "", false, "Take the password from stdin")

	return cmd
}

type loginOptions struct {
	serverAddress string
	user          string
	password      string
	passwordStdin bool
}

func login(opts loginOptions) error {
	if opts.passwordStdin {
		contents, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}

		opts.password = strings.TrimSuffix(string(contents), "\n")
		opts.password = strings.TrimSuffix(opts.password, "\r")
	}
	if opts.user == "" && opts.password == "" {
		return errors.New("username and password required")
	}
	cf, err := config.Load(os.Getenv("DOCKER_CONFIG"))
	if err != nil {
		return err
	}
	creds := cf.GetCredentialsStore(opts.serverAddress)
	if opts.serverAddress == name.DefaultRegistry {
		opts.serverAddress = authn.DefaultAuthKey
	}
	if err := creds.Store(types.AuthConfig{
		ServerAddress: opts.serverAddress,
		Username:      opts.user,
		Password:      opts.password,
	}); err != nil {
		return err
	}

	return cf.Save()
}
