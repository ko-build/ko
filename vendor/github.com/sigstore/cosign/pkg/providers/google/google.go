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

package google

import (
	"context"
	"io/ioutil"
	"os"
	"strings"

	"google.golang.org/api/idtoken"
	"google.golang.org/api/impersonate"

	"github.com/sigstore/cosign/pkg/providers"
)

func init() {
	providers.Register("google-workload-identity", &googleWorkloadIdentity{})
	providers.Register("google-impersonate", &googleImpersonate{})
}

type googleWorkloadIdentity struct{}

var _ providers.Interface = (*googleWorkloadIdentity)(nil)

// gceProductNameFile is the product file path that contains the cloud service name.
// This is a variable instead of a const to enable testing.
var gceProductNameFile = "/sys/class/dmi/id/product_name"

// Enabled implements providers.Interface
// This is based on k8s.io/kubernetes/pkg/credentialprovider/gcp
func (gwi *googleWorkloadIdentity) Enabled(ctx context.Context) bool {
	data, err := ioutil.ReadFile(gceProductNameFile)
	if err != nil {
		return false
	}
	name := strings.TrimSpace(string(data))
	if name == "Google" || name == "Google Compute Engine" {
		// Just because we're on Google, does not mean workload identity is available.
		// TODO(mattmoor): do something better than this.
		_, err := gwi.Provide(ctx, "garbage")
		return err == nil
	}
	return false
}

// Provide implements providers.Interface
func (gwi *googleWorkloadIdentity) Provide(ctx context.Context, audience string) (string, error) {
	ts, err := idtoken.NewTokenSource(ctx, audience)
	if err != nil {
		return "", err
	}
	tok, err := ts.Token()
	if err != nil {
		return "", err
	}
	return tok.AccessToken, nil
}

type googleImpersonate struct{}

var _ providers.Interface = (*googleImpersonate)(nil)

// Enabled implements providers.Interface
func (gi *googleImpersonate) Enabled(ctx context.Context) bool {
	// The "impersonate" method requires a target service account to impersonate.
	return os.Getenv("GOOGLE_SERVICE_ACCOUNT_NAME") != ""
}

// Provide implements providers.Interface
func (gi *googleImpersonate) Provide(ctx context.Context, audience string) (string, error) {
	target := os.Getenv("GOOGLE_SERVICE_ACCOUNT_NAME")
	ts, err := impersonate.IDTokenSource(ctx, impersonate.IDTokenConfig{
		Audience:        "sigstore",
		TargetPrincipal: target,
		IncludeEmail:    true,
	})
	if err != nil {
		return "", err
	}
	tok, err := ts.Token()
	if err != nil {
		return "", err
	}
	return tok.AccessToken, nil
}
