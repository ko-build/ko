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

import "context"

// NoCache is pretty much what it says, it caches nothing. Just meant to
// implement the interface that we can test with as well as if there is no
// caching wanted, we can do that by injecting this.
type NoCache struct {
}

func (nc *NoCache) Get(ctx context.Context, image, uid, resourceVersion string) *CacheResult {
	return nil
}

func (nc *NoCache) Set(ctx context.Context, image, name, uid, resourceVersion string, cacheResult *CacheResult) {
}
