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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/kmeta"
)

// TrustRoot defines the keys and certificates that are trusted for
// validating against. These can be specified as TUF Roots, serialized TUF
// repository (for air-gap scenarios), as well as serialized keys/certificates,
// for bring your own keys/certs.
//
// +genclient
// +genclient:nonNamespaced
// +genclient:noStatus
// +genreconciler:krshapedlogic=false

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type TrustRoot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	// Spec is the definition for a trust root. This is either a TUF root and
	// remote or local repository. You can also bring your own keys/certs here.
	Spec TrustRootSpec `json:"spec"`
}

var (
	_ apis.Validatable   = (*TrustRoot)(nil)
	_ apis.Defaultable   = (*TrustRoot)(nil)
	_ kmeta.OwnerRefable = (*TrustRoot)(nil)
)

// GetGroupVersionKind implements kmeta.OwnerRefable
func (c *TrustRoot) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("TrustRoot")
}

// TrustRootSpec defines a trusted Root. This is typically either a TUF Root
// or a bring your own keys variation.
// It specifies either:
// root.json and remote
// or
// fully gzipped / tarred directory containing root and metadata directories
// or
// serialized keys / certificate chains (bring your own keys).
type TrustRootSpec struct {
	// Remote specifies initial root of trust & remote mirror.
	// +optional
	Remote *Remote `json:"remote,omitempty"`

	// Repository contains the serialized TUF remote repository.
	// +optional
	Repository *Repository `json:"repository,omitempty"`

	// SigstoreKeys contains the serialized keys.
	// +optional
	SigstoreKeys *SigstoreKeys `json:"sigstoreKeys,omitempty"`
}

// Remote specifies the TUF with trusted initial root and remote mirror where
// to fetch updates from.
type Remote struct {
	// Root is the base64 encoded, json trusted initial root.
	Root []byte `json:"root"`

	// Mirror is the remote mirror, for example:
	// https://sigstore-tuf-root.storage.googleapis.com
	Mirror apis.URL `json:"mirror"`

	// Targets is where the targets live off of the root of the Remote
	// If not specified 'targets' is defaulted.
	// +optional
	Targets string `json:"targets,omitempty"`
}

// Repository specifies an airgapped TUF. Specifies the trusted initial root as
// well as a serialized repository.
type Repository struct {
	// Root is the base64 encoded, json trusted initial root.
	Root []byte `json:"root"`

	// MirrorFS is the base64 tarred, gzipped, and base64 encoded remote
	// repository that can be used for example in air-gap environments. Will
	// not make outbound network connections, and must then be kept up to date
	// in some other manner.
	// The repository must contain metadata as well as targets.
	MirrorFS []byte `json:"mirrorFS"`

	// Targets is where the targets live off of the root of the Repository
	// above. If not specified 'targets' is defaulted.
	// +optional
	Targets string `json:"targets,omitempty"`
}

// TransparencyLogInstance describes the immutable parameters from a
// transparency log.
// See https://www.rfc-editor.org/rfc/rfc9162.html#name-log-parameters
// for more details.
// The incluced parameters are the minimal set required to identify a log,
// and verify an inclusion promise.
type TransparencyLogInstance struct {
	// The base URL which can be used for URLs for clients.
	BaseURL apis.URL `json:"baseURL"`
	// / The hash algorithm used for the Merkle Tree
	HashAlgorithm string `json:"hashAlgorithm"`
	// PEM encoded public key
	PublicKey []byte `json:"publicKey"`
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
	// The certificate chain for this CA in PEM format. Last entry in this
	// chain is the Root certificate.
	CertChain []byte `json:"certChain"`

	// TODO(vaikas): How to best represent this
	// The time the *entire* chain was valid. This is at max the
	// longest interval when *all* certificates in the chain where valid,
	// but it MAY be shorter.
	//       dev.sigstore.common.v1.TimeRange valid_for = 4;
}

// SigstoreKeys contains all the necessary Keys and Certificates for validating
// against a specific instance of Sigstore. This is used for bringing your own
// trusted keys/certs.
// TODO(vaikas): See about replacing these with the protos here once they land
// and see how easy it is to replace with protos instead of our custom defs
// above.
// https://github.com/sigstore/protobuf-specs/pull/5
// And in particular: https://github.com/sigstore/protobuf-specs/pull/5/files#diff-b1f89b7fd3eb27b519380b092a2416f893a96fbba3f8c90cfa767e7687383ad4R70
// Well, not the multi-root, but one instance of that is exactly the
// SigstoreKeys.
type SigstoreKeys struct {
	// Trusted certificate authorities (e.g Fulcio).
	CertificateAuthorities []CertificateAuthority `json:"certificateAuthorities"`
	// Rekor log specifications
	// +optional
	TLogs []TransparencyLogInstance `json:"tLogs,omitempty"`
	// Certificate Transparency Log
	// +optional
	CTLogs []TransparencyLogInstance `json:"ctLogs,omitempty"`
	// Trusted timestamping authorities
	// +optional
	TimeStampAuthorities []CertificateAuthority `json:"timestampAuthorities,omitempty"`
}

// TrustRootList is a list of TrustRoot resources
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type TrustRootList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []TrustRoot `json:"items"`
}
