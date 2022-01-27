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
	"golang.org/x/mod/semver"
	"golang.org/x/mod/sumdb/dirhash"

	"github.com/CycloneDX/cyclonedx-gomod/internal/gocmd"
	"github.com/CycloneDX/cyclonedx-gomod/internal/util"
)

// See https://golang.org/ref/mod#go-list-m
type Module struct {
	Path     string  // module path
	Version  string  // module version
	Replace  *Module // replaced by this module
	Main     bool    // is this the main module?
	Indirect bool    // is this module only an indirect dependency of main module?
	Dir      string  // directory holding files for this module, if any

	Dependencies []*Module `json:"-"` // modules this module depends on
	Local        bool      `json:"-"` // is this a local module?
	Packages     []Package `json:"-"` // packages in this module
	Sum          string    `json:"-"` // checksum for path, version (as in go.sum)
	TestOnly     bool      `json:"-"` // is this module only required for tests?
	Vendored     bool      `json:"-"` // is this a vendored module?
}

func (m Module) Coordinates() string {
	if m.Version == "" {
		return m.Path
	}

	return m.Path + "@" + m.Version
}

func (m Module) Hash() (string, error) {
	h1, err := dirhash.HashDir(m.Dir, m.Coordinates(), dirhash.Hash1)
	if err != nil {
		return "", err
	}

	return h1, nil
}

func (m Module) PackageURL() string {
	return fmt.Sprintf("pkg:golang/%s?type=module", m.Coordinates())
}

// IsModule determines whether dir is a Go module.
func IsModule(dir string) bool {
	return util.FileExists(filepath.Join(dir, "go.mod"))
}

// ErrNoModule indicates that a given path is not a valid Go module
var ErrNoModule = errors.New("not a go module")

func LoadModule(logger zerolog.Logger, moduleDir string) (*Module, error) {
	logger.Debug().
		Str("moduleDir", moduleDir).
		Msg("loading module")

	buf := new(bytes.Buffer)
	err := gocmd.ListModule(logger, moduleDir, buf)
	if err != nil {
		return nil, fmt.Errorf("listing module failed: %w", err)
	}

	var module Module
	err = json.NewDecoder(buf).Decode(&module)
	if err != nil {
		return nil, fmt.Errorf("decoding module info failed: %w", err)
	}

	return &module, nil
}

func LoadModules(logger zerolog.Logger, moduleDir string, includeTest bool) ([]Module, error) {
	logger.Debug().
		Str("moduleDir", moduleDir).
		Bool("includeTest", includeTest).
		Msg("loading modules")

	if !IsModule(moduleDir) {
		return nil, ErrNoModule
	}

	buf := new(bytes.Buffer)
	err := gocmd.ListModules(logger, moduleDir, buf)
	if err != nil {
		return nil, fmt.Errorf("listing modules failed: %w", err)
	}

	modules, err := parseModules(buf)
	if err != nil {
		return nil, fmt.Errorf("parsing modules failed: %w", err)
	}

	modules, err = FilterModules(logger, moduleDir, modules, includeTest)
	if err != nil {
		return nil, fmt.Errorf("filtering modules failed: %w", err)
	}

	err = ResolveLocalReplacements(logger, moduleDir, modules)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve local replacements: %w", err)
	}

	sortModules(modules)

	return modules, nil
}

// parseModules parses the output of `go list -json -m` into a Module slice.
func parseModules(reader io.Reader) ([]Module, error) {
	modules := make([]Module, 0)
	jsonDecoder := json.NewDecoder(reader)

	// Output is not a JSON array, so we have to parse one object after another
	for {
		var mod Module
		if err := jsonDecoder.Decode(&mod); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		modules = append(modules, mod)
	}
	return modules, nil
}

// sortModules sorts a given Module slice ascending by path.
// Main modules take precedence, so that they will represent the first elements of the sorted slice.
// If the path of two modules are equal, they'll be compared by their semantic version instead.
func sortModules(modules []Module) {
	sort.Slice(modules, func(i, j int) bool {
		if modules[i].Main && !modules[j].Main {
			return true
		} else if !modules[i].Main && modules[j].Main {
			return false
		}

		if modules[i].Path == modules[j].Path {
			return semver.Compare(modules[i].Version, modules[j].Version) == -1
		}

		return modules[i].Path < modules[j].Path
	})
}

// ResolveLocalReplacements tries to resolve paths and versions for local replacement modules.
func ResolveLocalReplacements(logger zerolog.Logger, mainModuleDir string, modules []Module) error {
	for i, module := range modules {
		if module.Replace == nil {
			// Only replacements can be local
			continue
		}

		if !strings.HasPrefix(module.Replace.Path, "./") &&
			!strings.HasPrefix(module.Replace.Path, "../") {
			// According to the specification, local paths must start with either one of these prefixes.
			continue
		}

		var localModuleDir string
		if filepath.IsAbs(module.Replace.Path) {
			localModuleDir = modules[i].Replace.Path
		} else { // Replacement path is relative to main module
			localModuleDir = filepath.Join(mainModuleDir, modules[i].Replace.Path)
		}

		if !IsModule(localModuleDir) {
			logger.Warn().
				Str("moduleDir", localModuleDir).
				Msg("local replacement does not exist or is not a module")
			continue
		}

		err := resolveLocalReplacement(logger, localModuleDir, modules[i].Replace)
		if err != nil {
			return fmt.Errorf("resolving local module %s failed: %w", module.Replace.Coordinates(), err)
		}
	}

	return nil
}

func resolveLocalReplacement(logger zerolog.Logger, localModuleDir string, module *Module) error {
	logger.Debug().
		Str("moduleDir", localModuleDir).
		Msg("resolving local replacement module")

	localModule, err := LoadModule(logger, localModuleDir)
	if err != nil {
		return err
	}

	module.Path = localModule.Path
	module.Local = true

	// Try to resolve the version. Only works when module.Dir is a Git repo.
	if module.Version == "" {
		version, err := GetModuleVersion(logger, module.Dir)
		if err == nil {
			module.Version = version
		} else {
			// We don't fail with an error here, because our possibilities are limited.
			// module.Dir may be a Mercurial repo or just a normal directory, in which case we
			// cannot detect versions reliably right now.
			logger.Warn().
				Err(err).
				Str("module", module.Path).
				Str("moduleDir", localModuleDir).
				Msg("failed to resolve version of local module")
		}
	}

	return nil
}

func chunkModules(modules []Module, chunkSize int) [][]Module {
	var chunks [][]Module

	for i := 0; i < len(modules); i += chunkSize {
		j := i + chunkSize

		if j > len(modules) {
			j = len(modules)
		}

		chunks = append(chunks, modules[i:j])
	}

	return chunks
}
