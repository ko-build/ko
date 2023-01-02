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
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sigstore/cosign/v2/pkg/oci"
	"github.com/sigstore/cosign/v2/pkg/oci/mutate"
	"github.com/sigstore/cosign/v2/pkg/oci/static"
	"github.com/sigstore/sigstore/pkg/signature"
)

// NewDupeDetector creates a new DupeDetector that looks for matching signatures that
// can verify the provided signature's payload.
func NewDupeDetector(v signature.Verifier) mutate.DupeDetector {
	return &dd{verifier: v}
}

func NewReplaceOp(predicateURI string) mutate.ReplaceOp {
	return &ro{predicateURI: predicateURI}
}

type dd struct {
	verifier signature.Verifier
}

type ro struct {
	predicateURI string
}

var _ mutate.DupeDetector = (*dd)(nil)
var _ mutate.ReplaceOp = (*ro)(nil)

func (dd *dd) Find(sigImage oci.Signatures, newSig oci.Signature) (oci.Signature, error) {
	newDigest, err := newSig.Digest()
	if err != nil {
		return nil, err
	}
	newMediaType, err := newSig.MediaType()
	if err != nil {
		return nil, err
	}
	newAnnotations, err := newSig.Annotations()
	if err != nil {
		return nil, err
	}

	sigs, err := sigImage.Get()
	if err != nil {
		return nil, err
	}

LayerLoop:
	for _, sig := range sigs {
		existingAnnotations, err := sig.Annotations()
		if err != nil {
			continue LayerLoop
		}

		// if there are any new annotations, then this isn't a duplicate
		for a, value := range newAnnotations {
			if a == static.SignatureAnnotationKey {
				continue // Ignore the signature key, we check it with custom logic below.
			}
			if val, ok := existingAnnotations[a]; !ok || val != value {
				continue LayerLoop
			}
		}
		if existingDigest, err := sig.Digest(); err != nil || existingDigest != newDigest {
			continue LayerLoop
		}
		if existingMediaType, err := sig.MediaType(); err != nil || existingMediaType != newMediaType {
			continue LayerLoop
		}

		existingSignature, err := sig.Base64Signature()
		if err != nil || existingSignature == "" {
			continue LayerLoop
		}
		uploadedSig, err := base64.StdEncoding.DecodeString(existingSignature)
		if err != nil {
			continue LayerLoop
		}
		r, err := newSig.Uncompressed()
		if err != nil {
			return nil, err
		}
		if err := dd.verifier.VerifySignature(bytes.NewReader(uploadedSig), r); err == nil {
			return sig, nil
		}
	}
	return nil, nil
}

func (r *ro) Replace(signatures oci.Signatures, o oci.Signature) (oci.Signatures, error) {
	sigs, err := signatures.Get()
	if err != nil {
		return nil, err
	}

	ros := &replaceOCISignatures{Signatures: signatures}

	sigsCopy := make([]oci.Signature, 0, len(sigs))
	sigsCopy = append(sigsCopy, o)

	if len(sigs) == 0 {
		ros.attestations = append(ros.attestations, sigsCopy...)
		return ros, nil
	}

	for _, s := range sigs {
		var signaturePayload map[string]interface{}
		p, err := s.Payload()
		if err != nil {
			return nil, fmt.Errorf("could not get payload: %w", err)
		}
		err = json.Unmarshal(p, &signaturePayload)
		if err != nil {
			return nil, fmt.Errorf("unmarshal payload data: %w", err)
		}

		val, ok := signaturePayload["payload"]
		if !ok {
			return nil, fmt.Errorf("could not find 'payload' in payload data")
		}
		decodedPayload, err := base64.StdEncoding.DecodeString(val.(string))
		if err != nil {
			return nil, fmt.Errorf("could not decode 'payload': %w", err)
		}

		var payloadData map[string]interface{}
		if err := json.Unmarshal(decodedPayload, &payloadData); err != nil {
			return nil, fmt.Errorf("unmarshal payloadData: %w", err)
		}
		val, ok = payloadData["predicateType"]
		if !ok {
			return nil, fmt.Errorf("could not find 'predicateType' in payload data")
		}
		if r.predicateURI == val {
			fmt.Fprintln(os.Stderr, "Replacing attestation predicate:", r.predicateURI)
			continue
		} else {
			fmt.Fprintln(os.Stderr, "Not replacing attestation predicate:", val)
			sigsCopy = append(sigsCopy, s)
		}
	}

	ros.attestations = append(ros.attestations, sigsCopy...)

	return ros, nil
}

type replaceOCISignatures struct {
	oci.Signatures
	attestations []oci.Signature
}

func (r *replaceOCISignatures) Get() ([]oci.Signature, error) {
	return r.attestations, nil
}
