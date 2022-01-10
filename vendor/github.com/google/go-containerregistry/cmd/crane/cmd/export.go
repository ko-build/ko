// Copyright 2018 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/spf13/cobra"
)

// NewCmdExport creates a new cobra.Command for the export subcommand.
func NewCmdExport(options *[]crane.Option) *cobra.Command {
	return &cobra.Command{
		Use:   "export IMAGE TARBALL",
		Short: "Export contents of a remote image as a tarball",
		Example: `  # Write tarball to stdout
  crane export ubuntu -

  # Write tarball to file
  crane export ubuntu ubuntu.tar`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(_ *cobra.Command, args []string) error {
			src, dst := args[0], "-"
			if len(args) > 1 {
				dst = args[1]
			}

			f, err := openFile(dst)
			if err != nil {
				return fmt.Errorf("failed to open %s: %w", dst, err)
			}
			defer f.Close()

			img, err := crane.Pull(src, *options...)
			if err != nil {
				return fmt.Errorf("pulling %s: %w", src, err)
			}

			return crane.Export(img, f)
		},
	}
}

func openFile(s string) (*os.File, error) {
	if s == "-" {
		return os.Stdout, nil
	}
	return os.Create(s)
}
