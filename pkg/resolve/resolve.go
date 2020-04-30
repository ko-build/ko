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
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/dprotaso/go-yit"
	"github.com/google/ko/pkg/build"
	"github.com/google/ko/pkg/commands/options"
	"github.com/google/ko/pkg/parameters"
	"github.com/google/ko/pkg/publish"
	"github.com/mattmoor/dep-notify/pkg/graph"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/labels"
)

// ImageReferences resolves supported references to images within the input yaml
// to published image digests.
//
// If a reference can be built and pushed, its yaml.Node will be mutated.
func ImageReferences(ctx context.Context, docs []*yaml.Node, strict bool, builder build.Interface, publisher publish.Interface) error {
	// First, walk the input objects and collect a list of supported references
	refs := make(map[string][]*yaml.Node)

	for _, doc := range docs {
		it := refsFromDoc(doc, strict)

		for node, ok := it(); ok; node, ok = it() {
			ref := strings.TrimSpace(node.Value)

			if builder.IsSupportedReference(ref) {
				refs[ref] = append(refs[ref], node)
			} else if strict {
				return fmt.Errorf("found strict reference but %s is not a valid import path", ref)
			}
		}
	}

	// Next, perform parallel builds for each of the supported references.
	var sm sync.Map
	var errg errgroup.Group
	for ref := range refs {
		ref := ref
		errg.Go(func() error {
			img, err := builder.Build(ctx, ref)
			if err != nil {
				return err
			}
			digest, err := publisher.Publish(img, ref)
			if err != nil {
				return err
			}
			sm.Store(ref, digest.String())
			return nil
		})
	}
	if err := errg.Wait(); err != nil {
		return err
	}

	// Walk the tags and update them with their digest.
	for ref, nodes := range refs {
		digest, ok := sm.Load(ref)

		if !ok {
			return fmt.Errorf("resolved reference to %q not found", ref)
		}

		for _, node := range nodes {
			node.Value = digest.(string)
		}
	}

	return nil
}

func refsFromDoc(doc *yaml.Node, strict bool) yit.Iterator {
	it := yit.FromNode(doc).
		RecurseNodes().
		Filter(yit.StringValue)

	if strict {
		return it.Filter(yit.WithPrefix(build.StrictScheme))
	}

	return it
}

// resolvedFuture represents a "future" for the bytes of a resolved file.
type resolvedFuture chan []byte

