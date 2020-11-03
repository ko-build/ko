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

package publish

import (
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/ko/pkg/build"
)

// Interface abstracts different methods for publishing images.
type Interface interface {
	// Publish uploads the given build.Result to a registry incorporating the
	// provided string into the image's repository name.  Returns the digest
	// of the published image.
	Publish(build.Result, string) (name.Reference, error)

	// Close exists for the tarball implementation so we can
	// do the whole thing in one write.
	Close() error

	// MultiPublish uploads the given build.Results to a registry
	// incorporating the associated importpath string into the image's
	// repository name. It returns a map of importpaths to digest
	// references of published images.
	//
	// MultiPublish can be implemented naively using NaiveMultiPublish,
	// which simply calls the implementation's Publish method for each item
	// in the map.
	MultiPublish(map[string]build.Result) (map[string]name.Reference, error)
}

// NaiveMultiPublish naively implements MultiPublish by calling the Interface
// implementation's Publish method for each item in the map, not taking
// advantage of any publisher-specific optimizations.
func NaiveMultiPublish(i Interface, m map[string]build.Result) (map[string]name.Reference, error) {
	// Fallback to naively publishing each image one at a time.
	out := map[string]name.Reference{}
	for k, v := range m {
		r, err := i.Publish(v, k)
		if err != nil {
			return nil, err
		}
		out[k] = r
	}
	return out, nil
}
