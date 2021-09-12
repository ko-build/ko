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

package cosign

import (
	"encoding/base64"
	"encoding/json"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/in-toto/in-toto-golang/in_toto"
	"github.com/in-toto/in-toto-golang/pkg/ssl"
	"github.com/pkg/errors"
	"github.com/sigstore/sigstore/pkg/signature/payload"
)

// SimpleClaimVerifier verifies that SignedPayload.Payload is a SimpleContainerImage payload which references the given image digest and contains the given annotations.
func SimpleClaimVerifier(sp SignedPayload, imageDigest v1.Hash, annotations map[string]interface{}) error {
	ss := &payload.SimpleContainerImage{}
	if err := json.Unmarshal(sp.Payload, ss); err != nil {
		return err
	}

	if err := sp.VerifyClaims(imageDigest, ss); err != nil {
		return err
	}

	if annotations != nil {
		if !correctAnnotations(annotations, ss.Optional) {
			return errors.New("missing or incorrect annotation")
		}
	}
	return nil
}

// IntotoSubjectClaimVerifier verifies that SignedPayload.Payload is an Intoto statement which references the given image digest.
func IntotoSubjectClaimVerifier(sp SignedPayload, imageDigest v1.Hash, _ map[string]interface{}) error {
	// The payload here is an envelope. We already verified the signature earlier.
	e := ssl.Envelope{}
	if err := json.Unmarshal(sp.Payload, &e); err != nil {
		return err
	}
	stBytes, err := base64.StdEncoding.DecodeString(e.Payload)
	if err != nil {
		return err
	}

	st := in_toto.Statement{}
	if err := json.Unmarshal(stBytes, &st); err != nil {
		return err
	}
	for _, subj := range st.StatementHeader.Subject {
		dgst, ok := subj.Digest["sha256"]
		if !ok {
			continue
		}
		subjDigest := "sha256:" + dgst
		if subjDigest == imageDigest.String() {
			return nil
		}
	}
	return errors.New("no matching subject digest found")
}