func ResolveFilesToWriter(
	ctx context.Context,
	builder *build.Caching,
	publisher publish.Interface,
	fo *parameters.FilenameParameters,
	so *parameters.SelectorParameters,
	sto *parameters.StrictParameters,
	out io.WriteCloser) error {
	defer out.Close()

	// By having this as a channel, we can hook this up to a filesystem
	// watcher and leave `fs` open to stream the names of yaml files
	// affected by code changes (including the modification of existing or
	// creation of new yaml files).
	fs := options.EnumerateFiles(fo)

	// This tracks filename -> []importpath
	var sm sync.Map

	var g graph.Interface
	var errCh chan error
	var err error
	if fo.Watch {
		// Start a dep-notify process that on notifications scans the
		// file-to-recorded-build map and for each affected file resends
		// the filename along the channel.
		g, errCh, err = graph.New(func(ss graph.StringSet) {
			sm.Range(func(k, v interface{}) bool {
				key := k.(string)
				value := v.([]string)

				for _, ip := range value {
					if ss.Has(ip) {
						// See the comment above about how "builder" works.
						builder.Invalidate(ip)
						fs <- key
					}
				}
				return true
			})
		})
		if err != nil {
			return fmt.Errorf("creating dep-notify graph: %v", err)
		}
		// Cleanup the fsnotify hooks when we're done.
		defer g.Shutdown()
	}

	// This tracks resolution errors and ensures we cancel other builds if an
	// individual build fails.
	errs, ctx := errgroup.WithContext(ctx)

	var futures []resolvedFuture
	for {
		// Each iteration, if there is anything in the list of futures,
		// listen to it in addition to the file enumerating channel.
		// A nil channel is never available to receive on, so if nothing
		// is available, this will result in us exclusively selecting
		// on the file enumerating channel.
		var bf resolvedFuture
		if len(futures) > 0 {
			bf = futures[0]
		} else if fs == nil {
			// There are no more files to enumerate and the futures
			// have been drained, so quit.
			break
		}

		select {
		case file, ok := <-fs:
			if !ok {
				// a nil channel is never available to receive on.
				// This allows us to drain the list of in-process
				// futures without this case of the select winning
				// each time.
				fs = nil
				break
			}

			// Make a new future to use to ship the bytes back and append
			// it to the list of futures (see comment below about ordering).
			ch := make(resolvedFuture)
			futures = append(futures, ch)

			// Kick off the resolution that will respond with its bytes on
			// the future.
			f := file // defensive copy
			errs.Go(func() error {
				defer close(ch)
				// Record the builds we do via this builder.
				recordingBuilder := &build.Recorder{
					Builder: builder,
				}
				b, err := ResolveFile(ctx, f, recordingBuilder, publisher, so, sto)
				if err != nil {
					// This error is sometimes expected during watch mode, so this
					// isn't fatal. Just print it and keep the watch open.
					err := fmt.Errorf("error processing import paths in %q: %v", f, err)
					if fo.Watch {
						log.Print(err)
						return nil
					}
					return err
				}
				// Associate with this file the collection of binary import paths.
				sm.Store(f, recordingBuilder.ImportPaths)
				ch <- b
				if fo.Watch {
					for _, ip := range recordingBuilder.ImportPaths {
						// Technically we never remove binary targets from the graph,
						// which will increase our graph's watch load, but the
						// notifications that they change will result in no affected
						// yamls, and no new builds or deploys.
						if err := g.Add(ip); err != nil {
							// If we're in watch mode, just fail.
							err := fmt.Errorf("adding importpath to dep graph: %v", err)
							errCh <- err
							return err
						}
					}
				}
				return nil
			})

		case b, ok := <-bf:
			// Once the head channel returns something, dequeue it.
			// We listen to the futures in order to be respectful of
			// the kubectl apply ordering, which matters!
			futures = futures[1:]
			if ok {
				// Write the next body and a trailing delimiter.
				// We write the delimeter LAST so that when streamed to
				// kubectl it knows that the resource is complete and may
				// be applied.
				out.Write(append(b, []byte("\n---\n")...))
			}

		case err := <-errCh:
			return fmt.Errorf("watching dependencies: %v", err)
		}
	}

	// Make sure we exit with an error.
	// See https://github.com/google/ko/issues/84
	return errs.Wait()
}

func ResolveFile(
	ctx context.Context,
	f string,
	builder build.Interface,
	pub publish.Interface,
	so *parameters.SelectorParameters,
	sto *parameters.StrictParameters) (b []byte, err error) {

	var selector labels.Selector
	if so.Selector != "" {
		var err error
		selector, err = labels.Parse(so.Selector)

		if err != nil {
			return nil, fmt.Errorf("unable to parse selector: %v", err)
		}
	}

	if f == "-" {
		b, err = ioutil.ReadAll(os.Stdin)
	} else {
		b, err = ioutil.ReadFile(f)
	}
	if err != nil {
		return nil, err
	}

	var docNodes []*yaml.Node

	// The loop is to support multi-document yaml files.
	// This is handled by using a yaml.Decoder and reading objects until io.EOF, see:
	// https://godoc.org/gopkg.in/yaml.v3#Decoder.Decode
	decoder := yaml.NewDecoder(bytes.NewBuffer(b))
	for {
		var doc yaml.Node
		if err := decoder.Decode(&doc); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if selector != nil {
			if match, err := MatchesSelector(&doc, selector); err != nil {
				return nil, fmt.Errorf("error evaluating selector: %v", err)
			} else if !match {
				continue
			}
		}

		docNodes = append(docNodes, &doc)

	}

	if err := ImageReferences(ctx, docNodes, sto.Strict, builder, pub); err != nil {
		return nil, fmt.Errorf("error resolving image references: %v", err)
	}

	buf := &bytes.Buffer{}
	e := yaml.NewEncoder(buf)
	e.SetIndent(2)

	for _, doc := range docNodes {
		err := e.Encode(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to encode output: %v", err)
		}
	}
	e.Close()

	return buf.Bytes(), nil
}
