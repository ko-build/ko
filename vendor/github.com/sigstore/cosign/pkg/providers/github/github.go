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

package github

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/sigstore/cosign/pkg/providers"
)

func init() {
	providers.Register("github-actions", &githubActions{})
}

type githubActions struct{}

var _ providers.Interface = (*githubActions)(nil)

const (
	RequestTokenEnvKey = "ACTIONS_ID_TOKEN_REQUEST_TOKEN"
	RequestURLEnvKey   = "ACTIONS_ID_TOKEN_REQUEST_URL"
)

// Enabled implements providers.Interface
func (ga *githubActions) Enabled(ctx context.Context) bool {
	if os.Getenv(RequestTokenEnvKey) == "" {
		return false
	}
	if os.Getenv(RequestURLEnvKey) == "" {
		return false
	}
	return true
}

// Provide implements providers.Interface
func (ga *githubActions) Provide(ctx context.Context, audience string) (string, error) {
	url := os.Getenv(RequestURLEnvKey) + "&audience=" + audience

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", "bearer "+os.Getenv(RequestTokenEnvKey))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var payload struct {
		Value string `json:"value"`
	}

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&payload); err != nil {
		return "", err
	}
	return payload.Value, nil
}
