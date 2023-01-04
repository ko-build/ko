//
// Copyright 2021 The Sigstore Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package options

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const EnvPrefix = "COSIGN"

// RootOptions define flags and options for the root cosign cli.
type RootOptions struct {
	OutputFile string
	Verbose    bool
	Timeout    time.Duration
}

// DefaultTimeout specifies the default timeout for commands.
const DefaultTimeout = 3 * time.Minute

var _ Interface = (*RootOptions)(nil)

// AddFlags implements Interface
func (o *RootOptions) AddFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&o.OutputFile, "output-file", "",
		"log output to a file")
	_ = cmd.Flags().SetAnnotation("output-file", cobra.BashCompFilenameExt, []string{})

	cmd.PersistentFlags().BoolVarP(&o.Verbose, "verbose", "d", false,
		"log debug output")

	cmd.PersistentFlags().DurationVarP(&o.Timeout, "timeout", "t", DefaultTimeout,
		"timeout for commands")
}

func BindViper(cmd *cobra.Command, args []string) {
	callPersistentPreRun(cmd, args)
	v := viper.New()
	v.SetEnvPrefix(EnvPrefix)
	v.AutomaticEnv()
	bindFlags(cmd, v)
}

// callPersistentPreRun calls parent commands. PersistentPreRun
// does not call parents PersistentPreRun functions
func callPersistentPreRun(cmd *cobra.Command, args []string) {
	if parent := cmd.Parent(); parent != nil {
		if parent.PersistentPreRun != nil {
			parent.PersistentPreRun(parent, args)
		}
		if parent.PersistentPreRunE != nil {
			err := parent.PersistentPreRunE(parent, args)
			if err != nil {
				cmd.PrintErrln("Error:", err.Error())
				os.Exit(1)
			}
		}
		callPersistentPreRun(parent, args)
	}
}

func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if strings.Contains(f.Name, "-") {
			_ = v.BindEnv(f.Name, flagToEnvVar(f.Name))
		}
		if !f.Changed && v.IsSet((f.Name)) {
			val := v.Get(f.Name)
			_ = cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}

func flagToEnvVar(f string) string {
	f = strings.ToUpper(f)
	return fmt.Sprintf("%s_%s", EnvPrefix, strings.ReplaceAll(f, "-", "_"))
}
