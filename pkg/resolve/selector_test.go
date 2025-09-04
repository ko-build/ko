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

package resolve

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/labels"
)

var (
	webSelector    = selector(`app=web`)
	notWebSelector = selector(`app!=web`)
	nopSelector    = selector(`foo!=bark`)

	hasSelector    = selector(`app`)
	notHasSelector = selector(`!app`)
)

const (
	webPod = `apiVersion: v1
kind: Pod
metadata:
  labels:
    # comments should be preserved
    app: web
  name: rss-site
`
	dbPod = `apiVersion: v1
kind: Pod
metadata:
  labels:
    # comments should be preserved
    app: db
  name: rss-db
`
	podNoLabel = `apiVersion: v1
kind: Pod
metadata:
  name: rss-site
`
	podList = `apiVersion: v1
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
`
	webPodList = `apiVersion: v1
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
`
	dbPodList = `apiVersion: v1
kind: List
metadata:
  resourceVersion: ""
  selfLink: ""
items:
- apiVersion: v1
  kind: Pod
  metadata:
    labels:
      app: db
    name: rss-db
`
	podListNoLabel = `apiVersion: v1
kind: List
metadata:
  resourceVersion: ""
  selfLink: ""
items:
- apiVersion: v1
  kind: Pod
  metadata:
    name: rss-site
- apiVersion: v1
  kind: Pod
  metadata:
    name: rss-db
`
)

func TestMatchesSelector(t *testing.T) {
	tests := []struct {
		desc     string
		input    string
		selector labels.Selector
		output   string
		matches  bool
	}{{
		desc:     "single object with matching selector",
		input:    webPod,
		selector: webSelector,
		output:   webPod,
		matches:  true,
	}, {
		desc:     "single object with non-matching selector",
		input:    webPod,
		selector: notWebSelector,
		matches:  false,
	}, {
		desc:     "single object with noop selector",
		input:    dbPod,
		selector: nopSelector,
		output:   dbPod,
		matches:  true,
	}, {
		desc:     "single object with has selector",
		input:    webPod,
		selector: hasSelector,
		output:   webPod,
		matches:  true,
	}, {
		desc:     "single object with not-has selector",
		input:    webPod,
		selector: notHasSelector,
		matches:  false,
	}, {
		desc:     "single non-labeled object with has selector",
		input:    podNoLabel,
		selector: hasSelector,
		matches:  false,
	}, {
		desc:     "single non-labeled object with not-has selector",
		input:    podNoLabel,
		selector: notHasSelector,
		output:   podNoLabel,
		matches:  true,
	}, {
		desc:     "selector matching elements of list object",
		input:    podList,
		selector: webSelector,
		output:   webPodList,
		matches:  true,
	}, {
		desc:     "selector matching other elements of list object",
		input:    podList,
		selector: notWebSelector,
		output:   dbPodList,
		matches:  true,
	}, {
		desc:     "has selector matching no non-labeled element of list object",
		input:    podListNoLabel,
		selector: hasSelector,
		matches:  false,
	}, {
		desc:     "not-has selector matching all non-labeled elements of list object",
		input:    podListNoLabel,
		selector: notHasSelector,
		output:   podListNoLabel,
		matches:  true,
	}, {
		desc:     "selector matching all elements of list object",
		input:    podList,
		selector: labels.Everything(),
		output:   podList,
		matches:  true,
	}, {
		desc:     "selector matching no element of list object",
		input:    podList,
		selector: labels.Nothing(),
		matches:  false,
	}}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			doc := strToYAML(t, test.input)
			matches, err := MatchesSelector(doc, test.selector)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if matches != test.matches {
				t.Errorf("unexpected result: got %v - want %v", matches, test.matches)
			}

			// assert doc is mutated correctly
			if test.output != "" {
				// Normalize whitespace formatting
				output := normalizeYAML(t, test.output)

				if diff := cmp.Diff(output, yamlToStr(t, doc)); diff != "" {
					t.Errorf("unexpected diff (-want, +got) %v", diff)
				}
			}
		})
	}
}

