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
	"log"
	"os"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/grafana/ko/pkg/commands/options"
	"github.com/grafana/ko/pkg/publish"
	ocimutate "github.com/sigstore/cosign/pkg/oci/mutate"
	"github.com/sigstore/cosign/pkg/oci/signed"
	"github.com/spf13/cobra"
)

// addPublishLayout augments our CLI surface with publish-layout.
func addPublishLayout(topLevel *cobra.Command) {
	publishLocal := false
	var po options.PublishOptions
	cmd := &cobra.Command{
		Use:   "publish-layout APP OCI-LAYOUT",
		Short: "Publish manifest list from the given OCI image layout.",
		Long:  `This sub-command loads a manifest list from the given OCI image layout, and pushes it to the configured repository.`,
		Example: `
  # Push manifest list from OCI image layout to a Docker Registry as:
  #   ${KO_DOCKER_REPO}/<app>
  ko publish-layout example docker-images`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			idx, err := imageIndex(ctx, args[1])
			if err != nil {
				return err
			}
			name := args[0]
			if !publishLocal {
				if err := pushImageIndex(ctx, idx, name, po); err != nil {
					return err
				}
			} else {
				if err := publishToDaemon(ctx, idx, name, po); err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().StringSliceVarP(&po.Tags, "tags", "t", []string{"latest"},
		"Which tags to use for the produced image instead of the default 'latest' tag "+
			"(may not work properly with --base-import-paths or --bare).")
	cmd.Flags().BoolVarP(&po.BaseImportPaths, "base-import-paths", "B", po.BaseImportPaths,
		"Whether to use the base path without MD5 hash after KO_DOCKER_REPO (may not work properly with --tags).")
	cmd.Flags().BoolVar(&po.Bare, "bare", po.Bare,
		"Whether to just use KO_DOCKER_REPO without additional context (may not work properly with --tags).")
	cmd.Flags().BoolVarP(&publishLocal, "local", "L", publishLocal,
		"Load images into local Docker daemon.")
	topLevel.AddCommand(cmd)
}

// pushImageIndex pushes an image index to an image repository.
func pushImageIndex(ctx context.Context, idx v1.ImageIndex, name string, po options.PublishOptions) error {
	dockerRepo := os.Getenv("KO_DOCKER_REPO")
	if dockerRepo == "" {
		return fmt.Errorf("KO_DOCKER_REPO environment variable is unset")
	}

	namer := options.MakeNamer(&po)
	publisher, err := publish.NewDefault(dockerRepo,
		publish.WithUserAgent(ua()),
		publish.WithAuthFromKeychain(keychain),
		publish.WithNamer(namer),
		publish.WithTags(po.Tags),
	)
	if err != nil {
		return err
	}
	_, err = publisher.Publish(ctx, idx, name)
	return err
}

// publishToDaemon publishes an image index with the local Docker daemon.
func publishToDaemon(ctx context.Context, idx v1.ImageIndex, name string, po options.PublishOptions) error {
	dockerRepo := os.Getenv("KO_DOCKER_REPO")
	if dockerRepo == "" {
		return fmt.Errorf("KO_DOCKER_REPO environment variable is unset")
	}
	namer := options.MakeNamer(&po)
	publisher, err := publish.NewDaemon(namer, po.Tags, publish.WithLocalDomain(dockerRepo))
	if err != nil {
		return err
	}
	_, err = publisher.Publish(ctx, idx, name)
	return err
}

// imageIndex loads an image index from an OCI image layout.
func imageIndex(ctx context.Context, layoutPath string) (v1.ImageIndex, error) {
	idx, err := layout.ImageIndexFromPath(layoutPath)
	if err != nil {
		return nil, err
	}
	im, err := idx.IndexManifest()
	if err != nil {
		return nil, err
	}

	var adds []ocimutate.IndexAddendum

	// The cointained image index should have one image index within it
	d := im.Manifests[0]
	nestedIdx, err := idx.ImageIndex(d.Digest)
	if err != nil {
		return nil, err
	}

	nestedMt, err := nestedIdx.MediaType()
	if err != nil {
		return nil, err
	}
	mm, err := nestedIdx.IndexManifest()
	if err != nil {
		return nil, err
	}

	log.Printf("The image index nested within the OCI image layout has type %s and %d manifest(s)\n", nestedMt, len(mm.Manifests))
	for i, m := range mm.Manifests {
		img, err := nestedIdx.Image(m.Digest)
		if err != nil {
			return nil, err
		}

		if m.MediaType != types.OCIManifestSchema1 && m.MediaType != types.DockerManifestSchema1 && m.MediaType != types.DockerManifestSchema2 {
			return nil, fmt.Errorf("Image #%d in OCI image layout has wrong media type: %q", i+1, m.MediaType)
		}
		log.Printf("Image #%d has media type %q and platform %s/%s", i+1, m.MediaType, m.Platform.OS, m.Platform.Architecture)
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

	return ocimutate.AppendManifests(mutate.IndexMediaType(empty.Index, nestedMt), adds...), nil
}
