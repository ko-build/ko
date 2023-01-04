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

// GenerateKeyPairOptions is the top level wrapper for the generate-key-pair command.
type GenerateKeyPairOptions struct {
	// KMS Key Management Service
	KMS string
}

var _ Interface = (*GenerateKeyPairOptions)(nil)

// AddFlags implements Interface
func (o *GenerateKeyPairOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.KMS, "kms", "",
		"create key pair in KMS service to use for signing")
}
