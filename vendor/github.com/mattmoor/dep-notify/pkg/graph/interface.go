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

// Interface for manipulating the dependency graph.
type Interface interface {
	// This adds a given importpath to the collection of roots that we are tracking.
	Add(importpath string) error

	// Shutdown stops tracking all Add'ed import paths for changes.
	Shutdown() error
}

// Observer is the type for the callback that will happen on changes.
// The callback is supplied with the transitive dependents (aka "affected
// targets") of the file that has changed.
type Observer func(affected StringSet)
