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
	"fmt"
	gb "go/build"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// New creates a new Interface for building up dependency graphs.
// It starts in the provided working directory, and will call the provided
// Observer for any changes.
//
// The returned graph is empty, but new targets may be added via the returned
// Interface.  New also returns any immediate errors, and a channel through which
// errors watching for changes in the dependencies will be returned until the
// graph is Shutdown.
//   // Create our empty graph
//   g, errCh, err := New(...)
//   if err != nil { ... }
//   // Cleanup when we're done.  This closes errCh.
//   defer g.Shutdown()
//   // Start tracking this target.
//   err := g.Add("github.com/mattmoor/warm-image/cmd/controller")
//   if err != nil { ... }
//   select {
//     case err := <- errCh:
//       // Handle errors that occur while watching the above target.
//     case <-stopCh:
//       // When some stop signal happens, we're done.
//   }
func New(obs Observer) (Interface, chan error, error) {
	return NewWithOptions(obs, DefaultOptions...)
}

var DefaultOptions = []Option{WithCurrentDirectory, WithContext(&gb.Default), WithFSNotify,
	WithFileFilter(OmitTest, OmitNonGo), WithPackageFilter(OmitVendor), WithOutsideWorkDirFilter}

func NewWithOptions(obs Observer, opts ...Option) (Interface, chan error, error) {
	m := &manager{
		packages: make(map[string]*node),
	}

	for _, opt := range opts {
		if err := opt(m); err != nil {
			if m.watcher != nil {
				m.watcher.Close()
			}
			return nil, nil, err
		}
	}

	// Start listening for events via the filesystem watcher.
	go func() {
		for {
			event, ok := <-m.eventCh
			if !ok {
				// When the channel has been closed, the watcher is shutting down
				// and we should return to cleanup the go routine.
				return
			}

			// Apply our file filters to improve the signal-to-noise.
			skip := false
			for _, f := range m.fileFilters {
				if f(event.Name) {
					skip = true
				}
			}
			if skip {
				continue
			}

			// Determine what package contains this file
			// and signal the change.  Call our Observer
			// on affected targets when we're done.
			if n := m.enclosingPackage(event.Name); n != nil {
				m.onChange(n, func(n *node) {
					obs(m.affectedTargets(n))
				})
			}
		}
	}()

	return m, m.errCh, nil
}

// notification is a callback that internal consumers of manager may use to get
// a crack at a node after it has been updated.
type notification func(*node)

// empty implements notification and does nothing.
func empty(n *node) {}

type watcher interface {
	Add(string) error
	Remove(string) error
	Close() error
}

type manager struct {
	m sync.Mutex

	ctx *gb.Context

	pkgFilters  []PackageFilter
	fileFilters []FileFilter

	packages map[string]*node
	watcher  watcher

	// The working directory relative to which import paths are evaluated.
	workdir string

	errCh   chan error
	eventCh chan fsnotify.Event
}

// manager implements Interface
var _ Interface = (*manager)(nil)

// manager implements fmt.Stringer
var _ fmt.Stringer = (*manager)(nil)

// Add implements Interface
func (m *manager) Add(importpath string) error {
	m.m.Lock()
	defer m.m.Unlock()

	_, _, err := m.add(importpath, nil)
	return err
}

// add adds the provided importpath (if it doesn't exist) and optionally
// adds the dependent node (if provided) as a dependent of the target.
// This returns the node for the provided importpath, whether the dependency
// structure has changed, and any errors that may have occurred adding the node.
func (m *manager) add(importpath string, dependent *node) (*node, bool, error) {
	// INVARIANT m.m must be held to call this.
	if pkg, ok := m.packages[importpath]; ok {
		return pkg, pkg.addDependent(dependent), nil
	}

	// New nodes always start as a simple shell, then we set up the
	// fsnotify and immediate simulate a change to prompt the package
	// to load its data.  A good analogy would be how the "diff" in
	// a code review for new files looks like everything being added;
	// so too does this first simulated change pull in the rest of the
	// dependency graph.
	newNode := &node{
		name: importpath,
	}
	m.packages[importpath] = newNode
	newNode.addDependent(dependent)

	// Load the package once to determine it's filesystem location,
	// and set up a watch on that location.
	pkg, err := m.ctx.Import(importpath, m.workdir, gb.ImportComment)
	if err != nil {
		newNode.removeDependent(dependent)
		delete(m.packages, importpath)
		return nil, false, err
	}
	if err := m.watcher.Add(pkg.Dir); err != nil {
		newNode.removeDependent(dependent)
		delete(m.packages, importpath)
		return nil, false, err
	}
	newNode.dir = pkg.Dir

	// This is done via go routine so that it can take over the lock.
	go m.onChange(newNode, empty)

	return newNode, true, nil
}

