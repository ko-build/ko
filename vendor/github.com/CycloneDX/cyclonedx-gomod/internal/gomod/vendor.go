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
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"

	"github.com/CycloneDX/cyclonedx-gomod/internal/gocmd"
	"github.com/CycloneDX/cyclonedx-gomod/internal/util"
)

// IsVendoring determines whether of not the module at moduleDir is vendoring its dependencies.
func IsVendoring(moduleDir string) bool {
	return util.FileExists(filepath.Join(moduleDir, "vendor", "modules.txt"))
}

var ErrNotVendoring = errors.New("the module is not vendoring its dependencies")

func GetVendoredModules(logger zerolog.Logger, moduleDir string, includeTest bool) ([]Module, error) {
	if !IsModule(moduleDir) {
		return nil, ErrNoModule
	}
	if !IsVendoring(moduleDir) {
		return nil, ErrNotVendoring
	}

	logger.Debug().
		Str("moduleDir", moduleDir).
		Bool("includeTest", includeTest).
		Msg("loading vendored modules")

	buf := new(bytes.Buffer)
	err := gocmd.ListVendoredModules(logger, moduleDir, buf)
	if err != nil {
		return nil, fmt.Errorf("listing vendored modules failed: %w", err)
	}

	modules, err := parseVendoredModules(moduleDir, buf)
	if err != nil {
		return nil, fmt.Errorf("parsing vendored modules failed: %w", err)
	}

	modules, err = FilterModules(logger, moduleDir, modules, includeTest)
	if err != nil {
		return nil, fmt.Errorf("filtering modules failed: %w", err)
	}

	err = ResolveLocalReplacements(logger, moduleDir, modules)
	if err != nil {
		return nil, fmt.Errorf("resolving local modules failed: %w", err)
	}

	// Main module is not included in vendored module list, so we have to get it separately
	mainModule, err := LoadModule(logger, moduleDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get main module: %w", err)
	}

	modules = append(modules, *mainModule)

	sortModules(modules)

	return modules, nil
}

// parseVendoredModules parses the output of `go mod vendor -v` into a Module slice.
func parseVendoredModules(mainModulePath string, reader io.Reader) ([]Module, error) {
	modules := make([]Module, 0)
	modulesSeen := make(map[string]struct{})

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "# ") {
			continue
		}

		fields := strings.Fields(strings.TrimPrefix(line, "# "))

		// Replacements may be specified as
		//   Path [Version] => Path [Version]
		arrowIndex := util.StringsIndexOf(fields, "=>")

		var module Module

		if arrowIndex == -1 {
			if len(fields) != 2 {
				return nil, fmt.Errorf("expected two fields per line, but got %d: %s", len(fields), line)
			}

			module = Module{
				Path:     fields[0],
				Version:  fields[1],
				Dir:      filepath.Join(mainModulePath, "vendor", fields[0]),
				Vendored: true,
			}
		} else {
			pathParent := fields[0]
			versionParent := ""
			if arrowIndex == 2 {
				versionParent = fields[1]
			}

			pathReplacement := fields[arrowIndex+1]
			versionReplacement := ""
			if len(fields) == arrowIndex+3 {
				versionReplacement = fields[arrowIndex+2]
			}

			module = Module{
				Path:    pathParent,
				Version: versionParent,
				Replace: &Module{
					Path:     pathReplacement,
					Version:  versionReplacement,
					Dir:      filepath.Join(mainModulePath, "vendor", pathParent), // Replacements are copied to their parent's dir
					Vendored: true,
				},
			}
		}

		// Go will append all replacement constraints again at the very
		// bottom of `go mod vendor`'s output. This would cause duplicate
		// modules for us, which we prevent using this cheap deduplication.
		_, seen := modulesSeen[module.Path]
		if !seen {
			modules = append(modules, module)
			modulesSeen[module.Path] = struct{}{}
		}
	}

	return modules, nil
}
