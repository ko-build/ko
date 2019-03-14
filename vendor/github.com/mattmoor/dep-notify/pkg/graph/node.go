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
)

// node holds a single node in our dependency graph and its immediately adjacent neighbors.
type node struct {
	// The canonical import path for this node.
	name string

	// The directory from which we pull this import path.
	dir string

	// The dependency structure
	dependencies []*node
	dependents   []*node
}

// node implements fmt.Stringer
var _ fmt.Stringer = (*node)(nil)

// String implements fmt.Stringer
func (n *node) String() string {
	return fmt.Sprintf(`---
name: %s
depdnt: %v
depdncy: %v`, n.name, names(n.dependents), names(n.dependencies))
}

// addDependent adds the given dependent to the list of dependents for
// the given dependency.  Dependent may be nil.
// This returns whether the node's neighborhood changes.
// The manager's lock must be held before calling this.
func (dependency *node) addDependent(dependent *node) bool {
	if dependent == nil {
		return false
	}

	for _, depdnt := range dependency.dependents {
		if depdnt == dependent {
			// Already a dependent
			return false
		}
	}

	// log.Printf("Adding %s <- %s", dependent.name, dependency.name)
	dependency.dependents = append(dependency.dependents, dependent)
	return true
}

// addDependency adds the given dependency to the list of dependencies for
// the given dependent.  Neither parameter may be nil.
// This returns whether the node's neighborhood changes.
// The manager's lock must be held before calling this.
func (dependent *node) addDependency(dependency *node) bool {
	for _, depdcy := range dependent.dependencies {
		if depdcy == dependency {
			// Already a dependency
			return false
		}
	}

	// log.Printf("Adding %s -> %s", dependent.name, dependency.name)
	dependent.dependencies = append(dependent.dependencies, dependency)
	return true
}

// removeDependent removes the given dependent to the list of dependents for
// the given dependency.  Dependent may be nil.
// This returns whether the node's neighborhood changes.
// The manager's lock must be held before calling this.
func (dependency *node) removeDependent(dependent *node) bool {
	for i, depdnt := range dependency.dependents {
		if depdnt == dependent {
			dependency.dependents = append(
				dependency.dependents[:i], dependency.dependents[i+1:]...)
			return true
		}
	}

	return false
}

// removeDependency removes the given dependency to the list of dependencies for
// the given dependent.  Neither parameter may be nil.
// This returns whether the node's neighborhood changes.
// The manager's lock must be held before calling this.
func (dependent *node) removeDependency(dependency *node) bool {
	for i, depdcy := range dependent.dependencies {
		if depdcy == dependency {
			dependent.dependencies = append(
				dependent.dependencies[:i], dependent.dependencies[i+1:]...)
			return true
		}
	}
	return false
}

// removeUnseenDependencies removes dependencies that aren't in the list of seen
// names.  Returns the list of nodes that correspond to the named dependencies removed.
func (dependent *node) removeUnseenDependencies(seen StringSet) []*node {
	keep := make([]*node, 0, len(dependent.dependencies))
	toss := make([]*node, 0, len(dependent.dependencies))
	for _, dep := range dependent.dependencies {
		if seen.Has(dep.name) {
			keep = append(keep, dep)
		} else {
			toss = append(toss, dep)
			dependent.removeDependency(dep)
			dep.removeDependent(dependent)
		}
	}
	dependent.dependencies = keep
	return toss
}

// transitiveDependents collects the set of node names transitively reachable
// by following the "dependents" edges.
func (n *node) transitiveDependents() StringSet {
	return n.transitiveFoo(func(n *node) []*node {
		return n.dependents
	})
}

// transitiveDependencies collects the set of node names transitively reachable
// by following the "dependencies" edges.
func (n *node) transitiveDependencies() StringSet {
	return n.transitiveFoo(func(n *node) []*node {
		return n.dependencies
	})
}

// neighbors defines a function for returning the neighbors of a
// particular node.
type neighbors func(*node) []*node

// transitiveDependents collects the set of node names transitively reachable
// by following the edges determines by the "neighbors" function.
func (n *node) transitiveFoo(nbr neighbors) StringSet {
	// We will use "set" as the visited group for our DFS.
	set := make(StringSet)
	queue := []*node{n}

	for len(queue) != 0 {
		// Pop the top element off of the queue.
		top := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		// Check/Mark visited
		if set.Has(top.name) {
			continue
		}
		set.Add(top.name)

		// Append this node's dependents to our search.
		queue = append(queue, nbr(top)...)
	}

	return set
}

// names returns the deduplicated and sorted names of the provided nodes.
func names(ns []*node) []string {
	ss := make(StringSet, len(ns))
	for _, n := range ns {
		ss.Add(n.name)
	}
	return ss.InOrder()
}
