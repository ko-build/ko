//
// Copyright 2022 The Sigstore Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package webhook

import (
	"context"
)

type cacheKey struct{}

// CacheResult wraps PolicyResult and errors that are suitable for caching
// purposes. By doing this we can make choices that control things like, should
// errors be cached, and if so, for how long that's independent of the
// successful validations.
type CacheResult struct {
	PolicyResult *PolicyResult
	Errors       []error
}

// FromContext extracts a cache from the provided context. If one has not been
// set, return the NoCache to fulfill the interface but it provides no caching.
func FromContext(ctx context.Context) ResultCache {
	x, ok := ctx.Value(cacheKey{}).(ResultCache)
	if ok {
		return x
	}
	return &NoCache{}
}

func ToContext(ctx context.Context, cache ResultCache) context.Context {
	return context.WithValue(ctx, cacheKey{}, cache)
}

type ResultCache interface {
	// Set caches a PolicyResult for a given CIP evaluated for a given image at
	// a particular point in time. image, uid & resourceVersion will give a
	// unique point in time, so we can make sure we're not caching things that
	// are out of date.
	Set(ctx context.Context, image, name, uid, resourceVersion string, cacheResult *CacheResult)

	// Get returns a cached result for a given image or nil if there are none.
	Get(ctx context.Context, image, uid, resourceVersion string) *CacheResult
}
