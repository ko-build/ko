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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/sigstore/policy-controller/pkg/apis/policy/v1alpha1"
	"github.com/sigstore/policy-controller/pkg/apis/policy/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"knative.dev/pkg/apis"
)

var (
	// ErrEmptyDocument is the error returned when no document body is
	// specified.
	ErrEmptyDocument = errors.New("document is required to create policy")

	// ErrUnknownType is the error returned when a type contained in the policy
	// is unrecognized.
	ErrUnknownType = errors.New("unknown type")
)

// Validate decodes a provided YAML document containing zero or more objects
// and performs limited validation on them.
func Validate(ctx context.Context, document string) (warns error, err error) {
	if len(document) == 0 {
		return nil, ErrEmptyDocument
	}

	uol, err := Parse(ctx, document)
	if err != nil {
		return nil, err
	}

	for i, uo := range uol {
		switch uo.GroupVersionKind() {
		case v1beta1.SchemeGroupVersion.WithKind("ClusterImagePolicy"):
			if warns, err = validate(ctx, uo, &v1beta1.ClusterImagePolicy{}); err != nil {
				return
			}

		case v1alpha1.SchemeGroupVersion.WithKind("ClusterImagePolicy"):
			if warns, err = validate(ctx, uo, &v1alpha1.ClusterImagePolicy{}); err != nil {
				return
			}

		case corev1.SchemeGroupVersion.WithKind("Secret"):
			if uo.GetNamespace() != "cosign-system" {
				return warns, apis.ErrInvalidValue(uo.GetNamespace(), "metadata.namespace").ViaIndex(i)
			}
			// Any additional validation worth performing?  Should we check the
			// schema of the secret matches the expectations of cosigned?

		default:
			return warns, fmt.Errorf("%w: %v", ErrUnknownType, uo.GroupVersionKind())
		}
	}
	return warns, nil
}

type crd interface {
	apis.Validatable
	apis.Defaultable
}

func validate(ctx context.Context, uo *unstructured.Unstructured, v crd) (warns error, err error) {
	b, err := json.Marshal(uo)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal: %w", err)
	}

	dec := json.NewDecoder(bytes.NewBuffer(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return nil, fmt.Errorf("unable to unmarshal: %w", err)
	}

	// Apply defaulting to simulate the defaulting webhook that runs prior
	// to validation.
	v.SetDefaults(ctx)

	// We can't just return v.Validate(ctx) because of Go's typed nils.
	// nolint:revive
	if ve := v.Validate(ctx); ve != nil {
		// Separate validation warnings from errors so the caller can discern between them.
		if warnFE := ve.Filter(apis.WarningLevel); warnFE != nil {
			warns = warnFE
		}
		if errorFE := ve.Filter(apis.ErrorLevel); errorFE != nil {
			err = errorFE
		}
	}
	return
}
