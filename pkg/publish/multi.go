// Copyright 2020 Google LLC All Rights Reserved.
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
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

// MultiPublisher creates a publisher that publishes to all
// the provided publishers, similar to the Unix tee(1) command.
//
// When calling Publish, the name.Reference returned will be the return value
// of the last publisher passed to MultiPublisher (last one wins).
func MultiPublisher(publishers ...Interface) Interface {
	return &multiPublisher{publishers}
}

type multiPublisher struct {
	publishers []Interface
}

// Publish implements publish.Interface.
func (p *multiPublisher) Publish(img v1.Image, s string) (ref name.Reference, err error) {
	for _, pub := range p.publishers {
		ref, err = pub.Publish(img, s)
		if err != nil {
			return
		}
	}

	return
}
