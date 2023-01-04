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
	"fmt"

	"knative.dev/pkg/apis"
)

// SetDefaults implements apis.Defaultable
func (c *ClusterImagePolicy) SetDefaults(ctx context.Context) {
	c.Spec.SetDefaults(ctx)
}

func (spec *ClusterImagePolicySpec) SetDefaults(ctx context.Context) {
	if spec.Mode == "" {
		spec.Mode = "enforce"
	}
	for i, authority := range spec.Authorities {
		if authority.Name == "" {
			spec.Authorities[i].Name = fmt.Sprintf("authority-%d", i)
		}
		if authority.Key == nil && authority.Static == nil && authority.Keyless != nil && authority.Keyless.CACert == nil && authority.Keyless.URL == nil {
			authority.Keyless.URL = apis.HTTPS("fulcio.sigstore.dev")
		}
	}
}
