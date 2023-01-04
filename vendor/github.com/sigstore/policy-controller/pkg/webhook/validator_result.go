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

package webhook

import (
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

// PolicyResult is the result of a successful ValidatePolicy call.
// These are meant to be consumed by a higher level Policy engine that
// can reason about validated results. The 'first' level pass will verify
// signatures and attestations, and make the results then available for
// a policy that can be used to gate a passing of a ClusterImagePolicy.
// Some examples are, at least 'vulnerability' has to have been done
// and the scan must have been attested by a particular entity (sujbect/issuer)
// or a particular key.
// Other examples are N-of-M must be satisfied and so forth.
// We do not expose the low level details of signatures / attestations here
// since they have already been validated as per the Authority configuration
// and optionally by the Attestations which contain a particular policy that
// can be used to validate the Attestations (say vulnerability scanner must not
// have any High sev issues).
type PolicyResult struct {
	// AuthorityMatches will have an entry for each successful Authority check
	// on it. Key in the map is the Attestation.Name
	AuthorityMatches map[string]AuthorityMatch `json:"authorityMatches,omitempty"`

	// Config contains the Config for each of the normalized os/architectures
	// where key to the map is the {OS}/{Architecture}[/{Variant}]
	//
	// Some examples are:
	// linux/arm64
	// linux/arm/v7
	// linux/arm/v6
	//
	// This field is only available for evaluation if
	// CIP.Spec.Policy.FetchConfigFile is set to true.
	Config map[string]*v1.ConfigFile `json:"config,omitempty"`

	// Spec contains the Spec for the resource that was evaluated. Note
	// that because this is resource specific, so you can use MatchResource
	// to filter to only specific resource to get only the Specs you want.
	//
	// This field is only available for evaluation if
	// CIP.Spec.Policy.IncludeSpec is set to true.
	Spec interface{} `json:"spec,omitempty"`

	// ObjectMeta contains the ObjectMeta for the resource that was evaluated.
	//
	// This field is only available for evaluation if
	// CIP.Spec.Policy.IncludeObjectMeta is set to true.
	ObjectMeta interface{} `json:"metadata,omitempty"`

	// TypeMeta contains the TypeMeta for the resource that was evaluated.
	//
	// This field is only available for evaluation if
	// CIP.Spec.Policy.IncludeTypeMeta is set to true.
	TypeMeta interface{} `json:"typemeta,omitempty"`
}

// AuthorityMatch returns either Signatures (if there are no Attestations
// specified), or Attestations if there are Attestations specified.
type AuthorityMatch struct {
	// All of the matching signatures for this authority
	// Wonder if for consistency this should also have the matching
	// attestations name, aka, make this into a map.
	Signatures []PolicySignature `json:"signatures,omitempty"`

	// Mapping from attestation name to all of verified attestations
	Attestations map[string][]PolicyAttestation `json:"attestations,omitempty"`

	// Static indicates whether this authority matched due to static
	// e.g. static: { action: pass }
	Static bool `json:"static,omitempty"`
}

// PolicySignature contains a normalized result of a validated signature, where
// signature could be a signature on the Image (.sig) or on an Attestation
// (.att).
type PolicySignature struct {
	// A unique identifier describing this signature.
	// This is typically the hash of this signature's OCI layer for images.
	ID string `json:"id,omitempty"`

	// Subject that was found to match on the Cert.
	Subject string `json:"subject,omitempty"`
	// Issure that was found to match on the Cert.
	Issuer string `json:"issuer,omitempty"`

	// GithubExtensions holds the Github-related OID extensions.
	// See also: https://github.com/sigstore/fulcio/blob/main/docs/oid-info.md
	GithubExtensions `json:",inline"`
}

// PolicyAttestation contains a normalized result of a validated attestation,
// which consists of the PolicySignature part, and some additional attestation
// specific fields.
type PolicyAttestation struct {
	PolicySignature `json:",inline"`

	// PredicateType is the in-toto predicate type of this attestation.
	PredicateType string `json:"predicateType,omitempty"`

	// Payload is the bytes of the in-toto statement's predicate payload.
	// This is included for the benefit of the caller of ValidatePolicy, and is
	// not intended for consumption in the ClusterImagePolicy's outer policy
	// block.
	Payload []byte `json:"-"`
}

// GithubExtensions holds the Github-related OID extensions.
// See also: https://github.com/sigstore/fulcio/blob/main/docs/oid-info.md
// NOTE: these field correlate with the names given in the cosign
// CertExtensionMap and must be prefixed with "github" to avoid ambiguity.
type GithubExtensions struct {
	// OID: 1.3.6.1.4.1.57264.1.2
	WorkflowTrigger string `json:"githubWorkflowTrigger,omitempty"`
	// OID: 1.3.6.1.4.1.57264.1.3
	WorkflowSHA string `json:"githubWorkflowSha,omitempty"`
	// OID: 1.3.6.1.4.1.57264.1.4
	WorkflowName string `json:"githubWorkflowName,omitempty"`
	// OID: 1.3.6.1.4.1.57264.1.5
	WorkflowRepo string `json:"githubWorkflowRepo,omitempty"`
	// OID: 1.3.6.1.4.1.57264.1.6
	WorkflowRef string `json:"githubWorkflowRef,omitempty"`
}
