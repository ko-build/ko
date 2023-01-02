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

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/kmeta"
)

// ClusterImagePolicy defines the images that go through verification
// and the authorities used for verification
//
// +genclient
// +genclient:nonNamespaced
// +genclient:noStatus
// +genreconciler:krshapedlogic=false

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ClusterImagePolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	// Spec holds the desired state of the ClusterImagePolicy (from the client).
	Spec ClusterImagePolicySpec `json:"spec"`
}

var (
	_ apis.Validatable   = (*ClusterImagePolicy)(nil)
	_ apis.Defaultable   = (*ClusterImagePolicy)(nil)
	_ kmeta.OwnerRefable = (*ClusterImagePolicy)(nil)
)

// GetGroupVersionKind implements kmeta.OwnerRefable
func (c *ClusterImagePolicy) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("ClusterImagePolicy")
}

// ClusterImagePolicySpec defines a list of images that should be verified
type ClusterImagePolicySpec struct {
	// Images defines the patterns of image names that should be subject to this policy.
	Images []ImagePattern `json:"images"`
	// Authorities defines the rules for discovering and validating signatures.
	// +optional
	Authorities []Authority `json:"authorities,omitempty"`
	// Policy is an optional policy that can be applied against all the
	// successfully validated Authorities. If no authorities pass, this does
	// not even get evaluated, as the Policy is considered failed.
	// +optional
	Policy *Policy `json:"policy,omitempty"`
	// Mode controls whether a failing policy will be rejected (not admitted),
	// or if errors are converted to Warnings.
	// enforce - Reject (default)
	// warn - allow but warn
	// +optional
	Mode string `json:"mode,omitempty"`
	// Match allows selecting resources based on their properties.
	// +optional
	Match []MatchResource `json:"match,omitempty"`
}

// ImagePattern defines a pattern and its associated authorties
// If multiple patterns match a particular image, then ALL of
// those authorities must be satisfied for the image to be admitted.
type ImagePattern struct {
	// Glob defines a globbing pattern.
	Glob string `json:"glob"`
}

// The authorities block defines the rules for discovering and
// validating signatures.  Signatures are
// cryptographically verified using one of the "key" or "keyless"
// fields.
// When multiple authorities are specified, any of them may be used
// to source the valid signature we are looking for to admit an
// image.

type Authority struct {
	// Name is the name for this authority. Used by the CIP Policy
	// validator to be able to reference matching signature or attestation
	// verifications.
	// If not specified, the name will be authority-<index in array>
	Name string `json:"name"`
	// Key defines the type of key to validate the image.
	// +optional
	Key *KeyRef `json:"key,omitempty"`
	// Keyless sets the configuration to verify the authority against a Fulcio instance.
	// +optional
	Keyless *KeylessRef `json:"keyless,omitempty"`
	// Static specifies that signatures / attestations are not validated but
	// instead a static policy is applied against matching images.
	// +optional
	Static *StaticRef `json:"static,omitempty"`
	// Sources sets the configuration to specify the sources from where to consume the signatures.
	// +optional
	Sources []Source `json:"source,omitempty"`
	// CTLog sets the configuration to verify the authority against a Rekor instance.
	// +optional
	CTLog *TLog `json:"ctlog,omitempty"`
	// Attestations is a list of individual attestations for this authority,
	// once the signature for this authority has been verified.
	// +optional
	Attestations []Attestation `json:"attestations,omitempty"`
	// RFC3161Timestamp sets the configuration to verify the signature timestamp against a RFC3161 time-stamping instance.
	// +optional
	RFC3161Timestamp *RFC3161Timestamp `json:"rfc3161timestamp,omitempty"`
}

// This references a public verification key stored in
// a secret in the cosign-system namespace.
// A KeyRef must specify only one of SecretRef, Data or KMS
type KeyRef struct {
	// SecretRef sets a reference to a secret with the key.
	// +optional
	SecretRef *v1.SecretReference `json:"secretRef,omitempty"`
	// Data contains the inline public key
	// +optional
	Data string `json:"data,omitempty"`
	// KMS contains the KMS url of the public key
	// Supported formats differ based on the KMS system used.
	// +optional
	KMS string `json:"kms,omitempty"`
	// HashAlgorithm always defaults to sha256 if the algorithm hasn't been explicitly set
	// +optional
	HashAlgorithm string `json:"hashAlgorithm,omitempty"`
}

// StaticRef specifies that signatures / attestations are not validated but
// instead a static policy is applied against matching images.
type StaticRef struct {
	// Action defines how to handle a matching policy.
	Action string `json:"action"`
}

// Source specifies the location of the signature
type Source struct {
	// OCI defines the registry from where to pull the signatures.
	// +optional
	OCI string `json:"oci,omitempty"`
	// SignaturePullSecrets is an optional list of references to secrets in the
	// same namespace as the deploying resource for pulling any of the signatures
	// used by this Source.
	// +optional
	SignaturePullSecrets []v1.LocalObjectReference `json:"signaturePullSecrets,omitempty"`
}

// TLog specifies the URL to a transparency log that holds
// the signature and public key information
type TLog struct {
	// URL sets the url to the rekor instance (by default the public rekor.sigstore.dev)
	// +optional
	URL *apis.URL `json:"url,omitempty"`
	// Use the Public Key from the referred TrustRoot.TLog
	// +optional
	TrustRootRef string `json:"trustRootRef,omitempty"`
}

