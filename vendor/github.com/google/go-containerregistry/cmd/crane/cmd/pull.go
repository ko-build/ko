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
	"errors"
	"fmt"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/cache"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/match"
	"github.com/google/go-containerregistry/pkg/v1/partial"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/spf13/cobra"
)

var defaultPlatform = &v1.Platform{
	OS:           "linux",
	Architecture: "amd64",
}

// NewCmdPull creates a new cobra.Command for the pull subcommand.
func NewCmdPull(options *[]crane.Option) *cobra.Command {
	var cachePath, format string

	cmd := &cobra.Command{
		Use:   "pull IMAGE TARBALL",
		Short: "Pull remote images by reference and store their contents locally",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			imageMap := map[string]v1.Image{}
			indexMap := map[string]v1.ImageIndex{}
			srcList, path := args[:len(args)-1], args[len(args)-1]
			for _, src := range srcList {
				o := crane.GetOptions(*options...)
				ref, err := name.ParseReference(src, o.Name...)
				if err != nil {
					return fmt.Errorf("parsing reference %q: %w", src, err)
				}

				rmt, err := remote.Get(ref, o.Remote...)
				if err != nil {
					return err
				}

				// If we're writing an index to a layout and --platform hasn't been set,
				// pull the entire index, not just a child image.
				if format == "oci" && rmt.MediaType.IsIndex() && o.Platform == nil {
					idx, err := rmt.ImageIndex()
					if err != nil {
						return err
					}
					indexMap[src] = idx
					continue
				}

				var img v1.Image
				if rmt.MediaType.IsImage() {
					if o.Platform != nil && !match.FuzzyPlatforms(*o.Platform)(rmt.Descriptor) {
						return errors.New("image does not match requested platform")
					}
					img, err = rmt.Image()
					if err != nil {
						return err
					}
				} else if rmt.MediaType.IsIndex() {
					idx, err := rmt.ImageIndex()
					if err != nil {
						return err
					}
					if o.Platform == nil {
						o.Platform = defaultPlatform
					}
					imgs, err := partial.FindImages(idx, match.FuzzyPlatforms(*o.Platform))
					if err != nil {
						return err
					}
					if len(imgs) == 0 {
						return fmt.Errorf("no matching images for platform %q", o.Platform)
					} else if len(imgs) > 1 {
						// TODO: print the matching platform strings.
						return fmt.Errorf("multiple matching images for platform %q", o.Platform)
					}
					img = imgs[0]
				} else {
					return fmt.Errorf("unknown media type: %s", rmt.MediaType)
				}

				if cachePath != "" {
					img = cache.Image(img, cache.NewFilesystemCache(cachePath))
				}
				imageMap[src] = img
			}

			switch format {
			case "tarball":
				if err := crane.MultiSave(imageMap, path); err != nil {
					return fmt.Errorf("saving tarball %s: %w", path, err)
				}
			case "legacy":
				if err := crane.MultiSaveLegacy(imageMap, path); err != nil {
					return fmt.Errorf("saving legacy tarball %s: %w", path, err)
				}
			case "oci":
				if err := crane.MultiSaveOCI(imageMap, path); err != nil {
					return fmt.Errorf("saving oci image layout %s: %w", path, err)
				}

				// crane.MultiSaveOCI doesn't support index, so just append these at the end.
				p, err := layout.FromPath(path)
				if err != nil {
					return err
				}
				for ref, idx := range indexMap {
					anns := map[string]string{
						"dev.ggcr.image.name": ref,
					}
					if err := p.AppendIndex(idx, layout.WithAnnotations(anns)); err != nil {
						return err
					}
				}
			default:
				return fmt.Errorf("unexpected --format: %q (valid values are: tarball, legacy, and oci)", format)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&cachePath, "cache_path", "c", "", "Path to cache image layers")
	cmd.Flags().StringVar(&format, "format", "tarball", fmt.Sprintf("Format in which to save images (%q, %q, or %q)", "tarball", "legacy", "oci"))

	return cmd
}
