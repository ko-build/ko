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

// PIVToolSetManagementKeyOptions is the wrapper for `piv-tool set-management-key` related options.
type PIVToolSetManagementKeyOptions struct {
	OldKey    string
	NewKey    string
	RandomKey bool
}

var _ Interface = (*PIVToolSetManagementKeyOptions)(nil)

// AddFlags implements Interface
func (o *PIVToolSetManagementKeyOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.OldKey, "old-key", "",
		"existing management key, uses default if empty")

	cmd.Flags().StringVar(&o.NewKey, "new-key", "",
		"new management key, uses default if empty")

	cmd.Flags().BoolVar(&o.RandomKey, "random-management-key", false,
		"if set to true, generates a new random management key and deletes it after")
}

// PIVToolSetPINOptions is the wrapper for `piv-tool set-pin` related options.
type PIVToolSetPINOptions struct {
	OldPIN string
	NewPIN string
}

var _ Interface = (*PIVToolSetPINOptions)(nil)

// AddFlags implements Interface
func (o *PIVToolSetPINOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.OldPIN, "old-pin", "",
		"existing PIN, uses default if empty")

	cmd.Flags().StringVar(&o.NewPIN, "new-pin", "",
		"new PIN, uses default if empty")
}

// PIVToolSetPUKOptions is the wrapper for `piv-tool set-puk` related options.
type PIVToolSetPUKOptions struct {
	OldPUK string
	NewPUK string
}

var _ Interface = (*PIVToolSetPUKOptions)(nil)

// AddFlags implements Interface
func (o *PIVToolSetPUKOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.OldPUK, "old-puk", "",
		"existing PUK, uses default if empty")

	cmd.Flags().StringVar(&o.NewPUK, "new-puk", "",
		"new PUK, uses default if empty")
}

// PIVToolUnblockOptions is the wrapper for `piv-tool unblock` related options.
type PIVToolUnblockOptions struct {
	PUK    string
	NewPIN string
}

var _ Interface = (*PIVToolUnblockOptions)(nil)

// AddFlags implements Interface
func (o *PIVToolUnblockOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.PUK, "puk", "",
		"existing PUK, uses default if empty")

	cmd.Flags().StringVar(&o.NewPIN, "new-PIN", "",
		"new PIN, uses default if empty")
}

// PIVToolAttestationOptions is the wrapper for `piv-tool attestation` related options.
type PIVToolAttestationOptions struct {
	Output string
	Slot   string
}

var _ Interface = (*PIVToolAttestationOptions)(nil)

// AddFlags implements Interface
func (o *PIVToolAttestationOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Output, "output", "o", "text",
		"format to output attestation information in. (text|json)")

	cmd.Flags().StringVar(&o.Slot, "slot", "",
		"Slot to use for generated key (authentication|signature|card-authentication|key-management)")
}

// PIVToolGenerateKeyOptions is the wrapper for `piv-tool generate-key` related options.
type PIVToolGenerateKeyOptions struct {
	ManagementKey string
	RandomKey     bool
	Slot          string
	PINPolicy     string
	TouchPolicy   string
}

var _ Interface = (*PIVToolGenerateKeyOptions)(nil)

// AddFlags implements Interface
func (o *PIVToolGenerateKeyOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.ManagementKey, "management-key", "",
		"management key, uses default if empty")

	cmd.Flags().BoolVar(&o.RandomKey, "random-management-key", false,
		"if set to true, generates a new random management key and deletes it after")

	cmd.Flags().StringVar(&o.Slot, "slot", "",
		"Slot to use for generated key (authentication|signature|card-authentication|key-management)")

	cmd.Flags().StringVar(&o.PINPolicy, "pin-policy", "",
		"PIN policy for slot (never|once|always)")

	cmd.Flags().StringVar(&o.TouchPolicy, "touch-policy", "",
		"Touch policy for slot (never|always|cached)")
}
