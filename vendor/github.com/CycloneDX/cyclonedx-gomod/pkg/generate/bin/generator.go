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

package bin

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/mod/module"

	"github.com/CycloneDX/cyclonedx-gomod/internal/gomod"
	"github.com/CycloneDX/cyclonedx-gomod/internal/sbom"
	modConv "github.com/CycloneDX/cyclonedx-gomod/internal/sbom/convert/module"
	"github.com/CycloneDX/cyclonedx-gomod/pkg/generate"
)

type generator struct {
	logger zerolog.Logger

	binaryPath      string
	detectLicenses  bool
	includeStdlib   bool
	versionOverride string
}

// NewGenerator returns a generator that is capable of generating BOMs from Go module binaries.
func NewGenerator(binaryPath string, opts ...Option) (generate.Generator, error) {
	g := generator{
		logger:     log.Logger,
		binaryPath: binaryPath,
	}

	var err error
	for _, opt := range opts {
		if err = opt(&g); err != nil {
			return nil, err
		}
	}

	return &g, nil
}

// Generate implements the generate.Generator interface.
func (g generator) Generate() (*cdx.BOM, error) {
	bi, err := gomod.LoadBuildInfo(g.logger, g.binaryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load build info: %w", err)
	} else if bi.Main == nil {
		return nil, fmt.Errorf("failed to parse any modules from %s", g.binaryPath)
	}

	modules := append([]gomod.Module{*bi.Main}, bi.Deps...)

	if g.includeStdlib {
		modules = append(modules, gomod.Module{
			Path:    gomod.StdlibModulePath,
			Version: bi.GoVersion,
		})
	}

	if g.versionOverride != "" {
		modules[0].Version = g.versionOverride
	} else if modules[0].Version == "(devel)" && len(bi.Settings) > 0 {
		g.logger.Debug().Msg("building pseudo version from buildinfo")
		modules[0].Version, err = buildPseudoVersion(bi)
		if err != nil {
			g.logger.Warn().Err(err).Msg("failed to build pseudo version from buildinfo")
		}
	}

	if g.detectLicenses {
		// Before we can resolve licenses, we have to download the modules first
		err = g.downloadModules(modules)
		if err != nil {
			return nil, fmt.Errorf("failed to download modules: %w", err)
		}
	}

	// Make all modules a direct dependency of the main module
	for i := 1; i < len(modules); i++ {
		modules[0].Dependencies = append(modules[0].Dependencies, &modules[i])
	}

	main, err := modConv.ToComponent(g.logger, modules[0],
		modConv.WithComponentType(cdx.ComponentTypeApplication),
		modConv.WithLicenses(g.detectLicenses))
	if err != nil {
		return nil, fmt.Errorf("failed to convert main module: %w", err)
	}
	components, err := modConv.ToComponents(g.logger, modules[1:],
		modConv.WithLicenses(g.detectLicenses))
	if err != nil {
		return nil, fmt.Errorf("failed to convert modules: %w", err)
	}
	dependencies := sbom.BuildDependencyGraph(modules)
	compositions := buildCompositions(main, components)

	binaryProperties, err := g.buildBinaryProperties(g.binaryPath, bi)
	if err != nil {
		return nil, fmt.Errorf("failed to create binary properties")
	}

	bom := cdx.NewBOM()
	bom.Metadata = &cdx.Metadata{
		Component:  main,
		Properties: &binaryProperties,
	}
	bom.Components = &components
	bom.Dependencies = &dependencies
	bom.Compositions = &compositions

	g.includeAppPathInMainComponentPURL(bi, bom)

	return bom, nil
}

// buildPseudoVersion builds a pseudo version for the main module.
// Requires that the binary was built with Go 1.18+ and the build
// settings include VCS information.
//
// Because major version and previous version are not known,
// this operation will always produce a v0.0.0-TIME-REF version.
func buildPseudoVersion(bi *gomod.BuildInfo) (string, error) {
	vcsRev, ok := bi.Settings["vcs.revision"]
	if !ok {
		return "", fmt.Errorf("no vcs.revision buildinfo")
	}
	if len(vcsRev) > 12 {
		vcsRev = vcsRev[:12]
	}
	vcsTimeStr, ok := bi.Settings["vcs.time"]
	if !ok {
		return "", fmt.Errorf("no vcs.time buildinfo")
	}
	vcsTime, err := time.Parse(time.RFC3339, vcsTimeStr)
	if err != nil {
		return "", err
	}

	return module.PseudoVersion("", "", vcsTime, vcsRev), nil
}

