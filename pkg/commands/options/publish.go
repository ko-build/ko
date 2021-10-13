/*
Copyright 2018 Google LLC All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package options

import (
	"crypto/md5" //nolint: gosec // No strong cryptography needed.
	"encoding/hex"
	"os"
	"path"
	"strings"

	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/ko/pkg/publish"
	"github.com/spf13/cobra"
)

// PublishOptions encapsulates options when publishing.
type PublishOptions struct {
	// DockerRepo configures the destination image repository.
	// In normal ko usage, this is populated with the value of $KO_DOCKER_REPO.
	DockerRepo string

	// LocalDomain overrides the default domain for images loaded into the local Docker daemon. Use with Local=true.
	LocalDomain string

	// UserAgent enables overriding the default value of the `User-Agent` HTTP
	// request header used when pushing the built image to an image registry.
	UserAgent string

	// DockerClient enables overriding the default docker client when embedding
	// ko as a module in other tools.
	// If left as the zero value, ko uses github.com/docker/docker/client.FromEnv
	DockerClient daemon.Client

	Tags []string
	// TagOnly resolves images into tag-only references.
	TagOnly bool

	// Push publishes images to a registry.
	Push bool

	// Local publishes images to a local docker daemon.
	Local            bool
	InsecureRegistry bool

	OCILayoutPath string
	TarballFile   string

	// PreserveImportPaths preserves the full import path after KO_DOCKER_REPO.
	PreserveImportPaths bool
	// BaseImportPaths uses the base path without MD5 hash after KO_DOCKER_REPO.
	BaseImportPaths bool
	// Bare uses a tag on the KO_DOCKER_REPO without anything additional.
	Bare bool
	// ImageNameSeparator is used to create the image name from base repo and
	// specific import path.
	ImageNameSeparator string

	// ImageNamer can be used to pass a custom image name function. When given
	// PreserveImportPaths, BaseImportPaths, Bare, and ImageNameSeparator has
	// no effect.
	ImageNamer publish.Namer
}

func AddPublishArg(cmd *cobra.Command, po *PublishOptions) {
	// Set DockerRepo from the KO_DOCKER_REPO envionment variable.
	// See https://github.com/google/ko/pull/351 for flag discussion.
	if dockerRepo, exists := os.LookupEnv("KO_DOCKER_REPO"); exists {
		po.DockerRepo = dockerRepo
	}

	cmd.Flags().StringSliceVarP(&po.Tags, "tags", "t", []string{"latest"},
		"Which tags to use for the produced image instead of the default 'latest' tag "+
			"(may not work properly with --base-import-paths or --bare).")
	cmd.Flags().BoolVar(&po.TagOnly, "tag-only", false,
		"Include tags but not digests in resolved image references. Useful when digests are not preserved when images are repopulated.")

	cmd.Flags().BoolVar(&po.Push, "push", true, "Push images to KO_DOCKER_REPO")

	cmd.Flags().BoolVarP(&po.Local, "local", "L", po.Local,
		"Load into images to local docker daemon.")
	cmd.Flags().BoolVar(&po.InsecureRegistry, "insecure-registry", po.InsecureRegistry,
		"Whether to skip TLS verification on the registry")

	cmd.Flags().StringVar(&po.OCILayoutPath, "oci-layout-path", "", "Path to save the OCI image layout of the built images")
	cmd.Flags().StringVar(&po.TarballFile, "tarball", "", "File to save images tarballs")

	cmd.Flags().BoolVarP(&po.PreserveImportPaths, "preserve-import-paths", "P", po.PreserveImportPaths,
		"Whether to preserve the full import path after KO_DOCKER_REPO.")
	cmd.Flags().BoolVarP(&po.BaseImportPaths, "base-import-paths", "B", po.BaseImportPaths,
		"Whether to use the base path without MD5 hash after KO_DOCKER_REPO (may not work properly with --tags).")
	cmd.Flags().BoolVar(&po.Bare, "bare", po.Bare,
		"Whether to just use KO_DOCKER_REPO without additional context (may not work properly with --tags).")
	cmd.Flags().StringVar(&po.ImageNameSeparator, "image-name-separator", defaultImageNameSeparator,
		"A separator to be used to concatenate KO_DOCKER_REPO with specific image name.")
}

const defaultImageNameSeparator = "/"

func pathJoin(sep string, paths ...string) string {
	if sep == "" {
		sep = defaultImageNameSeparator
	}
	for i, e := range paths {
		if e != "" {
			return path.Clean(strings.Join(paths[i:], sep))
		}
	}
	return ""
}

func (po *PublishOptions) packageWithMD5(base, importpath string) string {
	hasher := md5.New() //nolint: gosec // No strong cryptography needed.
	hasher.Write([]byte(importpath))
	return pathJoin(po.ImageNameSeparator,
		base, path.Base(importpath)+"-"+hex.EncodeToString(hasher.Sum(nil)))
}

func (po *PublishOptions) preserveImportPath(base, importpath string) string {
	return pathJoin(po.ImageNameSeparator, base, importpath)
}

func (po *PublishOptions) baseImportPaths(base, importpath string) string {
	return pathJoin(po.ImageNameSeparator, base, path.Base(importpath))
}

func bareDockerRepo(base, _ string) string {
	return base
}

func MakeNamer(po *PublishOptions) publish.Namer {
	if po.ImageNamer != nil {
		return po.ImageNamer
	} else if po.PreserveImportPaths {
		return po.preserveImportPath
	} else if po.BaseImportPaths {
		return po.baseImportPaths
	} else if po.Bare {
		return bareDockerRepo
	}
	return po.packageWithMD5
}
