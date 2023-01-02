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
	"github.com/spf13/cobra"
)

// PublicKeyOptions is the top level wrapper for the public-key command.
type PublicKeyOptions struct {
	Key         string
	SecurityKey SecurityKeyOptions
	OutFile     string
}

var _ Interface = (*PublicKeyOptions)(nil)

// AddFlags implements Interface
func (o *PublicKeyOptions) AddFlags(cmd *cobra.Command) {
	o.SecurityKey.AddFlags(cmd)

	cmd.Flags().StringVar(&o.Key, "key", "",
		"path to the private key file, KMS URI or Kubernetes Secret")
	_ = cmd.Flags().SetAnnotation("key", cobra.BashCompFilenameExt, []string{})

	cmd.Flags().StringVar(&o.OutFile, "outfile", "",
		"path to a payload file to use rather than generating one")
	_ = cmd.Flags().SetAnnotation("outfile", cobra.BashCompFilenameExt, []string{})
}
