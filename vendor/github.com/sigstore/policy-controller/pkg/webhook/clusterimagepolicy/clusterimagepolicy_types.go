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

package clusterimagepolicy

import (
	"context"
	"crypto"
	"encoding/json"
	"fmt"

	"github.com/google/go-containerregistry/pkg/authn/k8schain"
	"github.com/google/go-containerregistry/pkg/authn/kubernetes"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	ociremote "github.com/sigstore/cosign/v2/pkg/oci/remote"
	"github.com/sigstore/policy-controller/pkg/apis/policy/v1alpha1"
	signaturealgo "github.com/sigstore/policy-controller/pkg/apis/signaturealgo"
	"github.com/sigstore/sigstore/pkg/cryptoutils"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/apis"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/ptr"
)

// ClusterImagePolicy defines the images that go through verification
// and the authorities used for verification.
// This is the internal representation of the external v1alpha1.ClusterImagePolicy.
// KeyRef does not store secretRefs in internal representation.
// KeyRef does store parsed publicKeys from Data in internal representation.
type ClusterImagePolicy struct {
	// UID of the CIP so we can tell if they've been deleted/recreated
	UID types.UID `json:"uid,inline"`
	// ResourceVersion can be used to know if the CIP has been modified
	ResourceVersion string `json:"resourceVersion"`

	Images      []v1alpha1.ImagePattern `json:"images"`
	Authorities []Authority             `json:"authorities"`
	// Policy is an optional policy used to evaluate the results of valid
	// Authorities. Will not get evaluated unless at least one Authority
	// succeeds.
	Policy *AttestationPolicy `json:"policy,omitempty"`
	// Mode controls whether a failing policy will be rejected (not admitted),
	// or if errors are converted to Warnings.
	// enforce - Reject (default)
	// warn - allow but warn
	// +optional
	Mode string `json:"mode,omitempty"`
	// Match allows selecting resources based on their properties.
	Match []v1alpha1.MatchResource `json:"match,omitempty"`
}

type Authority struct {
	// Name is the name for this authority. Used by the CIP Policy
	// validator to be able to reference matching signature or attestation
	// verifications.
	Name string `json:"name"`
	// +optional
	Key *KeyRef `json:"key,omitempty"`
	// +optional
	Keyless *KeylessRef `json:"keyless,omitempty"`
	// +optional
	Static *StaticRef `json:"static,omitempty"`
	// +optional
	Sources []v1alpha1.Source `json:"source,omitempty"`
	// +optional
	CTLog *v1alpha1.TLog `json:"ctlog,omitempty"`
	// RemoteOpts are not marshalled because they are an unsupported type
	// RemoteOpts will be populated by the Authority UnmarshalJSON override
	// +optional
	RemoteOpts []ociremote.Option `json:"-"`
	// +optional
	Attestations []AttestationPolicy `json:"attestations,omitempty"`
	// +optional
	RFC3161Timestamp *RFC3161Timestamp `json:"rfc3161timestamp,omitempty"`
}

// This references a public verification key stored in
// a secret in the cosign-system namespace.
type KeyRef struct {
	// Data contains the inline public key
	// +optional
	Data string `json:"data,omitempty"`
	// HashAlgorithm always defaults to sha256 if the algorithm hasn't been explicitly set
	// +optional
	HashAlgorithm string `json:"hashAlgorithm,omitempty"`
	// HashAlgorithmCode sets the crypto.Hash code based on the value of HashAlgorithm.
	// HashAlgorithmCode is not marshalled, but we use the calculated crypto.Hash in the validations
	// +optional
	HashAlgorithmCode crypto.Hash `json:"-"`
	// PublicKeys are not marshalled because JSON unmarshalling
	// errors for *big.Int
	// +optional
	PublicKeys []crypto.PublicKey `json:"-"`
}

type KeylessRef struct {
	// +optional
	URL *apis.URL `json:"url,omitempty"`
	// +optional
	Identities []v1alpha1.Identity `json:"identities,omitempty"`
	// +optional
	CACert *KeyRef `json:"ca-cert,omitempty"`
	// Use the Certificate Chain from the referred TrustRoot.CertificateAuthorities and TrustRoot.CTLog
	// +optional
	TrustRootRef string `json:"trustRootRef,omitempty"`
}

type StaticRef struct {
	Action string `json:"action"`
}

