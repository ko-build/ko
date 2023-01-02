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

// ReferenceOptions is a wrapper for image reference options.
type ReferenceOptions struct {
	TagPrefix string
}

var _ Interface = (*ReferenceOptions)(nil)

// AddFlags implements Interface
func (o *ReferenceOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.TagPrefix, "attachment-tag-prefix", "", "optional custom prefix to use for attached image tags. Attachment images are tagged as: `[AttachmentTagPrefix]sha256-[TargetImageDigest].[AttachmentName]`")
}
