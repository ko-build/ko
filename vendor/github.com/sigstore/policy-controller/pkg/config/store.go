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

package config

import (
	"context"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/configmap"
)

type cfgKey struct{}

const (
	// PolicyControllerConfigName is the name of the configmap used to configure
	// policy-controller.
	PolicyControllerConfigName = "config-policy-controller"

	// Specifies that if an image is not found to match any policy, it should
	// be rejected.
	DenyAll = "deny"

	// Specifies that if an image is not found to match any policy, it should
	// be allowed.
	AllowAll = "allow"

	WarnAll = "warn"

	NoMatchPolicyKey = "no-match-policy"

	FailOnEmptyAuthorities = "fail-on-empty-authorities"
)

// PolicyControllerConfig controls the behaviour of policy-controller that needs
// to be more flexible than requiring a controller restart. Some examples are
// controlling behaviour for what to do if no matching policies are found.
// Point is that these apply to the whole controller instead of specific CIP
// policies that apply only to matching images.
type PolicyControllerConfig struct {
	// NoMatchPolicy says what do in the case where an image does not match
	// any policy.
	NoMatchPolicy string `json:"no-match-policy"`
	// FailOnEmptyAuthorities configures the validating webhook to allow creating CIP without a list authorities
	FailOnEmptyAuthorities bool `json:"fail-on-empty-authorities"`
}

func NewPolicyControllerConfigFromMap(data map[string]string) (*PolicyControllerConfig, error) {
	ret := &PolicyControllerConfig{NoMatchPolicy: "deny", FailOnEmptyAuthorities: true}
	switch data[NoMatchPolicyKey] {
	case DenyAll:
		ret.NoMatchPolicy = DenyAll
	case AllowAll:
		ret.NoMatchPolicy = AllowAll
	case WarnAll:
		ret.NoMatchPolicy = WarnAll
	default:
		ret.NoMatchPolicy = DenyAll
	}
	if val, ok := data[FailOnEmptyAuthorities]; ok {
		var err error
		ret.FailOnEmptyAuthorities, err = strconv.ParseBool(val)
		return ret, err
	}
	ret.FailOnEmptyAuthorities = true
	return ret, nil
}

func NewPolicyControllerConfigFromConfigMap(config *corev1.ConfigMap) (*PolicyControllerConfig, error) {
	return NewPolicyControllerConfigFromMap(config.Data)
}

// FromContext extracts a PolicyControllerConfig from the provided context.
func FromContext(ctx context.Context) *PolicyControllerConfig {
	x, ok := ctx.Value(cfgKey{}).(*PolicyControllerConfig)
	if ok {
		return x
	}
	return nil
}

// FromContextOrDefaults is like FromContext, but when no
// PolicyControllerConfig is attached, it returns a PolicyControllerConfig
// populated with the defaults for each of the fields.
func FromContextOrDefaults(ctx context.Context) *PolicyControllerConfig {
	if cfg := FromContext(ctx); cfg != nil {
		return cfg
	}
	return &PolicyControllerConfig{
		NoMatchPolicy:          DenyAll,
		FailOnEmptyAuthorities: true,
	}
}

// ToContext attaches the provided PolicyControllerConfig to the provided
// context, returning the new context with the Config attached.
func ToContext(ctx context.Context, c *PolicyControllerConfig) context.Context {
	return context.WithValue(ctx, cfgKey{}, c)
}

// Store is a typed wrapper around configmap.Untyped store to handle our configmaps.
// +k8s:deepcopy-gen=false
type Store struct {
	*configmap.UntypedStore
}

// NewStore creates a new store of Configs and optionally calls functions when ConfigMaps are updated.
func NewStore(logger configmap.Logger, onAfterStore ...func(name string, value interface{})) *Store {
	store := &Store{
		UntypedStore: configmap.NewUntypedStore(
			PolicyControllerConfigName,
			logger,
			configmap.Constructors{
				PolicyControllerConfigName: NewPolicyControllerConfigFromConfigMap,
			},
			onAfterStore...,
		),
	}

	return store
}

// ToContext attaches the current PolicyControllerConfig state to the provided
// context.
func (s *Store) ToContext(ctx context.Context) context.Context {
	return ToContext(ctx, s.Load())
}

// Load creates a PolicyControllerConfig from the current config state of the
// Store.
func (s *Store) Load() *PolicyControllerConfig {
	return s.UntypedLoad(PolicyControllerConfigName).(*PolicyControllerConfig)
}
