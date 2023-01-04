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

package cosign

import (
	"crypto/x509"
)

type CertExtensions struct {
	Cert *x509.Certificate
}

var (
	// Fulcio cert-extensions, documented here: https://github.com/sigstore/fulcio/blob/main/docs/oid-info.md
	CertExtensionOIDCIssuer               = "1.3.6.1.4.1.57264.1.1"
	CertExtensionGithubWorkflowTrigger    = "1.3.6.1.4.1.57264.1.2"
	CertExtensionGithubWorkflowSha        = "1.3.6.1.4.1.57264.1.3"
	CertExtensionGithubWorkflowName       = "1.3.6.1.4.1.57264.1.4"
	CertExtensionGithubWorkflowRepository = "1.3.6.1.4.1.57264.1.5"
	CertExtensionGithubWorkflowRef        = "1.3.6.1.4.1.57264.1.6"

	CertExtensionMap = map[string]string{
		CertExtensionOIDCIssuer:               "oidcIssuer",
		CertExtensionGithubWorkflowTrigger:    "githubWorkflowTrigger",
		CertExtensionGithubWorkflowSha:        "githubWorkflowSha",
		CertExtensionGithubWorkflowName:       "githubWorkflowName",
		CertExtensionGithubWorkflowRepository: "githubWorkflowRepository",
		CertExtensionGithubWorkflowRef:        "githubWorkflowRef",
	}
)

func (ce *CertExtensions) certExtensions() map[string]string {
	extensions := map[string]string{}
	for _, ext := range ce.Cert.Extensions {
		readableName, ok := CertExtensionMap[ext.Id.String()]
		if ok {
			extensions[readableName] = string(ext.Value)
		} else {
			extensions[ext.Id.String()] = string(ext.Value)
		}
	}
	return extensions
}

// GetIssuer returns the issuer for a Certificate
func (ce *CertExtensions) GetIssuer() string {
	return ce.certExtensions()["oidcIssuer"]
}

// GetCertExtensionGithubWorkflowTrigger returns the GitHub Workflow Trigger for a Certificate
func (ce *CertExtensions) GetCertExtensionGithubWorkflowTrigger() string {
	return ce.certExtensions()["githubWorkflowTrigger"]
}

// GetExtensionGithubWorkflowSha returns the GitHub Workflow SHA for a Certificate
func (ce *CertExtensions) GetExtensionGithubWorkflowSha() string {
	return ce.certExtensions()["githubWorkflowSha"]
}

// GetCertExtensionGithubWorkflowName returns the GitHub Workflow Name for a Certificate
func (ce *CertExtensions) GetCertExtensionGithubWorkflowName() string {
	return ce.certExtensions()["githubWorkflowName"]
}

// GetCertExtensionGithubWorkflowRepository returns the GitHub Workflow Repository for a Certificate
func (ce *CertExtensions) GetCertExtensionGithubWorkflowRepository() string {
	return ce.certExtensions()["githubWorkflowRepository"]
}

// GetCertExtensionGithubWorkflowRef returns the GitHub Workflow Ref for a Certificate
func (ce *CertExtensions) GetCertExtensionGithubWorkflowRef() string {
	return ce.certExtensions()["githubWorkflowRef"]
}
