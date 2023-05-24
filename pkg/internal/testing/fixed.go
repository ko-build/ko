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

package testing

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/ko/pkg/build"
	"github.com/google/ko/pkg/publish"
)

type fixedBuild struct {
	entries map[string]build.Result
}

// NewFixedBuild returns a build.Interface implementation that simply resolves
// particular references to fixed v1.Image objects
func NewFixedBuild(entries map[string]build.Result) build.Interface {
	return &fixedBuild{entries}
}

// QualifyImport implements build.Interface
func (f *fixedBuild) QualifyImport(ip string) (string, error) {
	return ip, nil
}

// IsSupportedReference implements build.Interface
func (f *fixedBuild) IsSupportedReference(s string) error {
	s = strings.TrimPrefix(s, build.StrictScheme)
	if _, ok := f.entries[s]; !ok {
		return errors.New("importpath is not supported")
	}
	return nil
}

// Build implements build.Interface
func (f *fixedBuild) Build(_ context.Context, s string) (build.Result, error) {
	s = strings.TrimPrefix(s, build.StrictScheme)
	if img, ok := f.entries[s]; ok {
		return img, nil
	}
	return nil, fmt.Errorf("unsupported reference: %q", s)
}

type fixedPublish struct {
	base    name.Repository
	entries map[string]v1.Hash
}

// NewFixedPublish returns a publish.Interface implementation that simply
// resolves particular references to fixed name.Digest references.
func NewFixedPublish(base name.Repository, entries map[string]v1.Hash) publish.Interface {
	return &fixedPublish{base, entries}
}

// Publish implements publish.Interface
func (f *fixedPublish) Publish(_ context.Context, _ build.Result, s string) (name.Reference, error) {
	s = strings.TrimPrefix(s, build.StrictScheme)
	h, ok := f.entries[s]
	if !ok {
		return nil, fmt.Errorf("unsupported importpath: %q", s)
	}
	d, err := name.NewDigest(fmt.Sprintf("%s/%s@%s", f.base, s, h))
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (f *fixedPublish) Close() error {
	return nil
}

func ComputeDigest(base name.Repository, ref string, h v1.Hash) string {
	d, err := name.NewDigest(fmt.Sprintf("%s/%s@%s", base, ref, h))
	if err != nil {
		panic(err)
	}
	return d.String()
}
