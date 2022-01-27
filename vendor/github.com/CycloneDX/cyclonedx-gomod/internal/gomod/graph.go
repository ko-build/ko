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
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/rs/zerolog"
	"golang.org/x/mod/semver"

	"github.com/CycloneDX/cyclonedx-gomod/internal/gocmd"
)

func ApplyModuleGraph(logger zerolog.Logger, moduleDir string, modules []Module) error {
	logger.Debug().
		Str("moduleDir", moduleDir).
		Int("moduleCount", len(modules)).
		Msg("applying module graph")

	buf := new(bytes.Buffer)
	err := gocmd.GetModuleGraph(logger, moduleDir, buf)
	if err != nil {
		return err
	}

	err = parseModuleGraph(logger, buf, modules)
	if err != nil {
		return err
	}

	for i := range modules {
		sortDependencies(modules[i].Dependencies)
	}

	return nil
}

// parseModuleGraph parses the output of `go mod graph` and populates
// the Dependencies field of a given Module slice.
//
// The Module slice is expected to contain only "effective" modules,
// with only a single version per module, as provided by `go list -m` or `go list -deps`.
func parseModuleGraph(logger zerolog.Logger, reader io.Reader, modules []Module) error {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) != 2 {
			return fmt.Errorf("expected two fields per line, but got %d: %s", len(fields), line)
		}

		// The module graph contains dependency relationships for multiple versions of a module.
		// When identifying the ACTUAL dependant, we search for it in strict mode (versions must match).
		dependant := findModule(modules, fields[0], true)
		if dependant == nil {
			continue
		}

		// The identified module may depend on an older version of its dependency.
		// Due to Go's minimal version selection, that version may not be present in
		// the effective modules slice. Hence, we search for the dependency in non-strict mode.
		dependency := findModule(modules, fields[1], false)
		if dependency == nil {
			logger.Debug().
				Str("dependant", dependant.Coordinates()).
				Str("dependency", fields[1]).
				Str("reason", "dependency not in list of selected modules").
				Msg("skipping graph edge")
			continue
		}

		if dependant.Main && dependency.Indirect {
			logger.Debug().
				Str("dependant", dependant.Coordinates()).
				Str("dependency", dependency.Coordinates()).
				Str("reason", "indirect dependency").
				Msg("skipping graph edge")
			continue
		}

		if dependant.Dependencies == nil {
			dependant.Dependencies = []*Module{dependency}
		} else {
			dependant.Dependencies = append(dependant.Dependencies, dependency)
		}
	}

	return nil
}

func findModule(modules []Module, coordinates string, strict bool) *Module {
	for i := range modules {
		if coordinates == modules[i].Coordinates() || (!strict && strings.HasPrefix(coordinates, modules[i].Path+"@")) {
			if modules[i].Replace != nil {
				return modules[i].Replace
			}
			return &modules[i]
		}
	}
	return nil
}

// sortDependencies sorts a given Module pointer slice ascendingly by path.
// If the path of two modules are equal, they'll be compared by their semantic version instead.
func sortDependencies(dependencies []*Module) {
	sort.Slice(dependencies, func(i, j int) bool {
		if dependencies[i].Path == dependencies[j].Path {
			return semver.Compare(dependencies[i].Version, dependencies[j].Version) == -1
		}

		return dependencies[i].Path < dependencies[j].Path
	})
}
