// This file is part of CycloneDX GoMod
//
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
// Copyright (c) OWASP Foundation. All Rights Reserved.

package pkg

import (
	"fmt"
	"path/filepath"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/rs/zerolog"

	"github.com/CycloneDX/cyclonedx-gomod/internal/gomod"
	fileConv "github.com/CycloneDX/cyclonedx-gomod/internal/sbom/convert/file"
)

type Option func(zerolog.Logger, gomod.Package, gomod.Module, *cdx.Component) error

func WithFiles(enabled bool) Option {
	return func(logger zerolog.Logger, pkg gomod.Package, module gomod.Module, component *cdx.Component) error {
		if !enabled {
			return nil
		}

		var files []string
		files = append(files, pkg.GoFiles...)
		files = append(files, pkg.CgoFiles...)
		files = append(files, pkg.CFiles...)
		files = append(files, pkg.CXXFiles...)
		files = append(files, pkg.MFiles...)
		files = append(files, pkg.HFiles...)
		files = append(files, pkg.FFiles...)
		files = append(files, pkg.SFiles...)
		files = append(files, pkg.SwigFiles...)
		files = append(files, pkg.SwigCXXFiles...)
		files = append(files, pkg.SysoFiles...)
		files = append(files, pkg.EmbedFiles...)

		var fileComponents []cdx.Component

		for _, file := range files {
			fileComponent, err := fileConv.ToComponent(
				logger,
				filepath.Join(pkg.Dir, file),
				file,
				fileConv.WithHashes(
					cdx.HashAlgoMD5,
					cdx.HashAlgoSHA1,
					cdx.HashAlgoSHA256,
					cdx.HashAlgoSHA384,
					cdx.HashAlgoSHA512,
				),
			)
			if err != nil {
				return err
			}

			fileComponents = append(fileComponents, *fileComponent)
		}

		if len(fileComponents) > 0 {
			component.Components = &fileComponents
		}

		return nil
	}
}

func ToComponent(logger zerolog.Logger, pkg gomod.Package, module gomod.Module, options ...Option) (*cdx.Component, error) {
	logger.Debug().
		Str("package", pkg.ImportPath).
		Msg("converting package to component")

	component := cdx.Component{
		Type:       cdx.ComponentTypeLibrary,
		Name:       pkg.ImportPath,
		Version:    module.Version,
		PackageURL: fmt.Sprintf("pkg:golang/%s@%s?type=package", pkg.ImportPath, module.Version),
	}

	for _, option := range options {
		if err := option(logger, pkg, module, &component); err != nil {
			return nil, err
		}
	}

	return &component, nil
}
