//
// Copyright 2022 The Sigstore Authors.
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

package v1beta1

import (
	"context"

	"knative.dev/pkg/apis"
)

// PodScalableValidator is a callback to validate a PodScalable.
type PodScalableValidator func(context.Context, *PodScalable) *apis.FieldError

// Validate implements apis.Validatable
func (ps *PodScalable) Validate(ctx context.Context) *apis.FieldError {
	if psv := GetPodScalableValidator(ctx); psv != nil {
		return psv(ctx, ps)
	}
	return nil
}

// psvKey is used for associating a PodScalableValidator with a context.Context
type psvKey struct{}

func WithPodScalableValidator(ctx context.Context, psv PodScalableValidator) context.Context {
	return context.WithValue(ctx, psvKey{}, psv)
}

// GetPodScalableValidator extracts the PodSpecValidator from the context.
func GetPodScalableValidator(ctx context.Context) PodScalableValidator {
	untyped := ctx.Value(psvKey{})
	if untyped == nil {
		return nil
	}
	return untyped.(PodScalableValidator)
}
