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

package file

import (
	"fmt"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/rs/zerolog"

	"github.com/CycloneDX/cyclonedx-gomod/internal/sbom"
)

type Option func(logger zerolog.Logger, absFilePath, relFilePath string, component *cdx.Component) error

func WithHashes(algos ...cdx.HashAlgorithm) Option {
	return func(logger zerolog.Logger, abs, _ string, component *cdx.Component) error {
		hashes, err := sbom.CalculateFileHashes(logger, abs, algos...)
		if err != nil {
			return err
		}

		component.Hashes = &hashes
		return nil
	}
}

func ToComponent(logger zerolog.Logger, absFilePath, relFilePath string, options ...Option) (*cdx.Component, error) {
	logger.Debug().
		Str("file", absFilePath).
		Msg("converting file to component")

	component := cdx.Component{
		Type:  cdx.ComponentTypeFile,
		Name:  relFilePath,
		Scope: cdx.ScopeRequired,
	}

	hashes, err := sbom.CalculateFileHashes(logger, absFilePath, cdx.HashAlgoSHA1)
	if err != nil {
		return nil, err
	}
	component.Version = fmt.Sprintf("v0.0.0-%s", hashes[0].Value[:12])

	for _, option := range options {
		err = option(logger, absFilePath, relFilePath, &component)
		if err != nil {
			return nil, err
		}
	}

	return &component, nil
}
