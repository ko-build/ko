//go:build !go1.18
// +build !go1.18

// Copyright 2021 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// TODO: All of this is copied from:
// https://cs.opensource.google/go/go/+/master:src/debug/buildinfo/buildinfo.go
// https://cs.opensource.google/go/go/+/master:src/runtime/debug/mod.go
// It should be replaced with runtime/buildinfo.Read on the binary file when Go 1.18 is released.

package sbom

import (
	"bytes"
	"fmt"
)

// BuildInfo represents the build information read from a Go binary.
type BuildInfo struct {
	GoVersion string         // Version of Go that produced this binary.
	Path      string         // The main package path
	Main      Module         // The module containing the main package
	Deps      []*Module      // Module dependencies
	Settings  []BuildSetting // Other information about the build.
}

// Module represents a module.
type Module struct {
	Path    string  // module path
	Version string  // module version
	Sum     string  // checksum
	Replace *Module // replaced by this module
}

// BuildSetting describes a setting that may be used to understand how the
// binary was built. For example, VCS commit and dirty status is stored here.
type BuildSetting struct {
	// Key and Value describe the build setting. They must not contain tabs
	// or newlines.
	Key, Value string
}

func (bi *BuildInfo) UnmarshalText(data []byte) (err error) {
	*bi = BuildInfo{}
	lineNum := 1
	defer func() {
		if err != nil {
			err = fmt.Errorf("could not parse Go build info: line %d: %w", lineNum, err)
		}
	}()

	var (
		pathLine  = []byte("path\t")
		modLine   = []byte("mod\t")
		depLine   = []byte("dep\t")
		repLine   = []byte("=>\t")
		buildLine = []byte("build\t")
		newline   = []byte("\n")
		tab       = []byte("\t")
	)

	readModuleLine := func(elem [][]byte) (Module, error) {
		if len(elem) != 2 && len(elem) != 3 {
			return Module{}, fmt.Errorf("expected 2 or 3 columns; got %d", len(elem))
		}
		sum := ""
		if len(elem) == 3 {
			sum = string(elem[2])
		}
		return Module{
			Path:    string(elem[0]),
			Version: string(elem[1]),
			Sum:     sum,
		}, nil
	}

	var (
		last *Module
		line []byte
		ok   bool
	)
	// Reverse of BuildInfo.String(), except for go version.
	for len(data) > 0 {
		line, data, ok = bytesCut(data, newline)
		if !ok {
			break
		}
		line = bytes.TrimPrefix(line, []byte("\t"))
		switch {
		case bytes.HasPrefix(line, pathLine):
			elem := line[len(pathLine):]
			bi.Path = string(elem)
		case bytes.HasPrefix(line, modLine):
			elem := bytes.Split(line[len(modLine):], tab)
			last = &bi.Main
			*last, err = readModuleLine(elem)
			if err != nil {
				return err
			}
		case bytes.HasPrefix(line, depLine):
			elem := bytes.Split(line[len(depLine):], tab)
			last = new(Module)
			bi.Deps = append(bi.Deps, last)
			*last, err = readModuleLine(elem)
			if err != nil {
				return err
			}
		case bytes.HasPrefix(line, repLine):
			elem := bytes.Split(line[len(repLine):], tab)
			if len(elem) != 3 {
				return fmt.Errorf("expected 3 columns for replacement; got %d", len(elem))
			}
			if last == nil {
				return fmt.Errorf("replacement with no module on previous line")
			}
			last.Replace = &Module{
				Path:    string(elem[0]),
				Version: string(elem[1]),
				Sum:     string(elem[2]),
			}
			last = nil
		case bytes.HasPrefix(line, buildLine):
			elem := bytes.Split(line[len(buildLine):], tab)
			if len(elem) != 2 {
				return fmt.Errorf("expected 2 columns for build setting; got %d", len(elem))
			}
			if len(elem[0]) == 0 {
				return fmt.Errorf("empty key")
			}
			bi.Settings = append(bi.Settings, BuildSetting{Key: string(elem[0]), Value: string(elem[1])})
		}
		lineNum++
	}
	return nil
}

// bytesCut slices s around the first instance of sep,
// returning the text before and after sep.
// The found result reports whether sep appears in s.
// If sep does not appear in s, cut returns s, nil, false.
//
// bytesCut returns slices of the original slice s, not copies.
func bytesCut(s, sep []byte) (before, after []byte, found bool) {
	if i := bytes.Index(s, sep); i >= 0 {
		return s[:i], s[i+len(sep):], true
	}
	return s, nil, false
}
