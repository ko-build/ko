//
// Copyright 2022 The Sigstore Authors.
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
	"github.com/spf13/cobra"
)

// EnvOptions is the top level wrapper for the env command.
type EnvOptions struct {
	ShowDescriptions    bool
	ShowSensitiveValues bool
}

var _ Interface = (*EnvOptions)(nil)

// AddFlags implements Interface
func (o *EnvOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.ShowDescriptions, "show-descriptions", true,
		"show descriptions for environment variables")

	cmd.Flags().BoolVar(&o.ShowSensitiveValues, "show-sensitive-values", false,
		"show values of sensitive environment variables")
}
