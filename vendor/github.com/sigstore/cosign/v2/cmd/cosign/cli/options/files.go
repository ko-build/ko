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
	"strings"

	"github.com/spf13/cobra"

	cremote "github.com/sigstore/cosign/v2/pkg/cosign/remote"
)

// FilesOptions is the wrapper for the files.
type FilesOptions struct {
	Files []string
}

var _ Interface = (*FilesOptions)(nil)

func (o *FilesOptions) Parse() ([]cremote.File, error) {
	fs := cremote.FilesFromFlagList(o.Files)

	// If we have multiple files, each file must have a platform.
	if len(fs) > 1 {
		for _, f := range fs {
			if f.Platform() == nil {
				return nil, fmt.Errorf("each file must include a unique platform, %s had no platform", f.Path())
			}
		}
	}

	return fs, nil
}

func (o *FilesOptions) String() string {
	return strings.Join(o.Files, ",")
}

// AddFlags implements Interface
func (o *FilesOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringSliceVarP(&o.Files, "files", "f", nil,
		"<filepath>:[platform/arch]")
	_ = cmd.Flags().SetAnnotation("files", cobra.BashCompFilenameExt, []string{})
}
