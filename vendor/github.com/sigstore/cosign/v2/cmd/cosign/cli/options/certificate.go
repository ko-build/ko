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

package options

import (
	"errors"

	"github.com/sigstore/cosign/v2/pkg/cosign"
	"github.com/spf13/cobra"
)

// CertVerifyOptions is the wrapper for certificate verification.
type CertVerifyOptions struct {
	Cert                         string
	CertIdentity                 string
	CertIdentityRegexp           string
	CertOidcIssuer               string
	CertOidcIssuerRegexp         string
	CertGithubWorkflowTrigger    string
	CertGithubWorkflowSha        string
	CertGithubWorkflowName       string
	CertGithubWorkflowRepository string
	CertGithubWorkflowRef        string
	CertChain                    string
	SCT                          string
	IgnoreSCT                    bool
}

var _ Interface = (*RekorOptions)(nil)

// AddFlags implements Interface
func (o *CertVerifyOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Cert, "certificate", "",
		"path to the public certificate. The certificate will be verified against the Fulcio roots if the --certificate-chain option is not passed.")
	_ = cmd.Flags().SetAnnotation("certificate", cobra.BashCompFilenameExt, []string{"cert"})

	cmd.Flags().StringVar(&o.CertIdentity, "certificate-identity", "",
		"The identity expected in a valid Fulcio certificate. Valid values include email address, DNS names, IP addresses, and URIs. Either --certificate-identity or --certificate-identity-regexp must be set for keyless flows.")

	cmd.Flags().StringVar(&o.CertIdentityRegexp, "certificate-identity-regexp", "",
		"A regular expression alternative to --certificate-identity. Accepts the Go regular expression syntax described at https://golang.org/s/re2syntax. Either --certificate-identity or --certificate-identity-regexp must be set for keyless flows.")

	cmd.Flags().StringVar(&o.CertOidcIssuer, "certificate-oidc-issuer", "",
		"The OIDC issuer expected in a valid Fulcio certificate, e.g. https://token.actions.githubusercontent.com or https://oauth2.sigstore.dev/auth. Either --certificate-oidc-issuer or --certificate-oidc-issuer-regexp must be set for keyless flows.")

	cmd.Flags().StringVar(&o.CertOidcIssuerRegexp, "certificate-oidc-issuer-regexp", "",
		"A regular expression alternative to --certificate-oidc-issuer. Accepts the Go regular expression syntax described at https://golang.org/s/re2syntax. Either --certificate-oidc-issuer or --certificate-oidc-issuer-regexp must be set for keyless flows.")

	// -- Cert extensions begin --
	// Source: https://github.com/sigstore/fulcio/blob/main/docs/oid-info.md
	cmd.Flags().StringVar(&o.CertGithubWorkflowTrigger, "certificate-github-workflow-trigger", "",
		"contains the event_name claim from the GitHub OIDC Identity token that contains the name of the event that triggered the workflow run")

	cmd.Flags().StringVar(&o.CertGithubWorkflowSha, "certificate-github-workflow-sha", "",
		"contains the sha claim from the GitHub OIDC Identity token that contains the commit SHA that the workflow run was based upon.")

	cmd.Flags().StringVar(&o.CertGithubWorkflowName, "certificate-github-workflow-name", "",
		"contains the workflow claim from the GitHub OIDC Identity token that contains the name of the executed workflow.")

	cmd.Flags().StringVar(&o.CertGithubWorkflowRepository, "certificate-github-workflow-repository", "",
		"contains the repository claim from the GitHub OIDC Identity token that contains the repository that the workflow run was based upon")

	cmd.Flags().StringVar(&o.CertGithubWorkflowRef, "certificate-github-workflow-ref", "",
		"contains the ref claim from the GitHub OIDC Identity token that contains the git ref that the workflow run was based upon.")
	// -- Cert extensions end --
	cmd.Flags().StringVar(&o.CertChain, "certificate-chain", "",
		"path to a list of CA certificates in PEM format which will be needed "+
			"when building the certificate chain for the signing certificate. "+
			"Must start with the parent intermediate CA certificate of the "+
			"signing certificate and end with the root certificate")
	_ = cmd.Flags().SetAnnotation("certificate-chain", cobra.BashCompFilenameExt, []string{"cert"})

	cmd.Flags().StringVar(&o.SCT, "sct", "",
		"path to a detached Signed Certificate Timestamp, formatted as a RFC6962 AddChainResponse struct. "+
			"If a certificate contains an SCT, verification will check both the detached and embedded SCTs.")
	cmd.Flags().BoolVar(&o.IgnoreSCT, "insecure-ignore-sct", false,
		"when set, verification will not check that a certificate contains an embedded SCT, a proof of "+
			"inclusion in a certificate transparency log")
}

func (o *CertVerifyOptions) Identities() ([]cosign.Identity, error) {
	if o.CertIdentity == "" && o.CertIdentityRegexp == "" {
		return nil, errors.New("--certificate-identity or --certificate-identity-regexp is required for verification in keyless mode")
	}
	if o.CertOidcIssuer == "" && o.CertOidcIssuerRegexp == "" {
		return nil, errors.New("--certificate-oidc-issuer or --certificate-oidc-issuer-regexp is required for verification in keyless mode")
	}
	return []cosign.Identity{{IssuerRegExp: o.CertOidcIssuerRegexp, Issuer: o.CertOidcIssuer, SubjectRegExp: o.CertIdentityRegexp, Subject: o.CertIdentity}}, nil
}