type AttestationPolicy struct {
	// Name of the Attestation
	Name string `json:"name"`
	// PredicateType to attest, one of the accepted in verify-attestation
	PredicateType string `json:"predicateType"`
	// Type specifies how to evaluate policy, only rego/cue are understood.
	Type string `json:"type,omitempty"`
	// Data is the inlined version of the Policy used to evaluate the
	// Attestation.
	Data string `json:"data,omitempty"`
	// FetchConfigFile controls whether ConfigFile will be fetched and made
	// available for CIP level policy evaluation. Note that this only gets
	// evaluated (and hence fetched) iff at least one authority matches.
	// The ConfigFile will then be available in this format:
	// https://github.com/opencontainers/image-spec/blob/main/config.md
	FetchConfigFile *bool `json:"fetchConfigFile,omitempty"`
	// IncludeSpec controls whether resource `Spec` will be included and
	// made available for CIP level policy evaluation. Note that this only gets
	// evaluated iff at least one authority matches.
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

// RFC3161Timestamp specifies the URL to a RFC3161 time-stamping server that holds
// the time-stamped verification for the signature
type RFC3161Timestamp struct {
	// Use the Certificate Chain from the referred TrustRoot.TimeStampAuthorities
	// +optional
	TrustRootRef string `json:"trustRootRef,omitempty"`
}

// UnmarshalJSON populates the PublicKeys using Data because
// JSON unmashalling errors for *big.Int
func (k *KeyRef) UnmarshalJSON(data []byte) error {
	var publicKeys []crypto.PublicKey
	var err error

	ret := make(map[string]string)
	if err = json.Unmarshal(data, &ret); err != nil {
		return err
	}

	k.Data = ret["data"]
	k.HashAlgorithmCode = crypto.SHA256
	k.HashAlgorithm = signaturealgo.DefaultSignatureAlgorithm
	if ret["hashAlgorithm"] != "" {
		k.HashAlgorithm = ret["hashAlgorithm"]
		k.HashAlgorithmCode, err = signaturealgo.HashAlgorithm(ret["hashAlgorithm"])
		if err != nil {
			return err
		}
	}

	if ret["data"] != "" {
		publicKey, err := cryptoutils.UnmarshalPEMToPublicKey([]byte(ret["data"]))
		if err != nil {
			return fmt.Errorf("failed to unmarshal PEM public key %w", err)
		}
		publicKeys = append(publicKeys, publicKey)
	}
	k.PublicKeys = publicKeys

	return nil
}

// UnmarshalJSON populates the authority with the remoteOpts
// from authority sources
func (a *Authority) UnmarshalJSON(data []byte) error {
	// Create a new type to avoid recursion
	type RawAuthority Authority

	var rawAuthority RawAuthority
	err := json.Unmarshal(data, &rawAuthority)
	if err != nil {
		return err
	}

	// Determine additional RemoteOpts
	if len(rawAuthority.Sources) > 0 {
		for _, source := range rawAuthority.Sources {
			if source.OCI != "" {
				if targetRepoOverride, err := name.NewRepository(source.OCI); err != nil {
					return fmt.Errorf("failed to determine source: %w", err)
				} else if (targetRepoOverride != name.Repository{}) {
					rawAuthority.RemoteOpts = append(rawAuthority.RemoteOpts, ociremote.WithTargetRepository(targetRepoOverride))
				}
			}
		}
	}

	// Set the new type instance to casted original
	*a = Authority(rawAuthority)
	return nil
}

// SourceSignaturePullSecretsOpts creates the signaturePullSecrets remoteOpts
// This is not stored in the Authority under RemoteOpts as the namespace can be different
func (a *Authority) SourceSignaturePullSecretsOpts(ctx context.Context, namespace string) ([]ociremote.Option, error) {
	var ret []ociremote.Option
	for _, source := range a.Sources {
		if len(source.SignaturePullSecrets) > 0 {
			signaturePullSecrets := make([]string, 0, len(source.SignaturePullSecrets))
			for _, s := range source.SignaturePullSecrets {
				signaturePullSecrets = append(signaturePullSecrets, s.Name)
			}

			// Use NoServiceAccount when setting a signaturePullSecrets to avoid unnecessary API calls.
			opt := k8schain.Options{
				Namespace:          namespace,
				ServiceAccountName: kubernetes.NoServiceAccount,
				ImagePullSecrets:   signaturePullSecrets,
			}

			kc, err := k8schain.New(ctx, kubeclient.Get(ctx), opt)
			if err != nil {
				logging.FromContext(ctx).Errorf("failed creating keychain: %+v", err)
				return nil, err
			}

			ret = append(ret, ociremote.WithRemoteOptions(
				remote.WithContext(ctx),
				remote.WithAuthFromKeychain(kc),
			))
		}
	}

	return ret, nil
}

func ConvertClusterImagePolicyV1alpha1ToWebhook(in *v1alpha1.ClusterImagePolicy) *ClusterImagePolicy {
	copyIn := in.DeepCopy()

	outAuthorities := make([]Authority, 0)
	for _, authority := range copyIn.Spec.Authorities {
		outAuthority := convertAuthorityV1Alpha1ToWebhook(authority)
		outAuthorities = append(outAuthorities, *outAuthority)
	}

	// If there's a ClusterImagePolicy level AttestationPolicy, convert it here.
	var cipAttestationPolicy *AttestationPolicy
	if in.Spec.Policy != nil {
		cipAttestationPolicy = &AttestationPolicy{
			Type: in.Spec.Policy.Type,
			Data: in.Spec.Policy.Data,
		}
		if in.Spec.Policy.FetchConfigFile != nil {
			cipAttestationPolicy.FetchConfigFile = ptr.Bool(*in.Spec.Policy.FetchConfigFile)
		}
		if in.Spec.Policy.IncludeSpec != nil {
			cipAttestationPolicy.IncludeSpec = ptr.Bool(*in.Spec.Policy.IncludeSpec)
		}
		if in.Spec.Policy.IncludeObjectMeta != nil {
			cipAttestationPolicy.IncludeObjectMeta = ptr.Bool(*in.Spec.Policy.IncludeObjectMeta)
		}
		if in.Spec.Policy.IncludeTypeMeta != nil {
			cipAttestationPolicy.IncludeTypeMeta = ptr.Bool(*in.Spec.Policy.IncludeTypeMeta)
		}
	}
	return &ClusterImagePolicy{
		UID:             copyIn.UID,
		ResourceVersion: copyIn.ResourceVersion,
		Images:          copyIn.Spec.Images,
		Authorities:     outAuthorities,
		Policy:          cipAttestationPolicy,
		Mode:            in.Spec.Mode,
		Match:           in.Spec.Match,
	}
}

func convertAuthorityV1Alpha1ToWebhook(in v1alpha1.Authority) *Authority {
	keyRef := convertKeyRefV1Alpha1ToWebhook(in.Key)
	keylessRef := convertKeylessRefV1Alpha1ToWebhook(in.Keyless)
	staticRef := convertStaticRefV1Alpha1ToWebhook(in.Static)
	attestations := convertAttestationsV1Alpha1ToWebhook(in.Attestations)
	rfc3161Timestamp := convertRFC3161TimestampV1Alpha1ToWebhook(in.RFC3161Timestamp)

	return &Authority{
		Name:             in.Name,
		Key:              keyRef,
		Keyless:          keylessRef,
		Static:           staticRef,
		Sources:          in.Sources,
		CTLog:            in.CTLog,
		RFC3161Timestamp: rfc3161Timestamp,
		Attestations:     attestations,
	}
}

func convertRFC3161TimestampV1Alpha1ToWebhook(in *v1alpha1.RFC3161Timestamp) *RFC3161Timestamp {
	if in == nil {
		return nil
	}

	return &RFC3161Timestamp{
		TrustRootRef: in.TrustRootRef,
	}
}

func convertAttestationsV1Alpha1ToWebhook(in []v1alpha1.Attestation) []AttestationPolicy {
	ret := []AttestationPolicy{}
	for _, inAtt := range in {
		outAtt := AttestationPolicy{
			Name:          inAtt.Name,
			PredicateType: inAtt.PredicateType,
		}
		if inAtt.Policy != nil {
			outAtt.Type = inAtt.Policy.Type
			outAtt.Data = inAtt.Policy.Data
			if inAtt.Policy.FetchConfigFile != nil {
				outAtt.FetchConfigFile = ptr.Bool(*inAtt.Policy.FetchConfigFile)
			}
			if inAtt.Policy.IncludeSpec != nil {
				outAtt.IncludeSpec = ptr.Bool(*inAtt.Policy.IncludeSpec)
			}
			if inAtt.Policy.IncludeObjectMeta != nil {
				outAtt.IncludeObjectMeta = ptr.Bool(*inAtt.Policy.IncludeObjectMeta)
			}
			if inAtt.Policy.IncludeTypeMeta != nil {
				outAtt.IncludeTypeMeta = ptr.Bool(*inAtt.Policy.IncludeTypeMeta)
			}
		}
		ret = append(ret, outAtt)
	}
	return ret
}

func convertKeyRefV1Alpha1ToWebhook(in *v1alpha1.KeyRef) *KeyRef {
	if in == nil {
		return nil
	}
	// Convert the hash algorithm name to the code and reuse it everywhere else
	algorithmCode := crypto.SHA256
	algorithm := signaturealgo.DefaultSignatureAlgorithm
	if in.HashAlgorithm != "" {
		algorithm = in.HashAlgorithm
		// Ignore the error. It was already validated by the validation webhook
		algorithmCode, _ = signaturealgo.HashAlgorithm(in.HashAlgorithm) // nolint: staticcheck
	}

	return &KeyRef{
		Data:              in.Data,
		HashAlgorithm:     algorithm,
		HashAlgorithmCode: algorithmCode,
	}
}

func convertKeylessRefV1Alpha1ToWebhook(in *v1alpha1.KeylessRef) *KeylessRef {
	if in == nil {
		return nil
	}

	CACertRef := convertKeyRefV1Alpha1ToWebhook(in.CACert)

	return &KeylessRef{
		URL:          in.URL,
		Identities:   in.Identities,
		CACert:       CACertRef,
		TrustRootRef: in.TrustRootRef,
	}
}

func convertStaticRefV1Alpha1ToWebhook(in *v1alpha1.StaticRef) *StaticRef {
	if in == nil {
		return nil
	}

	return &StaticRef{
		Action: in.Action,
	}
}