// KeylessRef contains location of the validating certificate and the identities
// against which to verify. KeylessRef will contain either the URL to the verifying
// certificate, or it will contain the certificate data inline or in a secret.
type KeylessRef struct {
	// URL defines a url to the keyless instance.
	// +optional
	URL *apis.URL `json:"url,omitempty"`
	// Identities sets a list of identities.
	// +optional
	Identities []Identity `json:"identities,omitempty"`
	// CACert sets a reference to CA certificate
	// +optional
	CACert *KeyRef `json:"ca-cert,omitempty"`
	// Use the Certificate Chain from the referred TrustRoot.CertificateAuthorities and TrustRoot.CTLog
	// +optional
	TrustRootRef string `json:"trustRootRef,omitempty"`
}

// Attestation defines the type of attestation to validate and optionally
// apply a policy decision to it. Authority block is used to verify the
// specified attestation types, and if Policy is specified, then it's applied
// only after the validation of the Attestation signature has been verified.
type Attestation struct {
	// Name of the attestation. These can then be referenced at the CIP level
	// policy.
	Name string `json:"name"`
	// PredicateType defines which predicate type to verify. Matches cosign verify-attestation options.
	PredicateType string `json:"predicateType"`
	// Policy defines all of the matching signatures, and all of
	// the matching attestations (whose attestations are verified).
	// +optional
	Policy *Policy `json:"policy,omitempty"`
}

// MatchResource allows selecting resources based on its version, group and resource.
// It is also possible to select resources based on a list of matching labels.
type MatchResource struct {
	// +optional
	metav1.GroupVersionResource `json:",inline"`
	// +optional
	ResourceSelector *metav1.LabelSelector `json:"selector,omitempty"`
}

// Policy specifies a policy to use for Attestation or the CIP validation (iff
// at least one authority matches).
// Exactly one of Data, URL, or ConfigMapReference must be specified.
type Policy struct {
	// Which kind of policy this is, currently only rego or cue are supported.
	// Furthermore, only cue is tested :)
	Type string `json:"type"`
	// Data contains the policy definition.
	// +optional
	Data string `json:"data,omitempty"`
	// URL to the policy data.
	// +optional
	URL *apis.URL `json:"url,omitempty"`
	// ConfigMapRef defines the reference to a configMap with the policy definition.
	// +optional
	ConfigMapRef *ConfigMapReference `json:"configMapRef,omitempty"`
	// FetchConfigFile controls whether ConfigFile will be fetched and made
	// available for CIP level policy evaluation. Note that this only gets
	// evaluated (and hence fetched) iff at least one authority matches.
	// The ConfigFile will then be available in this format:
	// https://github.com/opencontainers/image-spec/blob/main/config.md
	// +optional
	FetchConfigFile *bool `json:"fetchConfigFile,omitempty"`
	// IncludeSpec controls whether resource `Spec` will be included and
	// made available for CIP level policy evaluation. Note that this only gets
	// evaluated iff at least one authority matches.
	// Also note that because Spec may be of a different shape depending
	// on the resource being evaluatied (see MatchResource for filtering)
	// you might want to configure these to match the policy file to ensure
	// the shape of the Spec is what you expect when evaling the policy.
	// +optional
	IncludeSpec *bool `json:"includeSpec,omitempty"`
	// IncludeObjectMeta controls whether the ObjectMeta will be included and
	// made available for CIP level policy evalutation. Note that this only gets
	// evaluated iff at least one authority matches.
	// +optional
	IncludeObjectMeta *bool `json:"includeObjectMeta,omitempty"`
	// IncludeTypeMeta controls whether the TypeMeta will be included and
	// made available for CIP level policy evalutation. Note that this only gets
	// evaluated iff at least one authority matches.
	// +optional
	IncludeTypeMeta *bool `json:"includeTypeMeta,omitempty"`
}

// ConfigMapReference is cut&paste from SecretReference, but for the life of me
// couldn't find one in the public types. If there's one, use it.
type ConfigMapReference struct {
	// Name is unique within a namespace to reference a configmap resource.
	// +optional
	Name string `json:"name,omitempty"`
	// Namespace defines the space within which the configmap name must be unique.
	// +optional
	Namespace string `json:"namespace,omitempty"`
	// Key defines the key to pull from the configmap.
	// +optional
	Key string `json:"key,omitempty"`
}

// Identity may contain the issuer and/or the subject found in the transparency
// log.
// Issuer/Subject uses a strict match, while IssuerRegExp and SubjectRegExp
// apply a regexp for matching.
type Identity struct {
	// Issuer defines the issuer for this identity.
	// +optional
	Issuer string `json:"issuer,omitempty"`
	// Subject defines the subject for this identity.
	// +optional
	Subject string `json:"subject,omitempty"`
	// IssuerRegExp specifies a regular expression to match the issuer for this identity.
	// +optional
	IssuerRegExp string `json:"issuerRegExp,omitempty"`
	// SubjectRegExp specifies a regular expression to match the subject for this identity.
	// +optional
	SubjectRegExp string `json:"subjectRegExp,omitempty"`
}

// RFC3161Timestamp specifies the URL to a RFC3161 time-stamping server that holds
// the time-stamped verification for the signature
type RFC3161Timestamp struct {
	// Use the Certificate Chain from the referred TrustRoot.TimeStampAuthorities
	// +optional
	TrustRootRef string `json:"trustRootRef,omitempty"`
}

// ClusterImagePolicyList is a list of ClusterImagePolicy resources
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ClusterImagePolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ClusterImagePolicy `json:"items"`
}
