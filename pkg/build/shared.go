// Copyright 2018 Google LLC All Rights Reserved.
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

package build

import (
	"sync"
	"unicode/utf8"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

// Caching wraps a builder implementation in a layer that shares build results
// for the same inputs using a simple "future" implementation.  Cached results
// may be invalidated by calling Invalidate with the same input passed to Build.
type Caching struct {
	inner Interface

	m            sync.Mutex
	results      map[string]*future
	cm           sync.Mutex
	supportCache map[string]bool
}

// Caching implements Interface
var _ Interface = (*Caching)(nil)

// NewCaching wraps the provided build.Interface in an implementation that
// shares build results for a given path until the result has been invalidated.
func NewCaching(inner Interface) (*Caching, error) {
	return &Caching{
		inner:        inner,
		results:      make(map[string]*future),
		supportCache: make(map[string]bool),
	}, nil
}

// Build implements Interface
func (c *Caching) Build(ip string) (v1.Image, error) {
	f := func() *future {
		// Lock the map of futures.
		c.m.Lock()
		defer c.m.Unlock()

		// If a future for "ip" exists, then return it.
		f, ok := c.results[ip]
		if ok {
			return f
		}
		// Otherwise create and record a future for a Build of "ip".
		f = newFuture(func() (v1.Image, error) {
			return c.inner.Build(ip)
		})
		c.results[ip] = f
		return f
	}()

	return f.Get()
}

// SafeArg is used to check if the import path is valid.
// Copy of github.com/golang/go/blob/master/src/cmd/go/internal/load/pkg.go
func SafeArg(name string) bool {
	if name == "" {
		return false
	}
	c := name[0]
	return '0' <= c && c <= '9' || 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || c == '.' || c == '_' || c == '/' || c >= utf8.RuneSelf
}

// IsSupportedReference implements Interface
func (c *Caching) IsSupportedReference(ip string) bool {
	if !SafeArg(ip) {
		return false
	}

	c.cm.Lock()
	defer c.cm.Unlock()

	if supported, ok := c.supportCache[ip]; ok {
		return supported
	}

	c.supportCache[ip] = c.inner.IsSupportedReference(ip)

	return c.supportCache[ip]
}

// Invalidate removes an import path's cached results.
func (c *Caching) Invalidate(ip string) {
	c.m.Lock()
	defer c.m.Unlock()

	c.cm.Lock()
	defer c.cm.Unlock()

	delete(c.supportCache, ip)
	delete(c.results, ip)
}
