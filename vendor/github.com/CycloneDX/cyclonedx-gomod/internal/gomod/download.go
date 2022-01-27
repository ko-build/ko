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

	"github.com/rs/zerolog"

	"github.com/CycloneDX/cyclonedx-gomod/internal/gocmd"
)

// See https://golang.org/ref/mod#go-mod-download
type ModuleDownload struct {
	Path    string // module path
	Version string // module version
	Error   string // error loading module
	Dir     string // absolute path to cached source root directory
	Sum     string // checksum for path, version (as in go.sum)
}

func (m ModuleDownload) Coordinates() string {
	if m.Version == "" {
		return m.Path
	}

	return m.Path + "@" + m.Version
}

func Download(logger zerolog.Logger, modules []Module) ([]ModuleDownload, error) {
	var downloads []ModuleDownload
	chunks := chunkModules(modules, 20)

	for _, chunk := range chunks {
		chunkDownloads, err := downloadInternal(logger, chunk)
		if err != nil {
			return nil, err
		}

		downloads = append(downloads, chunkDownloads...)
	}

	return downloads, nil
}

func downloadInternal(logger zerolog.Logger, modules []Module) ([]ModuleDownload, error) {
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)

	coordinates := make([]string, len(modules))
	for i := range modules {
		coordinates[i] = modules[i].Coordinates()
	}

	err := gocmd.DownloadModules(logger, coordinates, stdoutBuf, stderrBuf)
	if err != nil {
		// `go mod download` will exit with code 1 if *any* of the
		// module downloads failed. Download errors are reported for
		// each module separately via the .Error field (written to STDOUT).
		//
		// If a serious error occurred that prevented `go mod download`
		// from running alltogether, it's written to STDERR.
		//
		// See https://github.com/golang/go/issues/35380
		if stderrBuf.Len() != 0 {
			return nil, fmt.Errorf(stderrBuf.String())
		}
	}

	var downloads []ModuleDownload
	jsonDecoder := json.NewDecoder(stdoutBuf)

	for {
		var download ModuleDownload
		if err := jsonDecoder.Decode(&download); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return nil, err
		}

		downloads = append(downloads, download)
	}

	return downloads, nil
}
