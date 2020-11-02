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

package options

import (
	"crypto/md5" //nolint: gosec // No strong cryptography needed.
	"encoding/hex"
	"path/filepath"

	"github.com/google/ko/pkg/publish"
	"github.com/spf13/cobra"
)

// PublishOptions encapsulates options when publishing.
type PublishOptions struct {
	Tags []string

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
	// Naked uses a tag on the KO_DOCKER_REPO without anything additional.
	Naked bool
}

func AddPublishArg(cmd *cobra.Command, po *PublishOptions) {
	cmd.Flags().StringSliceVarP(&po.Tags, "tags", "t", []string{"latest"},
		"Which tags to use for the produced image instead of the default 'latest' tag.")

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
		"Whether to use the base path without MD5 hash after KO_DOCKER_REPO.")
	cmd.Flags().BoolVarP(&po.Naked, "naked", "N", po.Naked,
		"Whether to just use KO_DOCKER_REPO without additional context.")
}

func packageWithMD5(base, importpath string) string {
	hasher := md5.New() //nolint: gosec // No strong cryptography needed.
	hasher.Write([]byte(importpath))
	return filepath.Join(base, filepath.Base(importpath)+"-"+hex.EncodeToString(hasher.Sum(nil)))
}

func preserveImportPath(base, importpath string) string {
	return filepath.Join(base, importpath)
}

func baseImportPaths(base, importpath string) string {
	return filepath.Join(base, filepath.Base(importpath))
}

func nakedDockerRepo(base, _ string) string {
	return base
}

func MakeNamer(po *PublishOptions) publish.Namer {
	if po.PreserveImportPaths {
		return preserveImportPath
	} else if po.BaseImportPaths {
		return baseImportPaths
	} else if po.Naked {
		return nakedDockerRepo
	}
	return packageWithMD5
}
