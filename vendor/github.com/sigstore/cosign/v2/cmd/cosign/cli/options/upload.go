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

// UploadBlobOptions is the top level wrapper for the `upload blob` command.
type UploadBlobOptions struct {
	ContentType string
	Files       FilesOptions
	Registry    RegistryOptions
	Annotations map[string]string
}

var _ Interface = (*UploadBlobOptions)(nil)

// AddFlags implements Interface
func (o *UploadBlobOptions) AddFlags(cmd *cobra.Command) {
	o.Registry.AddFlags(cmd)
	o.Files.AddFlags(cmd)

	cmd.Flags().StringVar(&o.ContentType, "ct", "",
		"content type to set")
	cmd.Flags().StringToStringVarP(&o.Annotations, "annotation", "a", nil,
		"annotations to set")
}

// UploadWASMOptions is the top level wrapper for the `upload wasm` command.
type UploadWASMOptions struct {
	File     string
	Registry RegistryOptions
}

var _ Interface = (*UploadWASMOptions)(nil)

// AddFlags implements Interface
func (o *UploadWASMOptions) AddFlags(cmd *cobra.Command) {
	o.Registry.AddFlags(cmd)

	cmd.Flags().StringVarP(&o.File, "file", "f", "",
		"path to the wasm file to upload")
	_ = cmd.Flags().SetAnnotation("file", cobra.BashCompFilenameExt, []string{})
	_ = cmd.MarkFlagRequired("file")
}
