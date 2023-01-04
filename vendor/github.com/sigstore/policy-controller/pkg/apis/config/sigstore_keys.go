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
	"encoding/json"
	"fmt"

	"github.com/sigstore/policy-controller/pkg/apis/policy/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/yaml"
)

const (
	// SigstoreKeysConfigName is the name of ConfigMap created by the
	// reconciler and consumed by the admission webhook for determining
	// which Keys/Certificates are trusted for things like Fulcio/Rekor, etc.
	SigstoreKeysConfigName = "config-sigstore-keys"
)

// Note that these are 1:1 mapped to public API SigstoreKeys. Reasoning
// being that we may choose to serialize these differently, or use the protos
// that are defined upstream, so want to keep the public/private distinction, so
// that we can change things independend of breaking the API. Time will tell
// if this is the right call, but we can always reunify them later if we so
// want.
// TODO(vaikas): See about replacing these with the protos here once they land
// and see how easy it is to replace with protos instead of our custom defs
// above.
// https://github.com/sigstore/protobuf-specs/pull/5
// And in particular: https://github.com/sigstore/protobuf-specs/pull/5/files#diff-b1f89b7fd3eb27b519380b092a2416f893a96fbba3f8c90cfa767e7687383ad4R70

// TransparencyLogInstance describes the immutable parameters from a
// transparency log.
// See https://www.rfc-editor.org/rfc/rfc9162.html#name-log-parameters
// for more details.
// The incluced parameters are the minimal set required to identify a log,
// and verify an inclusion promise.
type TransparencyLogInstance struct {
	BaseURL       apis.URL `json:"baseURL"`
	HashAlgorithm string   `json:"hashAlgorithm"`
	// PEM encoded public key
	PublicKey []byte `json:"publicKey"`
	LogID     string `json:"logID"`
}

type DistinguishedName struct {
	Organization string `json:"organization"`
	CommonName   string `json:"commonName"`
}

type CertificateAuthority struct {
	// The root certificate MUST be self-signed, and so the subject and
	// issuer are the same.
	Subject DistinguishedName `json:"subject"`
	// The URI at which the CA can be accessed.
	URI apis.URL `json:"uri"`
	// The certificate chain for this CA.
	// CertChain is in PEM format.
	CertChain []byte `json:"certChain"`

	// TODO(vaikas): How to best represent this
	// The time the *entire* chain was valid. This is at max the
	// longest interval when *all* certificates in the chain where valid,
	// but it MAY be shorter.
	//       dev.sigstore.common.v1.TimeRange valid_for = 4;
}

// SigstoreKeys contains all the necessary Keys and Certificates for validating
// against a specific instance of Sigstore.
// TODO(vaikas): See about replacing these with the protos here once they land
// and see how easy it is to replace with protos instead of our custom defs
// above.
// https://github.com/sigstore/protobuf-specs/pull/5
// And in particular: https://github.com/sigstore/protobuf-specs/pull/5/files#diff-b1f89b7fd3eb27b519380b092a2416f893a96fbba3f8c90cfa767e7687383ad4R70
// Well, not the multi-root, but one instance of that is exactly the
// SigstoreKeys.
type SigstoreKeys struct {
	// Trusted certificate authorities (e.g Fulcio).
	CertificateAuthorities []CertificateAuthority `json:"certificateAuthorities,omitempty"`
	// Rekor log specifications
	TLogs []TransparencyLogInstance `json:"tLogs,omitempty"`
	// Certificate Transparency Log
	CTLogs []TransparencyLogInstance `json:"ctLogs,omitempty"`
	// Trusted timestamping authorities
	TimeStampAuthorities []CertificateAuthority `json:"timestampAuthorities"`
}

type SigstoreKeysMap struct {
	SigstoreKeys map[string]SigstoreKeys
}

// NewSigstoreKeysFromMap creates a map of SigstoreKeys to use for validation.
func NewSigstoreKeysFromMap(data map[string]string) (*SigstoreKeysMap, error) {
	ret := make(map[string]SigstoreKeys, len(data))
	// Spin through the ConfigMap. Each entry will have a serialized form of
	// necessary validation keys in the form of SigstoreKeys.
	for k, v := range data {
		// This is the example that we use to document / test the ConfigMap.
		if k == "_example" {
			continue
		}
		if v == "" {
			return nil, fmt.Errorf("configmap has an entry %q but no value", k)
		}
		sigstoreKeys := &SigstoreKeys{}

		if err := parseSigstoreKeys(v, sigstoreKeys); err != nil {
			return nil, fmt.Errorf("failed to parse the entry %q : %q : %w", k, v, err)
		}
		ret[k] = *sigstoreKeys
	}
	return &SigstoreKeysMap{SigstoreKeys: ret}, nil
}

// NewImagePoliciesConfigFromConfigMap creates a Features from the supplied ConfigMap
func NewSigstoreKeysFromConfigMap(config *corev1.ConfigMap) (*SigstoreKeysMap, error) {
	return NewSigstoreKeysFromMap(config.Data)
}

func parseSigstoreKeys(entry string, out interface{}) error {
	j, err := yaml.YAMLToJSON([]byte(entry))
	if err != nil {
		return fmt.Errorf("config's value could not be converted to JSON: %w : %s", err, entry)
	}
	return json.Unmarshal(j, &out)
}

// ConvertFrom takes a source and converts into a SigstoreKeys suitable
// for serialization into a ConfigMap entry.
func (sk *SigstoreKeys) ConvertFrom(ctx context.Context, source *v1alpha1.SigstoreKeys) {
	sk.CertificateAuthorities = make([]CertificateAuthority, len(source.CertificateAuthorities))
	for i := range source.CertificateAuthorities {
		sk.CertificateAuthorities[i] = ConvertCertificateAuthority(source.CertificateAuthorities[i])
	}

	sk.TLogs = make([]TransparencyLogInstance, len(source.TLogs))
	for i := range source.TLogs {
		sk.TLogs[i] = ConvertTransparencyLogInstance(source.TLogs[i])
	}

	sk.CTLogs = make([]TransparencyLogInstance, len(source.CTLogs))
	for i := range source.CTLogs {
		sk.CTLogs[i] = ConvertTransparencyLogInstance(source.CTLogs[i])
	}

	sk.TimeStampAuthorities = make([]CertificateAuthority, len(source.TimeStampAuthorities))
	for i := range source.TimeStampAuthorities {
		sk.TimeStampAuthorities[i] = ConvertCertificateAuthority(source.TimeStampAuthorities[i])
	}
}

// ConvertCertificateAuthority converts public into private CertificateAuthority
func ConvertCertificateAuthority(source v1alpha1.CertificateAuthority) CertificateAuthority {
	return CertificateAuthority{
		Subject: DistinguishedName{
			Organization: source.Subject.Organization,
			CommonName:   source.Subject.CommonName,
		},
		URI:       *source.URI.DeepCopy(),
		CertChain: source.CertChain,
	}
}

// ConvertTransparencyLogInstance converts public into private
// TransparencyLogInstance.
func ConvertTransparencyLogInstance(source v1alpha1.TransparencyLogInstance) TransparencyLogInstance {
	return TransparencyLogInstance{
		BaseURL:       *source.BaseURL.DeepCopy(),
		HashAlgorithm: source.HashAlgorithm,
		PublicKey:     source.PublicKey,
	}
}
