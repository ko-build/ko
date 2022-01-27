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

package module

import (
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/rs/zerolog"

	"github.com/CycloneDX/cyclonedx-gomod/internal/gomod"
	"github.com/CycloneDX/cyclonedx-gomod/internal/license"
	pkgConv "github.com/CycloneDX/cyclonedx-gomod/internal/sbom/convert/pkg"
)

type Option func(zerolog.Logger, gomod.Module, *cdx.Component) error

// WithLicenses attempts to resolve licenses for the module and attach them
// to the component's license evidence.
func WithLicenses(enabled bool) Option {
	return func(logger zerolog.Logger, module gomod.Module, component *cdx.Component) error {
		if !enabled {
			return nil
		}

		if module.Dir == "" {
			logger.Warn().
				Str("module", module.Coordinates()).
				Str("reason", "module not in cache").
				Msg("can't resolve module license")
			return nil
		}

		resolvedLicenses, err := license.Resolve(logger, module)

		if err == nil {
			componentLicenses := make(cdx.Licenses, len(resolvedLicenses))
			for i := range resolvedLicenses {
				componentLicenses[i] = cdx.LicenseChoice{License: &resolvedLicenses[i]}
			}

			component.Evidence = &cdx.Evidence{
				Licenses: &componentLicenses,
			}
		} else {
			if errors.Is(err, license.ErrNoLicenseDetected) {
				logger.Warn().Str("module", module.Coordinates()).Msg("no license detected")
				return nil
			}

			return fmt.Errorf("failed to resolve license for %s: %v", module.Coordinates(), err)
		}

		return nil
	}
}

// WithComponentType overrides the type of the component.
func WithComponentType(ctype cdx.ComponentType) Option {
	return func(_ zerolog.Logger, _ gomod.Module, component *cdx.Component) error {
		component.Type = ctype
		return nil
	}
}

func WithModuleHashes() Option {
	return func(logger zerolog.Logger, module gomod.Module, component *cdx.Component) error {
		if module.Main {
			// We currently don't have an accurate way of hashing the main module, as it may contain
			// files that are .gitignore'd and thus not part of the hashes in Go's sumdb.
			logger.Debug().Str("module", module.Coordinates()).Msg("not calculating hash for main module")
			return nil
		}

		if module.Vendored {
			// Go's vendoring mechanism doesn't copy all files that make up a module to the vendor dir.
			// Hashing vendored modules thus won't result in the expected hash, probably causing more
			// confusion than anything else.
			logger.Debug().Str("module", module.Coordinates()).Msg("not calculating hash for vendored module")
			return nil
		}

		if module.Path == gomod.StdlibModulePath {
			// There are no module hashes published for the standard library.
			logger.Debug().Str("module", module.Coordinates()).Msg("not calculating hash for stdlib module")
			return nil
		}

		logger.Debug().Str("module", module.Coordinates()).Msg("calculating module hash")
		h1, err := module.Hash()
		if err != nil {
			return fmt.Errorf("failed to calculate module hash: %w", err)
		}

		h1Bytes, err := base64.StdEncoding.DecodeString(h1[3:])
		if err != nil {
			return fmt.Errorf("failed to base64 decode module hash: %w", err)
		}

		component.Hashes = &[]cdx.Hash{
			{Algorithm: cdx.HashAlgoSHA256, Value: fmt.Sprintf("%x", h1Bytes)},
		}

		return nil
	}
}

func WithPackages(enabled bool, options ...pkgConv.Option) Option {
	return func(logger zerolog.Logger, module gomod.Module, component *cdx.Component) error {
		if !enabled {
			return nil
		}

		var pkgComponents []cdx.Component
		for i := range module.Packages {
			pkgComponent, err := pkgConv.ToComponent(logger, module.Packages[i], module, options...)
			if err != nil {
				return fmt.Errorf("failed to convert package: %w", err)
			}
			pkgComponents = append(pkgComponents, *pkgComponent)
		}

		if len(pkgComponents) > 0 {
			component.Components = &pkgComponents
		}

		return nil
	}
}

