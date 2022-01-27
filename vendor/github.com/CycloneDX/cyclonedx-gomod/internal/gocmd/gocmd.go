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

package gocmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"

	"github.com/rs/zerolog"
)

// GetVersion returns the version of Go in the environment.
func GetVersion(logger zerolog.Logger) (string, error) {
	buf := new(bytes.Buffer)
	err := executeGoCommand(logger, []string{"version"}, withStdout(buf))
	if err != nil {
		return "", err
	}

	return ParseVersion(buf.String())
}

var versionRegex = regexp.MustCompile(`^.*(?P<version>\bgo[^\s$]+)`)

// ParseVersion attempts to locate a Go version in a given string.
// Output of `go version` is not easily parseable in all cases.
// See https://github.com/golang/go/issues/21207
func ParseVersion(s string) (string, error) {
	matches := versionRegex.FindStringSubmatch(s)
	if len(matches) != 2 {
		return "", fmt.Errorf("regex did not match")
	}

	return matches[1], nil
}

// GetEnv executes `go env -json` and returns the result as a map.
// See https://pkg.go.dev/cmd/go#hdr-Print_Go_environment_information.
func GetEnv(logger zerolog.Logger) (map[string]string, error) {
	buf := new(bytes.Buffer)
	err := executeGoCommand(logger, []string{"env", "-json"}, withStdout(buf))
	if err != nil {
		return nil, err
	}

	var env map[string]string
	err = json.NewDecoder(buf).Decode(&env)
	if err != nil {
		return nil, err
	}

	return env, nil
}

// ListModule executes `go list -json -m` and writes the output to a given writer.
// See https://golang.org/ref/mod#go-list-m
func ListModule(logger zerolog.Logger, moduleDir string, writer io.Writer) error {
	return executeGoCommand(logger, []string{"list", "-mod", "readonly", "-json", "-m"}, withDir(moduleDir), withStdout(writer))
}

// ListModules executes `go list -json -m all` and writes the output to a given writer.
// See https://golang.org/ref/mod#go-list-m
func ListModules(logger zerolog.Logger, moduleDir string, writer io.Writer) error {
	return executeGoCommand(logger, []string{"list", "-mod", "readonly", "-json", "-m", "all"}, withDir(moduleDir), withStdout(writer))
}

// ListPackage executes `go list -json -e <PATTERN>` and writes the output to a given writer.
// See https://golang.org/cmd/go/#hdr-List_packages_or_modules.
func ListPackage(logger zerolog.Logger, moduleDir, packagePattern string, writer io.Writer) error {
	return executeGoCommand(logger, []string{"list", "-json", "-e", packagePattern},
		withDir(moduleDir),
		withStdout(writer),
		withStderr(newLoggerWriter(logger))) // reports download status
}

// ListPackages executes `go list -deps -json <PATTERN>` and writes the output to a given writer.
// See https://golang.org/cmd/go/#hdr-List_packages_or_modules.
func ListPackages(logger zerolog.Logger, moduleDir, packagePattern string, writer io.Writer) error {
	return executeGoCommand(logger, []string{"list", "-deps", "-json", packagePattern},
		withDir(moduleDir),
		withStdout(writer),
		withStderr(newLoggerWriter(logger)), // reports download status
	)
}

// ListVendoredModules executes `go mod vendor -v` and writes the output to a given writer.
// See https://golang.org/ref/mod#go-mod-vendor.
func ListVendoredModules(logger zerolog.Logger, moduleDir string, writer io.Writer) error {
	return executeGoCommand(logger, []string{"mod", "vendor", "-v", "-e"}, withDir(moduleDir), withStderr(writer))
}

// GetModuleGraph executes `go mod graph` and writes the output to a given writer.
// See https://golang.org/ref/mod#go-mod-graph.
func GetModuleGraph(logger zerolog.Logger, moduleDir string, writer io.Writer) error {
	return executeGoCommand(logger, []string{"mod", "graph"}, withDir(moduleDir), withStdout(writer))
}

// ModWhy executes `go mod why -m -vendor` and writes the output to a given writer.
// See https://golang.org/ref/mod#go-mod-why.
func ModWhy(logger zerolog.Logger, moduleDir string, modules []string, writer io.Writer) error {
	return executeGoCommand(logger,
		append([]string{"mod", "why", "-m", "-vendor"}, modules...),
		withDir(moduleDir),
		withStdout(writer),
		withStderr(newLoggerWriter(logger)), // reports download status
	)
}

// LoadBuildInfo executes `go version -m` and writes the output to a given writer.
// See https://golang.org/ref/mod#go-version-m.
func LoadBuildInfo(logger zerolog.Logger, binaryPath string, writer io.Writer) error {
	return executeGoCommand(logger, []string{"version", "-m", binaryPath}, withStdout(writer))
}

// DownloadModules executes `go mod download -json` and writes the output to the given writers.
// See https://golang.org/ref/mod#go-mod-download.
func DownloadModules(logger zerolog.Logger, modules []string, stdout, stderr io.Writer) error {
	return executeGoCommand(logger,
		append([]string{"mod", "download", "-json"}, modules...),
		withDir(os.TempDir()), // `mod download` modifies go.sum when executed in moduleDir
		withStdout(stdout),
		withStderr(newLoggerWriter(logger)), // reports download status
	)
}

type commandOption func(*exec.Cmd)

func withDir(dir string) commandOption {
	return func(c *exec.Cmd) {
		c.Dir = dir
	}
}

func withStderr(writer io.Writer) commandOption {
	return func(c *exec.Cmd) {
		c.Stderr = writer
	}
}

func withStdout(writer io.Writer) commandOption {
	return func(c *exec.Cmd) {
		c.Stdout = writer
	}
}

func executeGoCommand(logger zerolog.Logger, args []string, options ...commandOption) error {
	cmd := exec.Command("go", args...)

	for _, option := range options {
		option(cmd)
	}

	logger.Debug().
		Str("cmd", cmd.String()).
		Str("dir", cmd.Dir).
		Msg("executing command")

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("command `%s` failed: %w", cmd.String(), err)
	}

	return nil
}

// loggerWriter is a workaround to be able to log go output
// with zerolog loggers. zerolog.Logger does implement the io.Writer
// interface, but it ignores log levels per default.
type loggerWriter struct {
	logger zerolog.Logger
}

func newLoggerWriter(logger zerolog.Logger) io.Writer {
	return &loggerWriter{
		logger: logger,
	}
}

func (s loggerWriter) Write(p []byte) (n int, err error) {
	s.logger.Debug().Msg(string(bytes.TrimSpace(p)))
	return len(p), nil
}