func TestSelectorFailure(t *testing.T) {
	tests := []struct {
		desc  string
		input string
	}{
		{
			desc:  "not an object",
			input: "image: some.go/package",
		},
		{
			desc: "not an object in a list",
			input: `apiVersion: v1
kind: List
metadata:
  resourceVersion: ""
  selfLink: ""
items:
- blah: ha
`,
		},
		{
			desc: "not a valid list",
			input: `apiVersion: v1
kind: List
`,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			_, err := MatchesSelector(strToYAML(t, test.input), labels.Everything())

			if err == nil {
				t.Error("expected error")
			}
		})
	}
}

func selector(s string) labels.Selector {
	selector, err := labels.Parse(s)
	if err != nil {
		panic("unable to parse selector " + s)
	}
	return selector
}

func normalizeYAML(t *testing.T, yuml string) string {
	t.Helper()
	return yamlToStr(t, strToYAML(t, yuml))
}

func yamlToStr(t *testing.T, node *yaml.Node) string {
	result, err := yaml.Marshal(node)
	if err != nil {
		t.Fatalf("error marshalling yaml: %v", err)
	}

	return string(result)
}

func strToYAML(t *testing.T, yuml string) *yaml.Node {
	t.Helper()
	var node yaml.Node

	if err := yaml.Unmarshal([]byte(yuml), &node); err != nil {
		t.Fatalf("error unmarshalling yaml: %v\n%v", err, yuml)
	}

	return &node
}

func TestLabelsNode(t *testing.T) {
	tests := []struct {
		desc   string
		labels map[string]string
		label  string
		value  string
		exists bool
	}{{
		desc:   "label exists",
		labels: map[string]string{"app": "web", "version": "v1"},
		label:  "app",
		value:  "web",
		exists: true,
	}, {
		desc:   "label does not exist",
		labels: map[string]string{"app": "web", "version": "v1"},
		label:  "tier",
		value:  "",
		exists: false,
	}, {
		desc:   "empty labels",
		labels: map[string]string{},
		label:  "app",
		value:  "",
		exists: false,
	}, {
		desc:   "label with empty value",
		labels: map[string]string{"app": "", "version": "v1"},
		label:  "app",
		value:  "",
		exists: true,
	}, {
		desc:   "multiple labels with special characters",
		labels: map[string]string{"app.io/name": "web", "version": "v1", "owner": "team-a"},
		label:  "app.io/name",
		value:  "web",
		exists: true,
	}, {
		desc:   "label with numeric value",
		labels: map[string]string{"replicas": "3", "port": "8080"},
		label:  "replicas",
		value:  "3",
		exists: true,
	}, {
		desc:   "case sensitive label lookup",
		labels: map[string]string{"App": "web", "app": "database"},
		label:  "app",
		value:  "database",
		exists: true,
	}}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			// Create a mapping node directly
			node := &yaml.Node{
				Kind: yaml.MappingNode,
			}

			// Add key-value pairs to the node
			for key, value := range test.labels {
				keyNode := &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: key,
				}
				valueNode := &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: value,
				}
				node.Content = append(node.Content, keyNode, valueNode)
			}

			ln := labelsNode{node}

			// Test Get method
			gotValue := ln.Get(test.label)
			if gotValue != test.value {
				t.Errorf("Get(%q) = %q, want %q", test.label, gotValue, test.value)
			}

			// Test Has method
			gotExists := ln.Has(test.label)
			if gotExists != test.exists {
				t.Errorf("Has(%q) = %v, want %v", test.label, gotExists, test.exists)
			}

			// Test Lookup method
			gotLookupValue, gotLookupExists := ln.Lookup(test.label)
			if gotLookupValue != test.value {
				t.Errorf("Lookup(%q) value = %q, want %q", test.label, gotLookupValue, test.value)
			}
			if gotLookupExists != test.exists {
				t.Errorf("Lookup(%q) exists = %v, want %v", test.label, gotLookupExists, test.exists)
			}
		})
	}
}