func (g generator) buildBinaryProperties(binaryPath string, bi *gomod.BuildInfo) ([]cdx.Property, error) {
	properties := []cdx.Property{
		sbom.NewProperty("binary:name", filepath.Base(binaryPath)),
		sbom.NewProperty("build:env:GOVERSION", bi.GoVersion),
	}

	if len(bi.Settings) > 0 {
		addProperty := func(settingKey, property string) {
			if setting, ok := bi.Settings[settingKey]; ok {
				properties = append(properties, sbom.NewProperty(property, setting))
			}
		}

		addProperty("CGO_ENABLED", "build:env:CGO_ENABLED")
		addProperty("GOARCH", "build:env:GOARCH")
		addProperty("GOOS", "build:env:GOOS")
		addProperty("-compiler", "build:compiler")
		addProperty("vcs", "build:vcs")
		addProperty("vcs.revision", "build:vcs:revision")
		addProperty("vcs.time", "build:vcs:time")
		addProperty("vcs.modified", "build:vcs:modified")

		if tags, ok := bi.Settings["-tags"]; ok {
			for _, tag := range strings.Split(tags, ",") {
				properties = append(properties, sbom.NewProperty("build:tag", tag))
			}
		}
	}

	binaryHashes, err := sbom.CalculateFileHashes(g.logger, binaryPath,
		cdx.HashAlgoMD5, cdx.HashAlgoSHA1, cdx.HashAlgoSHA256, cdx.HashAlgoSHA384, cdx.HashAlgoSHA512)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate binary hashes: %w", err)
	}
	for _, hash := range binaryHashes {
		properties = append(properties, sbom.NewProperty(fmt.Sprintf("binary:hash:%s", hash.Algorithm), hash.Value))
	}

	sbom.SortProperties(properties)

	return properties, nil
}

func buildCompositions(mainComponent *cdx.Component, components []cdx.Component) []cdx.Composition {
	compositions := make([]cdx.Composition, 0)

	// We know all components that the main component directly or indirectly depends on,
	// thus the dependencies of it are considered complete.
	compositions = append(compositions, cdx.Composition{
		Aggregate: cdx.CompositionAggregateComplete,
		Dependencies: &[]cdx.BOMReference{
			cdx.BOMReference(mainComponent.BOMRef),
		},
	})

	// The exact relationships between the dependencies are unknown
	dependencyRefs := make([]cdx.BOMReference, 0, len(components))
	for _, component := range components {
		dependencyRefs = append(dependencyRefs, cdx.BOMReference(component.BOMRef))
	}
	compositions = append(compositions, cdx.Composition{
		Aggregate:    cdx.CompositionAggregateUnknown,
		Dependencies: &dependencyRefs,
	})

	return compositions
}

func (g generator) downloadModules(modules []gomod.Module) error {
	modulesToDownload := make([]gomod.Module, 0)
	for i := range modules {
		if modules[i].Path == gomod.StdlibModulePath {
			continue // We can't download the stdlib
		}

		// When modules are replaced, only download the replacement.
		if modules[i].Replace != nil {
			modulesToDownload = append(modulesToDownload, *modules[i].Replace)
		} else {
			modulesToDownload = append(modulesToDownload, modules[i])
		}
	}

	downloads, err := gomod.Download(g.logger, modulesToDownload)
	if err != nil {
		return err
	}

	for i, download := range downloads {
		if download.Error != "" {
			g.logger.Warn().
				Str("module", download.Coordinates()).
				Str("reason", download.Error).
				Msg("module download failed")
			continue
		}

		mm := matchModule(modules, download.Coordinates())
		if mm == nil {
			g.logger.Warn().
				Str("module", download.Coordinates()).
				Msg("downloaded module not found")
			continue
		}

		// Check that the hash of the downloaded module matches
		// the one found in the binary. We want to report the version
		// for the *exact* module version or nothing at all.
		if mm.Sum != "" && mm.Sum != download.Sum {
			g.logger.Warn().
				Str("binaryHash", mm.Sum).
				Str("downloadHash", download.Sum).
				Str("module", download.Coordinates()).
				Msg("module hash mismatch")
			continue
		}

		g.logger.Debug().
			Str("module", download.Coordinates()).
			Msg("module downloaded")

		mm.Dir = downloads[i].Dir
	}

	return nil
}

func matchModule(modules []gomod.Module, coordinates string) *gomod.Module {
	for i, m := range modules {
		if m.Replace != nil && coordinates == m.Replace.Coordinates() {
			return modules[i].Replace
		}
		if coordinates == m.Coordinates() {
			return &modules[i]
		}
	}

	return nil
}

func (g generator) includeAppPathInMainComponentPURL(bi *gomod.BuildInfo, bom *cdx.BOM) {
	if bi.Path != bi.Main.Path && strings.HasPrefix(bi.Path, bi.Main.Path) {
		subpath := strings.TrimPrefix(bi.Path, bi.Main.Path)
		subpath = strings.TrimPrefix(subpath, "/")

		oldPURL := bom.Metadata.Component.PackageURL
		newPURL := oldPURL + "#" + subpath

		// Update PURL of main component
		bom.Metadata.Component.BOMRef = newPURL
		bom.Metadata.Component.PackageURL = newPURL

		// Update PURL in dependency graph
		if bom.Dependencies != nil {
			for i, dep := range *bom.Dependencies {
				if dep.Ref == oldPURL {
					(*bom.Dependencies)[i].Ref = newPURL
					break
				}
			}
		}

		// Update PURL in compositions
		if bom.Compositions != nil {
			for i := range *bom.Compositions {
				if (*bom.Compositions)[i].Assemblies != nil {
					for j, assembly := range *(*bom.Compositions)[i].Assemblies {
						if string(assembly) == oldPURL {
							(*(*bom.Compositions)[i].Assemblies)[j] = cdx.BOMReference(newPURL)
						}
					}
				}

				if (*bom.Compositions)[i].Dependencies != nil {
					for j, dependency := range *(*bom.Compositions)[i].Dependencies {
						if string(dependency) == oldPURL {
							(*(*bom.Compositions)[i].Dependencies)[j] = cdx.BOMReference(newPURL)
						}
					}
				}
			}
		}
	}
}
