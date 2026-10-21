// Copyright 2018 ko Build Authors All Rights Reserved.
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

package publish

import (
	"context"
	"errors"
	"sync"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/ko/pkg/build"
)

// caching wraps a publisher implementation in a layer that shares publish results
// for the same inputs using a simple "future" implementation.
type caching struct {
	inner Interface

	m       sync.Mutex
	results map[string]*entry
}

// entry holds the last image published and the result of publishing it for a
// particular reference.
type entry struct {
	br build.Result
	f  *future
}

// caching implements Interface
var _ Interface = (*caching)(nil)

// NewCaching wraps the provided publish.Interface in an implementation that
// shares publish results for a given path until the passed image object changes.
func NewCaching(inner Interface) (Interface, error) {
	return &caching{
		inner:   inner,
		results: make(map[string]*entry),
	}, nil
}

// Publish implements Interface
func (c *caching) Publish(ctx context.Context, br build.Result, ref string) (name.Reference, error) {
	for {
		f, owner := func() (*future, bool) {
			// Lock the map of futures.
			c.m.Lock()
			defer c.m.Unlock()

			// If a future for "ref" exists, then return it.
			ent, ok := c.results[ref]
			if ok {
				// If the image matches, then return the same future.
				if ent.br == br {
					return ent.f, false
				}
			}
			// Otherwise create and record a future for publishing "br" to "ref".
			f := newFuture(func() (name.Reference, error) {
				return c.inner.Publish(ctx, br, ref)
			})
			c.results[ref] = &entry{br: br, f: f}
			return f, true
		}()

		res, err := f.Get()
		// The future runs with the context of the caller that created it. If
		// that context was cancelled but ours wasn't, forget the result and
		// publish again with our context instead of returning its error.
		if !owner && isContextError(err) && ctx.Err() == nil {
			c.m.Lock()
			if ent, ok := c.results[ref]; ok && ent.f == f {
				delete(c.results, ref)
			}
			c.m.Unlock()
			continue
		}
		return res, err
	}
}

func isContextError(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

func (c *caching) Close() error {
	return c.inner.Close()
}
