/*
Copyright 2018 Matt Moore

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package graph

import (
	gb "go/build"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

// PackageFilter is the type of functions that determine whether to omit a
// particular Go package from the dependency graph.
type PackageFilter func(*gb.Package) bool

// OmitVendor implements PackageFilter to exclude packages in the vendor directory.
func OmitVendor(pkg *gb.Package) bool {
	return strings.Contains(pkg.ImportPath, "/vendor/")
}

// FileFilter is the type of functions that determine whether to omit a particular
// file from triggering a package-level event.
type FileFilter func(string) bool

// OmitNonGo implements FileFilter to exclude non-Go files from triggering package
// change notifications.
func OmitNonGo(path string) bool {
	return filepath.Ext(path) != ".go"
}

// OmitTests implements FileFilter to exclude Go test files from triggering package
// change notifications.
func OmitTest(path string) bool {
	return strings.HasSuffix(path, "_test.go")
}

// Option is used to mutate the underlying graph implementation during construction.
// Since the implementation's type is private this may only be implemented from within
// this package.
type Option func(*manager) error

// WithWorkDir configures the graph to use the provided working directory.
func WithWorkDir(workdir string) Option {
	return func(m *manager) error {
		m.workdir = workdir
		return nil
	}
}

// WithCurrentDirectory configures the working directory to be the current directory
func WithCurrentDirectory(m *manager) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	return WithWorkDir(wd)(m)
}

// WithContext configures the graph to use the provided Go build context.
func WithContext(ctx *gb.Context) Option {
	return func(m *manager) error {
		m.ctx = ctx
		return nil
	}
}

// WithFSNotify configures the graph to use fsnotify to implement its watcher and error channel.
func WithFSNotify(m *manager) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	m.watcher = watcher
	m.errCh = watcher.Errors
	m.eventCh = watcher.Events
	return nil
}

// WithFileFilter configures the graph implementation with the provided file filters.
func WithFileFilter(ff ...FileFilter) Option {
	return func(m *manager) error {
		m.fileFilters = append(m.fileFilters, ff...)
		return nil
	}
}

// WithPackageFilter configures the graph implementation with the provided package filters.
func WithPackageFilter(ff ...PackageFilter) Option {
	return func(m *manager) error {
		m.pkgFilters = append(m.pkgFilters, ff...)
		return nil
	}
}

// WithOutsideWorkDirFilter configures the graph with a package filter that omits files outside
// of the configured working directory.
func WithOutsideWorkDirFilter(m *manager) error {
	m.pkgFilters = append(m.pkgFilters, func(pkg *gb.Package) bool {
		return !strings.HasPrefix(pkg.Dir, m.workdir)
	})
	return nil
}
