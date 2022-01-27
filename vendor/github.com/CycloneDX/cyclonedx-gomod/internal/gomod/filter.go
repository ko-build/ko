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
	"io"
	"strings"

	"github.com/rs/zerolog"

	"github.com/CycloneDX/cyclonedx-gomod/internal/gocmd"
)

// FilterModules queries `go mod why` with all provided modules to determine whether or not
// they're required by the main module. Modules required by the main module are returned in
// a new slice.
//
// Unless includeTest is true, test-only dependencies are not included in the returned slice.
// Test-only modules will have the TestOnly field set to true.
//
// Note that this method doesn't work when replacements have already been applied to the module slice.
// Consider a go.mod file containing the following lines:
//
// 		require golang.org/x/crypto v0.0.0-xxx-xxx
//		replace golang.org/x/crypto => github.com/ProtonMail/go-crypto v0.0.0-xxx-xxx
//
// Querying `go mod why -m` with `golang.org/x/crypto` yields the expected result, querying it with
// `github.com/ProtonMail/go-crypto` will always yield `(main module does not need github.com/ProtonMail/go-crypto)`.
// See:
//   - https://github.com/golang/go/issues/30720
//   - https://github.com/golang/go/issues/26904
func FilterModules(logger zerolog.Logger, moduleDir string, modules []Module, includeTest bool) ([]Module, error) {
	logger.Debug().
		Str("moduleDir", moduleDir).
		Int("moduleCount", len(modules)).
		Bool("includeTest", includeTest).
		Msg("filtering modules")

	buf := new(bytes.Buffer)
	filtered := make([]Module, 0)
	chunks := chunkModules(modules, 20)

	for _, chunk := range chunks {
		paths := make([]string, len(chunk))
		for i := range chunk {
			paths[i] = chunk[i].Path
		}

		if err := gocmd.ModWhy(logger, moduleDir, paths, buf); err != nil {
			return nil, err
		}

		for modPath, modPkgs := range parseModWhy(buf) {
			if len(modPkgs) == 0 {
				logger.Debug().
					Str("module", modPath).
					Str("reason", "not needed").
					Msg("filtering module")
				continue
			}

			// If the shortest package path contains test nodes, this is a test-only dependency.
			testOnly := false
			for _, pkg := range modPkgs {
				if strings.HasSuffix(pkg, ".test") {
					testOnly = true
					break
				}
			}
			if !includeTest && testOnly {
				logger.Debug().
					Str("module", modPath).
					Str("reason", "test only").
					Msg("filtering module")
				continue
			}

			for i := range chunk {
				if chunk[i].Path == modPath {
					mod := chunk[i]
					mod.TestOnly = testOnly
					filtered = append(filtered, mod)
				}
			}
		}

		buf.Reset()
	}

	return filtered, nil
}

// parseModWhy parses the output of `go mod why`,
// populating a map with module paths as keys and a list of packages as values.
func parseModWhy(reader io.Reader) map[string][]string {
	modPkgs := make(map[string][]string)
	currentModPath := ""

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "(main module does not need ") {
			continue
		}

		if strings.HasPrefix(line, "#") {
			currentModPath = strings.TrimSpace(strings.TrimPrefix(line, "#"))
			modPkgs[currentModPath] = make([]string, 0)
			continue
		}

		modPkgs[currentModPath] = append(modPkgs[currentModPath], line)
	}

	return modPkgs
}
