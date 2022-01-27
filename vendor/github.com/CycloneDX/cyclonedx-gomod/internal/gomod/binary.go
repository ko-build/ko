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

// BuildInfo represents the build information read from a Go binary.
// Adapted from https://github.com/golang/go/blob/931d80ec17374e52dbc5f9f63120f8deb80b355d/src/runtime/debug/mod.go#L41
type BuildInfo struct {
	GoVersion string            // Version of Go that produced this binary.
	Path      string            // The main package path
	Main      *Module           // The module containing the main package
	Deps      []Module          // Module dependencies
	Settings  map[string]string // Other information about the build.
}

func LoadBuildInfo(logger zerolog.Logger, binaryPath string) (*BuildInfo, error) {
	buf := new(bytes.Buffer)
	err := gocmd.LoadBuildInfo(logger, binaryPath, buf)
	if err != nil {
		return nil, err
	}

	buildInfo, err := parseBuildInfo(binaryPath, buf)
	if err != nil {
		return nil, err
	}

	return &buildInfo, nil
}

func parseBuildInfo(binaryPath string, reader io.Reader) (bi BuildInfo, err error) {
	moduleIndex := 0
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		switch fields[0] {
		case binaryPath + ":":
			var gv string
			gv, err = gocmd.ParseVersion(line)
			if err != nil {
				return
			} else {
				bi.GoVersion = gv
			}
		case "path": // Path of main package of main module
			bi.Path = fields[1]
		case "mod": // Main module
			bi.Main = &Module{
				Path:    fields[1],
				Version: fields[2],
				Main:    true,
			}
		case "dep": // Dependency module
			module := Module{
				Path:    fields[1],
				Version: fields[2],
			}
			if len(fields) == 4 {
				// Hash won't be available when the module is replaced
				module.Sum = fields[3]
			}
			bi.Deps = append(bi.Deps, module)
			moduleIndex += 1
		case "=>": // Replacement
			module := Module{
				Path:    fields[1],
				Version: fields[2],
			}
			if len(fields) == 4 {
				module.Sum = fields[3]
			}
			bi.Deps[moduleIndex-1].Replace = &module
		case "build": // Build settings (Go 1.18+)
			kv := strings.SplitN(fields[1], "=", 2)
			if len(kv) == 2 {
				if bi.Settings == nil {
					bi.Settings = make(map[string]string)
				}
				bi.Settings[kv[0]] = kv[1]
			}
		}
	}

	return
}
