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
	"io"
	"net/http"
	"os"

	"k8s.io/apimachinery/pkg/util/sets"
	"knative.dev/pkg/apis"
)

type Verification struct {
	// NoMatchPolicy specifies the behavior when a base image doesn't match any
	// of the listed policies.  It allows the values: allow, deny, and warn.
	NoMatchPolicy string `yaml:"no-match-policy,omitempty"`

	// Policies specifies a collection of policies to use to cover the base
	// images used as part of evaluation.  See "policy" below for usage.
	// Policies can be nil so that we can distinguish between an explicitly
	// specified empty list and when policies is unspecified.
	Policies *[]Source `yaml:"policies,omitempty"`
}

// Source contains a set of options for specifying policies.  Exactly
// one of the fields may be specified for each Source entry.
type Source struct {
	// Data is a collection of one or more ClusterImagePolicy resources.
	Data string `yaml:"data,omitempty"`

	// Path is a path to a file containing one or more ClusterImagePolicy
	// resources.
	Path string `yaml:"path,omitempty"`

	// URL links to a file containing one or more ClusterImagePolicy resources.
	URL string `yaml:"url,omitempty"`
}

func (v *Verification) Validate(ctx context.Context) (errs *apis.FieldError) {
	switch v.NoMatchPolicy {
	case "allow", "deny", "warn":
		// Good!
	case "":
		errs = errs.Also(apis.ErrMissingField("noMatchPolicy"))
	default:
		errs = errs.Also(apis.ErrInvalidValue(v.NoMatchPolicy, "noMatchPolicy"))
	}

	if v.Policies == nil {
		errs = errs.Also(apis.ErrMissingField("policies"))
	} else {
		for i, p := range *v.Policies {
			errs = errs.Also(p.Validate(ctx).ViaFieldIndex("policies", i))
		}
	}

	return errs
}

func (pd *Source) Validate(ctx context.Context) (errs *apis.FieldError) {
	// Check that exactly one of the fields is set.
	set := sets.NewString()
	if pd.Data != "" {
		set.Insert("data")
		_, _, err := ParseClusterImagePolicies(ctx, pd.Data)
		if err != nil {
			errs = errs.Also(apis.ErrInvalidValue(err.Error(), "data"))
		}
	}
	if pd.Path != "" {
		set.Insert("path")
		raw, err := os.ReadFile(pd.Path)
		if err != nil {
			errs = errs.Also(&apis.FieldError{
				Message: err.Error(),
				Paths:   []string{"path"},
			})
		} else {
			_, _, err := ParseClusterImagePolicies(ctx, string(raw))
			if err != nil {
				errs = errs.Also(apis.ErrInvalidValue(err.Error(), "path"))
			}
		}
	}
	if pd.URL != "" {
		set.Insert("url")
		resp, err := http.Get(pd.URL)
		if err != nil {
			errs = errs.Also(&apis.FieldError{
				Message: err.Error(),
				Paths:   []string{"url"},
			})
		} else {
			defer resp.Body.Close()
			raw, err := io.ReadAll(resp.Body)
			if err != nil {
				errs = errs.Also(&apis.FieldError{
					Message: err.Error(),
					Paths:   []string{"url"},
				})
			} else {
				_, _, err := ParseClusterImagePolicies(ctx, string(raw))
				if err != nil {
					errs = errs.Also(apis.ErrInvalidValue(err.Error(), "url"))
				}
			}
		}
	}

	switch set.Len() {
	case 0:
		errs = errs.Also(apis.ErrMissingOneOf("data", "path", "url"))
	case 1:
		// What we want.
	default:
		// This will be unreachable until we add more than one thing
		// to our oneof.
		errs = errs.Also(apis.ErrMultipleOneOf(set.List()...))
	}
	return errs
}
