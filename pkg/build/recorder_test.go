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

package build

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type fake struct {
	isr func(string) error
	b   func(string) (Result, error)
}

var _ Interface = (*fake)(nil)

// QualifyImport implements Interface
func (r *fake) QualifyImport(ip string) (string, error) { return ip, nil }

// IsSupportedReference implements Interface
func (r *fake) IsSupportedReference(ip string) error { return r.isr(ip) }

// Build implements Interface
func (r *fake) Build(_ context.Context, ip string) (Result, error) { return r.b(ip) }

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
				isr: func(ip string) error {
					called = true
					if ip != test.input {
						t.Errorf("ISR = %v, wanted %v", ip, test.input)
					}
					return nil
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
				b: func(_ string) (Result, error) {
					return nil, nil
				},
			}
			rec := &Recorder{
				Builder: inner,
			}
			for _, in := range test.inputs {
				rec.Build(context.Background(), in)
			}
			if diff := cmp.Diff(test.inputs, rec.ImportPaths); diff != "" {
				t.Errorf("Build (-want, +got): %s", diff)
			}
		})
	}
}
