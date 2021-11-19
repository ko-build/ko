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

package oci

import (
	"crypto/x509"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

// Signatures represents a set of signatures that are associated with a particular
// v1.Image.
type Signatures interface {
	v1.Image // The low-level representation of the signatures

	// Get retrieves the list of signatures stored.
	Get() ([]Signature, error)
}

// Signature holds a single image signature.
type Signature interface {
	v1.Layer

	// Annotations returns the annotations associated with this layer.
	Annotations() (map[string]string, error)

	// Payload fetches the opaque data that is being signed.
	// This will always return data when there is no error.
	Payload() ([]byte, error)

	// Base64Signature fetches the base64 encoded signature
	// of the payload.  This will always return data when
	// there is no error.
	Base64Signature() (string, error)

	// Cert fetches the optional public key from the key pair that
	// was used to sign the payload.
	Cert() (*x509.Certificate, error)

	// Chain fetches the optional "full certificate chain" rooted
	// at a Fulcio CA, the leaf of which was used to sign the
	// payload.
	Chain() ([]*x509.Certificate, error)

	// Bundle fetches the optional metadata that records the ephemeral
	// Fulcio key in the transparency log.
	Bundle() (*Bundle, error)
}

// Bundle holds metadata about recording a Signature's ephemeral key to
// a Rekor transparency log.
type Bundle struct {
	SignedEntryTimestamp []byte
	Payload              BundlePayload
}

type BundlePayload struct {
	Body           interface{} `json:"body"`
	IntegratedTime int64       `json:"integratedTime"`
	LogIndex       int64       `json:"logIndex"`
	LogID          string      `json:"logID"`
}
