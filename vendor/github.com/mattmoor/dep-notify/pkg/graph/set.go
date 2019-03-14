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
	"sort"
)

// StringSet is a simple abstraction for holding a collection of deduplicated strings.
type StringSet map[string]struct{}

// Add inserts a provided key into our set.
func (ss *StringSet) Add(key string) {
	(*ss)[key] = struct{}{}
}

// Remove deletes a provided key from our set.
func (ss *StringSet) Remove(key string) {
	delete(*ss, key)
}

// Has returns whether the set contains the provided key.
func (ss *StringSet) Has(key string) bool {
	_, ok := (*ss)[key]
	return ok
}

// InOrder returns the keys of the set in the ordering determined by sort.Strings.
func (ss StringSet) InOrder() []string {
	keys := make([]string, 0, len(ss))
	for k := range ss {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}
