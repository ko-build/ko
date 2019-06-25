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
	"testing"
)

func TestSelector(t *testing.T) {
	tests := []struct {
		desc     string
		input    string
		selector string
		expected string
	}{{
		desc: "single object with matching selector",
		input: `
apiVersion: v1
kind: Pod
metadata:
  labels:
    app: web
  name: rss-site
`,
		selector: `app=web`,
		expected: `apiVersion: v1
kind: Pod
metadata:
  labels:
    app: web
  name: rss-site
`,
	},
	{
		desc: "single object with non-matching selector",
		input: `
apiVersion: v1
kind: Pod
metadata:
  labels:
    app: web
  name: rss-site

`,
		selector: `app!=web`,
		expected: ``,
	},
	{
		desc: "selector matching 1 of two objects",
		input: `
apiVersion: v1
kind: Pod
metadata:
  labels:
    app: web
  name: rss-site
---
apiVersion: v1
kind: Pod
metadata:
  labels:
    app: db
  name: rss-db

`,
		selector: `app=web`,
		expected: `apiVersion: v1
kind: Pod
metadata:
  labels:
    app: web
  name: rss-site
`,
	},
	{
		desc: "selector matching 1 of two objects",
		input: `
apiVersion: v1
kind: Pod
metadata:
  labels:
    app: web
  name: rss-site
---
apiVersion: v1
kind: Pod
metadata:
  labels:
    app: db
  name: rss-db

`,
		selector: `app!=web`,
		expected: `apiVersion: v1
kind: Pod
metadata:
  labels:
    app: db
  name: rss-db
`,
	},
	{
		desc: "selector matching elements of list object",
		input: `
apiVersion: v1
kind: List
metadata:
  resourceVersion: ""
  selfLink: ""
items:
- apiVersion: v1
  kind: Pod
  metadata:
    labels:
      app: web
    name: rss-site
- apiVersion: v1
  kind: Pod
  metadata:
    labels:
      app: db
    name: rss-db

`,
		selector: `app=web`,
		expected: `apiVersion: v1
items:
- apiVersion: v1
  kind: Pod
  metadata:
    labels:
      app: web
    name: rss-site
kind: List
metadata:
  resourceVersion: ""
  selfLink: ""
`,
	},
	{
		desc: "selector matching elements of list object",
		input: `
apiVersion: v1
kind: List
metadata:
  resourceVersion: ""
  selfLink: ""
items:
- apiVersion: v1
  kind: Pod
  metadata:
    labels:
      app: web
    name: rss-site
- apiVersion: v1
  kind: Pod
  metadata:
    labels:
      app: db
    name: rss-db

`,
		selector: `app!=web`,
		expected: `apiVersion: v1
items:
- apiVersion: v1
  kind: Pod
  metadata:
    labels:
      app: db
    name: rss-db
kind: List
metadata:
  resourceVersion: ""
  selfLink: ""
`,
	}}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			filtered, err := FilterBySelector([]byte(test.input), test.selector)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if string(filtered) != test.expected {
				t.Errorf("expected \n%v\n to equal \n%v\n ", string(filtered), test.expected)
			}
		})
	}
}
