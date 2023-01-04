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

// PKCS11ToolListTokens is the wrapper for `pkcs11-tool list-tokens` related options.
type PKCS11ToolListTokensOptions struct {
	ModulePath string
}

var _ Interface = (*PKCS11ToolListTokensOptions)(nil)

// AddFlags implements Interface
func (o *PKCS11ToolListTokensOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.ModulePath, "module-path", "",
		"absolute path to the PKCS11 module")
}

// PKCS11ToolListKeysUrisOptions is the wrapper for `pkcs11-tool list-keys-uris` related options.
type PKCS11ToolListKeysUrisOptions struct {
	ModulePath string
	SlotID     uint
	Pin        string
}

var _ Interface = (*PKCS11ToolListKeysUrisOptions)(nil)

// AddFlags implements Interface
func (o *PKCS11ToolListKeysUrisOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.ModulePath, "module-path", "",
		"absolute path to the PKCS11 module")
	_ = cmd.Flags().SetAnnotation("module-path", cobra.BashCompFilenameExt, []string{})

	cmd.Flags().UintVar(&o.SlotID, "slot-id", 0,
		"id of the PKCS11 slot, uses 0 if empty")

	cmd.Flags().StringVar(&o.Pin, "pin", "",
		"pin of the PKCS11 slot, uses environment variable COSIGN_PKCS11_PIN if empty")
}
