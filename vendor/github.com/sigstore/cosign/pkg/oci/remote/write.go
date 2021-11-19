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

package remote

import (
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/sigstore/cosign/pkg/oci"
)

// WriteSignature publishes the signatures attached to the given entity
// into the provided repository.
func WriteSignatures(repo name.Repository, se oci.SignedEntity, opts ...Option) error {
	o := makeOptions(repo, opts...)

	// Access the signature list to publish
	sigs, err := se.Signatures()
	if err != nil {
		return err
	}

	// Determine the tag to which these signatures should be published.
	h, err := se.(digestable).Digest()
	if err != nil {
		return err
	}
	tag := o.TargetRepository.Tag(normalize(h, o.TagPrefix, o.SignatureSuffix))

	// Write the Signatures image to the tag, with the provided remote.Options
	return remoteWrite(tag, sigs, o.ROpt...)
}

// WriteAttestations publishes the attestations attached to the given entity
// into the provided repository.
func WriteAttestations(repo name.Repository, se oci.SignedEntity, opts ...Option) error {
	o := makeOptions(repo, opts...)

	// Access the signature list to publish
	atts, err := se.Attestations()
	if err != nil {
		return err
	}

	// Determine the tag to which these signatures should be published.
	h, err := se.(digestable).Digest()
	if err != nil {
		return err
	}
	tag := o.TargetRepository.Tag(normalize(h, o.TagPrefix, o.AttestationSuffix))

	// Write the Signatures image to the tag, with the provided remote.Options
	return remoteWrite(tag, atts, o.ROpt...)
}
