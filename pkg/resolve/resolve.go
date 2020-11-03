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

package resolve

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/dprotaso/go-yit"
	"github.com/google/ko/pkg/build"
	"github.com/google/ko/pkg/publish"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

// ImageReferences resolves supported references to images within the input yaml
// to published image digests.
//
// If a reference can be built and pushed, its yaml.Node will be mutated.
func ImageReferences(ctx context.Context, docs []*yaml.Node, strict bool, builder build.Interface, publisher publish.Interface) error {
	// First, walk the input objects and collect a list of supported references
	importpaths := make(map[string][]*yaml.Node)

	for _, doc := range docs {
		it := importPathsFromDoc(doc, strict)

		for node, ok := it(); ok; node, ok = it() {
			ip := strings.TrimSpace(node.Value)

			if err := builder.IsSupportedReference(ip); err == nil {
				importpaths[ip] = append(importpaths[ip], node)
			} else if strict {
				return fmt.Errorf("found strict reference but %s is not a valid import path", ip)
			}
		}
	}

	// Next, perform parallel builds for each of the supported references.
	var sm sync.Map // importpath string -> build.Result
	var errg errgroup.Group
	for ip := range importpaths {
		ip := ip
		errg.Go(func() error {
			img, err := builder.Build(ctx, ip)
			if err != nil {
				return err
			}
			sm.Store(ip, img)
			return nil
		})
	}
	if err := errg.Wait(); err != nil {
		return err
	}

	// Publish all images.
	m := map[string]build.Result{}
	sm.Range(func(k, v interface{}) bool {
		ip, br := k.(string), v.(build.Result)
		m[ip] = br
		return true
	})
	published, err := publisher.MultiPublish(m)
	if err != nil {
		return err
	}

	// Walk the tags and update them with their digest.
	for ip, nodes := range importpaths {
		ref, found := published[ip]
		if !found {
			return fmt.Errorf("resolved reference to %q not found", ip)
		}

		for _, node := range nodes {
			node.Value = ref.String()
		}
	}

	return nil
}

func importPathsFromDoc(doc *yaml.Node, strict bool) yit.Iterator {
	it := yit.FromNode(doc).
		RecurseNodes().
		Filter(yit.StringValue)

	if strict {
		return it.Filter(yit.WithPrefix(build.StrictScheme))
	}

	return it
}
