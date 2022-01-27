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
	"fmt"
	"path/filepath"

	"github.com/rs/zerolog"

	"github.com/CycloneDX/cyclonedx-gomod/internal/gocmd"
)

// StdlibModulePath defines the path used for Go's standard library module.
const StdlibModulePath = "std"

// LoadStdlibModule loads the standard library module.
func LoadStdlibModule(logger zerolog.Logger) (*Module, error) {
	env, err := gocmd.GetEnv(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to get go env: %w", err)
	}

	goroot, ok := env["GOROOT"]
	if !ok {
		return nil, fmt.Errorf("failed to determine GOROOT")
	}

	module, err := LoadModule(logger, filepath.Join(goroot, "src"))
	if err != nil {
		return nil, fmt.Errorf("failed to load stdlib module: %w", err)
	}

	module.Version, err = gocmd.GetVersion(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to determine go version: %w", err)
	}

	module.Local = true
	module.Main = false

	return module, nil
}