// WithScope overrides the scope of the component.
func WithScope(scope cdx.Scope) Option {
	return func(_ zerolog.Logger, _ gomod.Module, component *cdx.Component) error {
		component.Scope = scope
		return nil
	}
}

// WithTestScope overrides the scope of the component,
// if the corresponding module has the TestOnly flag set.
func WithTestScope(scope cdx.Scope) Option {
	return func(_ zerolog.Logger, module gomod.Module, component *cdx.Component) error {
		if module.TestOnly {
			component.Scope = scope
		}

		return nil
	}
}

// ToComponent converts a gomod.Module to a CycloneDX component.
// The component can be further customized using options, before it's returned.
func ToComponent(logger zerolog.Logger, module gomod.Module, options ...Option) (*cdx.Component, error) {
	if module.Replace != nil {
		return ToComponent(logger, *module.Replace, options...)
	}

	logger.Debug().
		Str("module", module.Coordinates()).
		Msg("converting module to component")

	component := cdx.Component{
		BOMRef:     module.PackageURL(),
		Type:       cdx.ComponentTypeLibrary,
		Name:       module.Path,
		Version:    module.Version,
		PackageURL: module.PackageURL(),
	}

	if !module.Main { // Main component can't have a scope
		if module.TestOnly {
			component.Scope = cdx.ScopeOptional
		} else {
			component.Scope = cdx.ScopeRequired
		}
	}

	if module.Sum != "" && strings.HasPrefix(module.Sum, "h1:") {
		h1Bytes, err := base64.StdEncoding.DecodeString(module.Sum[3:])
		if err != nil {
			return nil, err
		}

		component.Hashes = &[]cdx.Hash{
			{
				Algorithm: cdx.HashAlgoSHA256,
				Value:     fmt.Sprintf("%x", h1Bytes),
			},
		}
	}

	vcsURL := resolveVCSURL(module.Path)
	if vcsURL != "" {
		component.ExternalReferences = &[]cdx.ExternalReference{
			{
				Type: cdx.ERTypeVCS,
				URL:  vcsURL,
			},
		}
	}

	for _, option := range options {
		if err := option(logger, module, &component); err != nil {
			return nil, err
		}
	}

	return &component, nil
}

// ToComponents converts a slice of gomod.Module to a slice of CycloneDX components.
func ToComponents(logger zerolog.Logger, modules []gomod.Module, options ...Option) ([]cdx.Component, error) {
	components := make([]cdx.Component, 0, len(modules))

	for i := range modules {
		component, err := ToComponent(logger, modules[i], options...)
		if err != nil {
			return nil, err
		}
		components = append(components, *component)
	}

	return components, nil
}

var (
	// By convention, modules with a major version equal to or above v2
	// have it as suffix in their module path.
	vcsUrlMajorVersionSuffixRegex = regexp.MustCompile(`(/v[\d]+)$`)

	// gopkg.in with user segment
	// Example: gopkg.in/user/pkg.v3 -> github.com/user/pkg
	vcsUrlGoPkgInRegexWithUser = regexp.MustCompile(`^gopkg\.in/([^/]+)/([^.]+)\..*$`)

	// gopkg.in without user segment
	// Example: gopkg.in/pkg.v3 -> github.com/go-pkg/pkg
	vcsUrlGoPkgInRegexWithoutUser = regexp.MustCompile(`^gopkg\.in/([^.]+)\..*$`)
)

func resolveVCSURL(modulePath string) string {
	switch {
	case strings.HasPrefix(modulePath, "github.com/"):
		return "https://" + vcsUrlMajorVersionSuffixRegex.ReplaceAllString(modulePath, "")
	case vcsUrlGoPkgInRegexWithUser.MatchString(modulePath):
		return "https://" + vcsUrlGoPkgInRegexWithUser.ReplaceAllString(modulePath, "github.com/$1/$2")
	case vcsUrlGoPkgInRegexWithoutUser.MatchString(modulePath):
		return "https://" + vcsUrlGoPkgInRegexWithoutUser.ReplaceAllString(modulePath, "github.com/go-$1/$1")
	}

	return ""
}
