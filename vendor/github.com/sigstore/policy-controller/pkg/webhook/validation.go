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

package webhook

import (
	"context"
	"crypto"
	"encoding/pem"

	"github.com/google/go-containerregistry/pkg/name"
	"knative.dev/pkg/logging"

	"github.com/sigstore/cosign/v2/pkg/cosign"
	"github.com/sigstore/cosign/v2/pkg/oci"
	"github.com/sigstore/sigstore/pkg/signature"
)

func valid(ctx context.Context, ref name.Reference, keys []crypto.PublicKey, hashAlgo crypto.Hash, checkOpts *cosign.CheckOpts) ([]oci.Signature, error) {
	if len(keys) == 0 {
		return validSignatures(ctx, ref, checkOpts)
	}
	// We return nil if ANY key matches
	var lastErr error
	for _, k := range keys {
		verifier, err := signature.LoadVerifier(k, hashAlgo)
		if err != nil {
			logging.FromContext(ctx).Errorf("error creating verifier: %v", err)
			lastErr = err
			continue
		}
		checkOpts.SigVerifier = verifier
		sps, err := validSignatures(ctx, ref, checkOpts)
		if err != nil {
			logging.FromContext(ctx).Errorf("error validating signatures: %v", err)
			lastErr = err
			continue
		}
		return sps, nil
	}
	logging.FromContext(ctx).Debug("No valid signatures were found.")
	return nil, lastErr
}

// For testing
var cosignVerifySignatures = cosign.VerifyImageSignatures
var cosignVerifyAttestations = cosign.VerifyImageAttestations

func validSignatures(ctx context.Context, ref name.Reference, checkOpts *cosign.CheckOpts) ([]oci.Signature, error) {
	checkOpts.ClaimVerifier = cosign.SimpleClaimVerifier
	sigs, _, err := cosignVerifySignatures(ctx, ref, checkOpts)
	return sigs, err
}

func validAttestations(ctx context.Context, ref name.Reference, checkOpts *cosign.CheckOpts) ([]oci.Signature, error) {
	checkOpts.ClaimVerifier = cosign.IntotoSubjectClaimVerifier
	attestations, _, err := cosignVerifyAttestations(ctx, ref, checkOpts)
	return attestations, err
}

func parsePems(b []byte) []*pem.Block {
	p, rest := pem.Decode(b)
	if p == nil {
		return nil
	}
	pems := []*pem.Block{p}

	if rest != nil {
		return append(pems, parsePems(rest)...)
	}
	return pems
}
