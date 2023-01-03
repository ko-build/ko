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
)

// PodScalableDefaulter is a callback to validate a PodScalable.
type PodScalableDefaulter func(context.Context, *PodScalable)

// SetDefaults implements apis.Defaultable
func (ps *PodScalable) SetDefaults(ctx context.Context) {
	if psd := GetPodScalableDefaulter(ctx); psd != nil {
		psd(ctx, ps)
	}
}

// psdKey is used for associating a PodScalableDefaulter with a context.Context
type psdKey struct{}

func WithPodScalableDefaulter(ctx context.Context, psd PodScalableDefaulter) context.Context {
	return context.WithValue(ctx, psdKey{}, psd)
}

// GetPodScalableDefaulter extracts the PodScalableDefaulter from the context.
func GetPodScalableDefaulter(ctx context.Context) PodScalableDefaulter {
	untyped := ctx.Value(psdKey{})
	if untyped == nil {
		return nil
	}
	return untyped.(PodScalableDefaulter)
}
