//
// Copyright 2021 The Sigstore Authors.
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

package providers

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var (
	m         sync.Mutex
	providers = make(map[string]Interface)
)

// Interface is what providers need to implement to participate in furnishing OIDC tokens.
type Interface interface {
	// Enabled returns true if the provider is enabled.
	Enabled(ctx context.Context) bool

	// Provide returns an OIDC token scoped to the provided audience.
	Provide(ctx context.Context, audience string) (string, error)
}

// Register is used by providers to participate in furnishing OIDC tokens.
func Register(name string, p Interface) {
	m.Lock()
	defer m.Unlock()

	if prev, ok := providers[name]; ok {
		panic(fmt.Sprintf("duplicate provider for name %q, %T and %T", name, prev, p))
	}
	providers[name] = p
}

// Enabled checks whether any of the registered providers are enabled in this execution context.
func Enabled(ctx context.Context) bool {
	m.Lock()
	defer m.Unlock()

	for _, provider := range providers {
		if provider.Enabled(ctx) {
			return true
		}
	}
	return false
}

// Provide fetches an OIDC token from one of the active providers.
func Provide(ctx context.Context, audience string) (string, error) {
	m.Lock()
	defer m.Unlock()

	var id string
	var err error
	for _, provider := range providers {
		if !provider.Enabled(ctx) {
			continue
		}
		id, err = provider.Provide(ctx, audience)
		if err == nil {
			return id, err
		}
	}
	// return the last id/err combo, unless there wasn't an error in
	// which case provider.Enabled() wasn't checked.
	if err == nil {
		err = errors.New("no providers are enabled, check providers.Enabled() before providers.Provide()")
	}
	return id, err
}
