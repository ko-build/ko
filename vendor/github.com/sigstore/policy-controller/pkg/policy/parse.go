// Copyright 2023 The Sigstore Authors.
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

package policy

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sigstore/policy-controller/pkg/apis/policy/v1alpha1"
	"github.com/sigstore/policy-controller/pkg/apis/policy/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/yaml"
)

// Parse decodes a provided YAML document containing zero or more objects into
// a collection of unstructured.Unstructured objects.
func Parse(ctx context.Context, document string) ([]*unstructured.Unstructured, error) {
	docs := strings.Split(document, "\n---\n")

	objs := make([]*unstructured.Unstructured, 0, len(docs))
	for i, doc := range docs {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}
		var obj unstructured.Unstructured
		if err := yaml.Unmarshal([]byte(doc), &obj); err != nil {
			return nil, fmt.Errorf("decoding object[%d]: %w", i, err)
		}
		if obj.GetAPIVersion() == "" {
			return nil, apis.ErrMissingField("apiVersion").ViaIndex(i)
		}
		if obj.GetName() == "" {
			return nil, apis.ErrMissingField("metadata.name").ViaIndex(i)
		}
		objs = append(objs, &obj)
	}
	return objs, nil
}

// ParseClusterImagePolicies returns ClusterImagePolicy objects found in the
// policy document.
func ParseClusterImagePolicies(ctx context.Context, document string) (cips []*v1alpha1.ClusterImagePolicy, warns error, err error) {
	if warns, err = Validate(ctx, document); err != nil {
		return nil, warns, err
	}

	ol, err := Parse(ctx, document)
	if err != nil {
		// "Validate" above calls "Parse", so this is unreachable.
		return nil, warns, err
	}

	cips = make([]*v1alpha1.ClusterImagePolicy, 0, len(ol))
	for _, obj := range ol {
		gv, err := schema.ParseGroupVersion(obj.GetAPIVersion())
		if err != nil {
			// Practically speaking unstructured.Unstructured won't let this happen.
			return nil, warns, fmt.Errorf("error parsing apiVersion of: %w", err)
		}

		cip := &v1alpha1.ClusterImagePolicy{}

		switch gv.WithKind(obj.GetKind()) {
		case v1beta1.SchemeGroupVersion.WithKind("ClusterImagePolicy"):
			v1b1 := &v1beta1.ClusterImagePolicy{}
			if err := convert(obj, v1b1); err != nil {
				return nil, warns, err
			}
			if err := cip.ConvertFrom(ctx, v1b1); err != nil {
				return nil, warns, err
			}

		case v1alpha1.SchemeGroupVersion.WithKind("ClusterImagePolicy"):
			// This is allowed, but we should convert things.
			if err := convert(obj, cip); err != nil {
				return nil, warns, err
			}

		default:
			continue
		}

		cips = append(cips, cip)
	}
	return cips, warns, nil
}

func convert(from interface{}, to interface{}) error {
	bs, err := json.Marshal(from)
	if err != nil {
		return fmt.Errorf("Marshal() = %w", err)
	}
	if err := json.Unmarshal(bs, to); err != nil {
		return fmt.Errorf("Unmarshal() = %w", err)
	}
	return nil
}
