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
	"testing"

	v1 "github.com/google/go-containerregistry/pkg/v1"

	"github.com/google/go-cmp/cmp"
)

type fake struct {
	isr func(string) bool
	b   func(string) (v1.Image, error)
}

var _ Interface = (*fake)(nil)

// IsSupportedReference implements Interface
func (r *fake) IsSupportedReference(ip string) bool {
	return r.isr(ip)
}

// Build implements Interface
func (r *fake) Build(ip string) (v1.Image, error) {
	return r.b(ip)
}

func TestISRPassThrough(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{{
		name: "empty string",
	}, {
		name:  "non-empty string",
		input: "asdf asdf asdf",
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			called := false
			inner := &fake{
				isr: func(ip string) bool {
					called = true
					if ip != test.input {
						t.Errorf("ISR = %v, wanted %v", ip, test.input)
					}
					return true
				},
			}
			rec := &Recorder{
				Builder: inner,
			}
			rec.IsSupportedReference(test.input)
			if !called {
				t.Error("IsSupportedReference wasn't called, wanted called")
			}
		})
	}
}

func TestBuildRecording(t *testing.T) {
	tests := []struct {
		name   string
		inputs []string
	}{{
		name: "no calls",
	}, {
		name: "one call",
		inputs: []string{
			"github.com/foo/bar",
		},
	}, {
		name: "two calls",
		inputs: []string{
			"github.com/foo/bar",
			"github.com/foo/baz",
		},
	}, {
		name: "duplicates",
		inputs: []string{
			"github.com/foo/bar",
			"github.com/foo/baz",
			"github.com/foo/bar",
			"github.com/foo/baz",
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			inner := &fake{
				b: func(ip string) (v1.Image, error) {
					return nil, nil
				},
			}
			rec := &Recorder{
				Builder: inner,
			}
			for _, in := range test.inputs {
				rec.Build(in)
			}
			if diff := cmp.Diff(test.inputs, rec.ImportPaths); diff != "" {
				t.Errorf("Build (-want, +got): %s", diff)
			}
		})
	}
}
