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

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/google/ko/pkg/commands/options"
	"github.com/google/ko/pkg/publish"
	ocimutate "github.com/sigstore/cosign/pkg/oci/mutate"
	"github.com/sigstore/cosign/pkg/oci/signed"
	"github.com/spf13/cobra"
)

// addPush augments our CLI surface with push.
func addPush(topLevel *cobra.Command) {
	publishLocal := false
	push := &cobra.Command{
		Use:   "push APP OCI-LAYOUT",
		Short: "Push manifest list from the given OCI image layout.",
		Long:  `This sub-command loads a manifest list from the given OCI image layout, and pushes it to the configured repository.`,
		Example: `
  # Push manifest list from OCI image layout to a Docker Registry as:
  #   ${KO_DOCKER_REPO}/<app>
  ko push example docker-images`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			ml, err := manifestList(ctx, args[1])
			if err != nil {
				return err
			}
			name := args[0]
			if !publishLocal {
				if err := pushManifestList(ctx, ml, name); err != nil {
					return err
				}
			} else {
				if err := publishToDaemon(ctx, ml, name); err != nil {
					return err
				}
			}
			return nil
		},
	}
	push.Flags().BoolVarP(&publishLocal, "local", "L", publishLocal,
		"Load images into local Docker daemon.")
	topLevel.AddCommand(push)
}

func pushManifestList(ctx context.Context, ml v1.ImageIndex, name string) error {
	dockerRepo := os.Getenv("KO_DOCKER_REPO")
	if dockerRepo == "" {
		return fmt.Errorf("KO_DOCKER_REPO environment variable is unset")
	}

	// TODO
	namer := options.MakeNamer(&options.PublishOptions{})
	publisher, err := publish.NewDefault(dockerRepo,
		publish.WithUserAgent(ua()),
		publish.WithAuthFromKeychain(authn.DefaultKeychain),
		publish.WithNamer(namer),
	)
	if err != nil {
		return err
	}
	_, err = publisher.Publish(ctx, ml, name)
	return err
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

		m, err := tarball.LoadManifest(func() (io.ReadCloser, error) {
			return os.Open(fpath)
		})
		if err != nil {
			return nil, fmt.Errorf("failed to load manifest for %q from %q: %w", t, fpath, err)
		}

		for _, d := range m {
			fmt.Printf("Descriptor config: %s\n", d.Config)
		}
	}

	return imgs, nil
}

func publishToDaemon(ctx context.Context, ml v1.ImageIndex, name string) error {
	// TODO
	namer := options.MakeNamer(&options.PublishOptions{})
	// TODO
	tags := []string{}
	publisher, err := publish.NewDaemon(namer, tags)
	if err != nil {
		return err
	}
	_, err = publisher.Publish(ctx, ml, name)
	return err
}

func manifestList(ctx context.Context, layoutPath string) (v1.ImageIndex, error) {
	idx, err := layout.ImageIndexFromPath(layoutPath)
	if err != nil {
		return nil, err
	}
	mt, err := idx.MediaType()
	if err != nil {
		return nil, err
	}
	im, err := idx.IndexManifest()
	if err != nil {
		return nil, err
	}

	var adds []ocimutate.IndexAddendum

	fmt.Printf("The image index has media type %s and %d image manifests\n", mt, len(im.Manifests))
	var nestedMt types.MediaType
	for _, d := range im.Manifests {
		fmt.Printf("Manifest %s with type %s\n", d.Digest.Hex, d.MediaType)
		nestedIdx, err := idx.ImageIndex(d.Digest)
		if err != nil {
			return nil, err
		}

		nestedMt, err = nestedIdx.MediaType()
		if err != nil {
			return nil, err
		}
		mm, err := nestedIdx.IndexManifest()
		if err != nil {
			return nil, err
		}

		fmt.Printf("Nested index with type %s and %d manifests\n", nestedMt, len(mm.Manifests))
		for _, m := range mm.Manifests {
			img, err := nestedIdx.Image(m.Digest)
			if err != nil {
				return nil, err
			}

			fmt.Printf("Manifest with type %s, digest %s and platform %#v\n", m.MediaType, m.Digest.Hex, m.Platform)
			fmt.Printf("Adding addendum - URLs: %#v, media type: %s, annotations: %#v, platform: %#v\n", m.URLs, m.MediaType, m.Annotations, m.Platform)
			adds = append(adds, ocimutate.IndexAddendum{
				Add: signed.Image(img),
				Descriptor: v1.Descriptor{
					URLs:        m.URLs,
					MediaType:   m.MediaType,
					Annotations: m.Annotations,
					Platform:    m.Platform,
				},
			})
		}
	}

	return ocimutate.AppendManifests(mutate.IndexMediaType(empty.Index, nestedMt), adds...), nil
}
