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

package commands

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/spf13/cobra"
)

// addPush augments our CLI surface with push.
func addPush(topLevel *cobra.Command) {
	push := &cobra.Command{
		Use:   "push TARBALL...",
		Short: "Push manifest list from the given tarballs.",
		Long:  `This sub-command loads container images from the given tarballs, and pushes a manifest list combining them to the configured registry.`,
		Example: `
  # Push manifest list from images in tarballs to a Docker Registry as:
  #   ${KO_DOCKER_REPO}/<package name>-<hash of import path>
  ko push images.tar.gz`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			err := pushManifestList(ctx, args)
			if err != nil {
				return fmt.Errorf("failed to push manifest list: %w", err)
			}
			return nil
		},
	}
	topLevel.AddCommand(push)
}

func pushManifestList(ctx context.Context, args []string) error {
	var images []v1.Image
	for _, fpath := range args {
		is, err := loadImages(fpath)
		if err != nil {
			return err
		}

		for _, img := range is {
			h, err := img.ConfigName()
			if err != nil {
				return err
			}
			d, err := img.Digest()
			if err != nil {
				return err
			}
			fmt.Printf("Loaded image %s, digest %s, from %q\n", h.Hex, d.Hex, fpath)
		}
		images = append(images, is...)
	}

	fmt.Printf("Loaded %d image(s)\n", len(images))

	// TODO: Make manifest list

	// TODO: Push manifest list to registry

	return nil
}

func loadTags(fpath string) ([]*name.Tag, error) {
	m, err := tarball.LoadManifest(func() (io.ReadCloser, error) {
		return os.Open(fpath)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to load manifest from %q: %w", fpath, err)
	}

	var tags []*name.Tag
	for _, d := range m {
		tag, err := name.NewTag(d.RepoTags[0])
		if err != nil {
			return nil, err
		}

		tags = append(tags, &tag)
	}

	return tags, nil
}

// loadImages loads images from a tarball.
func loadImages(fpath string) ([]v1.Image, error) {
	tags, err := loadTags(fpath)
	if err != nil {
		return nil, err
	}

	var imgs []v1.Image
	for _, t := range tags {
		img, err := tarball.ImageFromPath(fpath, t)
		if err != nil {
			return nil, fmt.Errorf("failed to load image %q from %q: %w", t, fpath, err)
		}

		imgs = append(imgs, img)
	}

	return imgs, nil
}
