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

package options

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
)

const DefaultOIDCIssuerURL = "https://oauth2.sigstore.dev/auth"

// OIDCOptions is the wrapper for OIDC related options.
type OIDCOptions struct {
	Issuer                  string
	ClientID                string
	clientSecretFile        string
	RedirectURL             string
	Provider                string
	DisableAmbientProviders bool
}

func (o *OIDCOptions) ClientSecret() (string, error) {
	if o.clientSecretFile != "" {
		clientSecretBytes, err := os.ReadFile(o.clientSecretFile)
		if err != nil {
			return "", fmt.Errorf("reading OIDC client secret: %w", err)
		}
		if !utf8.Valid(clientSecretBytes) {
			return "", fmt.Errorf("OIDC client secret in file %s not valid utf8", o.clientSecretFile)
		}
		clientSecretString := string(clientSecretBytes)
		clientSecretString = strings.TrimSpace(clientSecretString)
		return clientSecretString, nil
	}
	return "", nil
}

var _ Interface = (*OIDCOptions)(nil)

// AddFlags implements Interface
func (o *OIDCOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Issuer, "oidc-issuer", DefaultOIDCIssuerURL,
		"[EXPERIMENTAL] OIDC provider to be used to issue ID token")

	cmd.Flags().StringVar(&o.ClientID, "oidc-client-id", "sigstore",
		"[EXPERIMENTAL] OIDC client ID for application")

	cmd.Flags().StringVar(&o.clientSecretFile, "oidc-client-secret-file", "",
		"[EXPERIMENTAL] Path to file containing OIDC client secret for application")
	_ = cmd.Flags().SetAnnotation("oidc-client-secret-file", cobra.BashCompFilenameExt, []string{})

	cmd.Flags().StringVar(&o.RedirectURL, "oidc-redirect-url", "",
		"[EXPERIMENTAL] OIDC redirect URL (Optional). The default oidc-redirect-url is 'http://localhost:0/auth/callback'.")

	cmd.Flags().StringVar(&o.Provider, "oidc-provider", "",
		"[EXPERIMENTAL] Specify the provider to get the OIDC token from (Optional). If unset, all options will be tried. Options include: [spiffe, google, github, filesystem]")

	cmd.Flags().BoolVar(&o.DisableAmbientProviders, "oidc-disable-ambient-providers", false,
		"[EXPERIMENTAL] Disable ambient OIDC providers. When true, ambient credentials will not be read")
}
