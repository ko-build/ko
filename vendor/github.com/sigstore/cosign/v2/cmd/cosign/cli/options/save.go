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

// SaveOptions is the top level wrapper for the load command.
type SaveOptions struct {
	Directory string
}

var _ Interface = (*SaveOptions)(nil)

// AddFlags implements Interface
func (o *SaveOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Directory, "dir", "",
		"path to dir where the signed image should be stored on disk")
	_ = cmd.Flags().SetAnnotation("dir", cobra.BashCompSubdirsInDir, []string{})
	_ = cmd.MarkFlagRequired("dir")
}
