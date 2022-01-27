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

package gomod

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rs/zerolog"

	"github.com/CycloneDX/cyclonedx-gomod/internal/gocmd"
)

// See https://golang.org/cmd/go/#hdr-List_packages_or_modules
type Package struct {
	Dir        string  // directory containing package sources
	ImportPath string  // import path of package in dir
	Name       string  // package name
	Standard   bool    // is this package part of the standard Go library?
	Module     *Module // info about package's containing module, if any (can be nil)

	GoFiles      []string // .go source files (excluding CgoFiles, TestGoFiles, XTestGoFiles)
	CgoFiles     []string // .go source files that import "C"
	CFiles       []string // .c source files
	CXXFiles     []string // .cc, .cxx and .cpp source files
	MFiles       []string // .m source files
	HFiles       []string // .h, .hh, .hpp and .hxx source files
	FFiles       []string // .f, .F, .for and .f90 Fortran source files
	SFiles       []string // .s source files
	SwigFiles    []string // .swig files
	SwigCXXFiles []string // .swigcxx files
	SysoFiles    []string // .syso object files to add to archive
	EmbedFiles   []string // files matched by EmbedPatterns

	Error *PackageError // error loading package
}

type PackageError struct {
	Err string
}

func (pe PackageError) Error() string {
	return pe.Err
}

func LoadPackage(logger zerolog.Logger, moduleDir, packagePattern string) (*Package, error) {
	logger.Debug().
		Str("moduleDir", moduleDir).
		Str("packagePattern", packagePattern).
		Msg("loading package")

	buf := new(bytes.Buffer)
	err := gocmd.ListPackage(logger, moduleDir, toRelativePackagePath(packagePattern), buf)
	if err != nil {
		return nil, err
	}

	var pkg Package
	err = json.NewDecoder(buf).Decode(&pkg)
	if err != nil {
		return nil, err
	}

	if pkg.Error != nil {
		return nil, pkg.Error
	}

	return &pkg, nil
}

func LoadModulesFromPackages(logger zerolog.Logger, moduleDir, packagePattern string) ([]Module, error) {
	logger.Debug().
		Str("moduleDir", moduleDir).
		Msg("loading modules")

	if !IsModule(moduleDir) {
		return nil, ErrNoModule
	}

	buf := new(bytes.Buffer)
	err := gocmd.ListPackages(logger, moduleDir, toRelativePackagePath(packagePattern), buf)
	if err != nil {
		return nil, fmt.Errorf("failed to list packages for pattern \"%s\": %w", packagePattern, err)
	}

	pkgMap, err := parsePackages(logger, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to parse `go list` output: %w", err)
	}

	modules, err := convertPackagesToModules(logger, moduleDir, pkgMap)
	if err != nil {
		return nil, fmt.Errorf("failed to convert packages to modules: %w", err)
	}

	err = ResolveLocalReplacements(logger, moduleDir, modules)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve local replacements: %w", err)
	}

	sortModules(modules)

	return modules, nil
}

// parsePackages parses the output of `go list -json`.
// The keys of the returned map are module coordinates (path@version).
func parsePackages(logger zerolog.Logger, reader io.Reader) (map[string][]Package, error) {
	pkgsMap := make(map[string][]Package)
	jsonDecoder := json.NewDecoder(reader)

	for {
		var pkg Package
		if err := jsonDecoder.Decode(&pkg); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return nil, err
		}

		if pkg.Error != nil {
			return nil, fmt.Errorf("failed to load package: %w", pkg.Error)
		}

		var coordinates string
		if pkg.Standard {
			coordinates = StdlibModulePath
		} else if pkg.Module == nil {
			logger.Debug().
				Str("package", pkg.ImportPath).
				Str("reason", "no associated module").
				Msg("skipping package")
			continue
		} else {
			coordinates = pkg.Module.Coordinates()
		}

		pkgs, ok := pkgsMap[coordinates]
		if !ok {
			pkgsMap[coordinates] = []Package{pkg}
		} else {
			pkgsMap[coordinates] = append(pkgs, pkg)
		}
	}

	return pkgsMap, nil
}

func convertPackagesToModules(logger zerolog.Logger, mainModuleDir string, pkgsMap map[string][]Package) ([]Module, error) {
	modules := make([]Module, 0, len(pkgsMap))
	isVendoring := IsVendoring(mainModuleDir)

	for coordinates, pkgs := range pkgsMap {
		if len(pkgs) == 0 {
			continue
		}

		var (
			module *Module
			err    error
		)

		if coordinates == StdlibModulePath {
			module, err = LoadStdlibModule(logger)
			if err != nil {
				return nil, fmt.Errorf("failed to load stdlib module: %w", err)
			}
		} else {
			module = pkgs[0].Module
		}

		if module == nil {
			// Shouldn't ever happen, because packages without module are not collected to pkgsMap.
			// We do the nil check anyway to make linters happy. :)
			return nil, fmt.Errorf("no module is associated with package %s", pkgs[0].ImportPath)
		}

		if !module.Main && module.Path != StdlibModulePath && isVendoring {
			module.Vendored = true
			vendorPath := filepath.Join(mainModuleDir, "vendor", module.Path)

			if module.Replace != nil {
				module.Replace.Vendored = true
				module.Replace.Dir = vendorPath
			} else {
				module.Dir = vendorPath
			}
		}

		modules = append(modules, *module)
	}

	for i := range modules {
		var pkgs []Package
		if modules[i].Path == StdlibModulePath {
			pkgs = pkgsMap[StdlibModulePath]
		} else {
			pkgs = pkgsMap[modules[i].Coordinates()]
		}

		for j := range pkgs {
			pkgs[j].Module = nil // we don't need this anymore
			modules[i].Packages = append(modules[i].Packages, pkgs[j])
		}

		sortPackages(modules[i].Packages)
	}

	return modules, nil
}

// sortPackages sorts a given Package slice ascending by import path.
func sortPackages(pkgs []Package) {
	sort.Slice(pkgs, func(i, j int) bool {
		return pkgs[i].ImportPath < pkgs[j].ImportPath
	})
}

// toRelativePackagePath ensures that Go will interpret the given packagePattern
// as relative package path. This is done by prefixing it with "./" if it isn't already.
// See also: `go help packages`
func toRelativePackagePath(packagePattern string) string {
	packagePattern = filepath.ToSlash(packagePattern)
	if !strings.HasPrefix(packagePattern, "./") {
		packagePattern = "./" + packagePattern
	}
	return packagePattern
}