// affectedTargets returns the set of targets that would be affected by a
// change to the target represented by the given node.  This set is comprised
// of the transitive dependents of the node, including itself.
func (m *manager) affectedTargets(n *node) StringSet {
	m.m.Lock()
	defer m.m.Unlock()

	return n.transitiveDependents()
}

// enclosingPackage returns the node for the package covering the
// watched path.
func (m *manager) enclosingPackage(path string) *node {
	m.m.Lock()
	defer m.m.Unlock()

	dir := filepath.Dir(path)
	for _, v := range m.packages {
		if strings.HasSuffix(dir, v.dir) {
			return v
		}
	}
	return nil
}

// onChange updates the graph based on the current state of the package
// represented by the given node.  Once the graph has been updated, the
// notification function is called on the node.
func (m *manager) onChange(changed *node, not notification) {
	m.m.Lock()
	defer m.m.Unlock()

	// Load the package information and update dependencies.
	pkg, err := m.ctx.Import(changed.name, m.workdir, gb.ImportComment)
	if err != nil {
		m.errCh <- err
		return
	}

	// haveDepsChanged := false
	seen := make(StringSet)
	for _, ip := range pkg.Imports {
		if ip == "C" {
			// skip cgo
			continue
		}
		subpkg, err := m.ctx.Import(ip, m.workdir, gb.ImportComment)
		if err != nil {
			m.errCh <- err
			return
		}

		skip := false
		for _, f := range m.pkgFilters {
			if f(subpkg) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		n, chg, err := m.add(subpkg.ImportPath, changed)
		if err != nil {
			m.errCh <- err
			return
		} else if chg {
			// haveDepsChanged = true
		}
		if changed.addDependency(n) {
			// haveDepsChanged = true
		}
		seen.Add(subpkg.ImportPath)
	}

	// Remove dependencies that we no longer have.
	removed := changed.removeUnseenDependencies(seen)
	if len(removed) > 0 {
		// haveDepsChanged = true
	}
	for _, dependency := range removed {
		d := dependency
		go m.maybeGC(d)
	}

	// log.Printf("Processing %s, have deps changed: %v", changed.name, haveDepsChanged)
	// Done via go routine so that we can be passed a callback that
	// takes the lock on manager.
	go not(changed)
}

func (m *manager) maybeGC(n *node) {
	m.m.Lock()
	defer m.m.Unlock()

	if len(n.dependents) > 0 {
		// It has dependents, so it should not be removed.
		return
	}

	// If it has zero dependents, then remove it from the packages map.
	delete(m.packages, n.name)

	// Lookup the package information, so that we know what directory to stop watching.
	subpkg, err := m.ctx.Import(n.name, m.workdir, gb.ImportComment)
	if err != nil {
		m.errCh <- err
		return
	}

	// Remove the watch on the package's directory
	if err := m.watcher.Remove(subpkg.Dir); err != nil {
		m.errCh <- err
		return
	}

	for _, dependency := range n.dependencies {
		dependency.removeDependent(n)

		d := dependency
		go m.maybeGC(d)
	}
}

// Shutdown implements Interface.
func (m *manager) Shutdown() error {
	return m.watcher.Close()
}

// String implements fmt.Stringer
func (m *manager) String() string {
	m.m.Lock()
	defer m.m.Unlock()

	// WTB Topo sort.
	order := []string{}
	for k := range m.packages {
		order = append(order, k)
	}
	sort.Strings(order)

	parts := []string{}
	for _, key := range order {
		parts = append(parts, m.packages[key].String())
	}
	return strings.Join(parts, "\n")
}
