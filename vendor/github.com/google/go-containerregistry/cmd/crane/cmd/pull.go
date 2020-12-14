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
	"log"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/v1/cache"
	"github.com/spf13/cobra"
)

// NewCmdPull creates a new cobra.Command for the pull subcommand.
func NewCmdPull(options *[]crane.Option) *cobra.Command {
	var cachePath, format string

	cmd := &cobra.Command{
		Use:   "pull IMAGE TARBALL",
		Short: "Pull a remote image by reference and store its contents in a tarball",
		Args:  cobra.ExactArgs(2),
		Run: func(_ *cobra.Command, args []string) {
			src, path := args[0], args[1]
			img, err := crane.Pull(src, *options...)
			if err != nil {
				log.Fatal(err)
			}
			if cachePath != "" {
				img = cache.Image(img, cache.NewFilesystemCache(cachePath))
			}

			switch format {
			case "tarball":
				if err := crane.Save(img, src, path); err != nil {
					log.Fatalf("saving tarball %s: %v", path, err)
				}
			case "legacy":
				if err := crane.SaveLegacy(img, src, path); err != nil {
					log.Fatalf("saving legacy tarball %s: %v", path, err)
				}
			case "oci":
				if err := crane.SaveOCI(img, path); err != nil {
					log.Fatalf("saving oci image layout %s: %v", path, err)
				}
			default:
				log.Fatalf("unexpected --format: %q (valid values are: tarball, legacy, and oci)", format)
			}
		},
	}
	cmd.Flags().StringVarP(&cachePath, "cache_path", "c", "", "Path to cache image layers")
	cmd.Flags().StringVar(&format, "format", "tarball", fmt.Sprintf("Format in which to save images (%q, %q, or %q)", "tarball", "legacy", "oci"))

	return cmd
}
