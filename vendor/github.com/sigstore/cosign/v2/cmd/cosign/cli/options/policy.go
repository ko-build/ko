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

// PolicyInitOptions is the top level wrapper for the policy-init command.
type PolicyInitOptions struct {
	ImageRef    string
	Maintainers []string
	Issuer      string
	Threshold   int
	Expires     int
	OutFile     string
	Registry    RegistryOptions
}

var _ Interface = (*PolicyInitOptions)(nil)

// AddFlags implements Interface
func (o *PolicyInitOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.ImageRef, "namespace", "ns",
		"registry namespace that the root policy belongs to")

	cmd.Flags().StringVar(&o.OutFile, "out", "o",
		"output policy locally")
	_ = cmd.Flags().SetAnnotation("out", cobra.BashCompSubdirsInDir, []string{})

	cmd.Flags().StringVar(&o.Issuer, "issuer", "",
		"trusted issuer to use for identity tokens, e.g. https://accounts.google.com")

	cmd.Flags().IntVar(&o.Threshold, "threshold", 1,
		"threshold for root policy signers")

	cmd.Flags().StringSliceVarP(&o.Maintainers, "maintainers", "m", nil,
		"list of maintainers to add to the root policy")

	cmd.Flags().IntVar(&o.Expires, "expires", 0,
		"total expire duration in days")

	o.Registry.AddFlags(cmd)
}

type PolicySignOptions struct {
	ImageRef         string
	OutFile          string
	Registry         RegistryOptions
	Fulcio           FulcioOptions
	Rekor            RekorOptions
	SkipConfirmation bool
	TlogUpload       bool
	TSAServerURL     string

	OIDC OIDCOptions
}

var _ Interface = (*PolicySignOptions)(nil)

// AddFlags implements Interface
func (o *PolicySignOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.ImageRef, "namespace", "ns",
		"registry namespace that the root policy belongs to")

	cmd.Flags().StringVar(&o.OutFile, "out", "o",
		"output policy locally")

	cmd.Flags().BoolVarP(&o.SkipConfirmation, "yes", "y", false,
		"skip confirmation prompts for non-destructive operations")

	cmd.Flags().BoolVar(&o.TlogUpload, "tlog-upload", true,
		"whether or not to upload to the tlog")

	cmd.Flags().StringVar(&o.TSAServerURL, "timestamp-server-url", "",
		"url to the Timestamp RFC3161 server, default none")

	o.Registry.AddFlags(cmd)
	o.Fulcio.AddFlags(cmd)
	o.Rekor.AddFlags(cmd)
	o.OIDC.AddFlags(cmd)
}
