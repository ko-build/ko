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
	"bytes"
	"context"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"runtime"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/pkg/errors"
	cremote "github.com/sigstore/cosign/pkg/cosign/remote"
	"github.com/sigstore/sigstore/pkg/cryptoutils"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

type SignedPayload struct {
	Base64Signature string
	Payload         []byte
	Cert            *x509.Certificate
	Chain           []*x509.Certificate
	Bundle          *cremote.Bundle
	bundleVerified  bool
}

// TODO: marshal the cert correctly.
// func (sp *SignedPayload) MarshalJSON() ([]byte, error) {
// 	x509.Certificate.
// 	pem.EncodeToMemory(&pem.Block{
// 		Type: "CERTIFICATE",
// 		Bytes:
// 	})
// }

const (
	SignatureTagSuffix   = ".sig"
	SBOMTagSuffix        = ".sbom"
	AttestationTagSuffix = ".att"
)

const (
	Signature   = "signature"
	SBOM        = "sbom"
	Attestation = "attestation"
)

func AttachedImageTag(repo name.Repository, digest v1.Hash, tagSuffix string) name.Tag {
	// sha256:d34db33f -> sha256-d34db33f.suffix
	tagStr := strings.ReplaceAll(digest.String(), ":", "-") + tagSuffix
	return repo.Tag(tagStr)
}

func FetchSignaturesForImage(ctx context.Context, signedImgRef name.Reference, sigRepo name.Repository, sigTagSuffix string, registryOpts ...remote.Option) ([]SignedPayload, error) {
	signedImgDesc, err := remote.Get(signedImgRef, registryOpts...)
	if err != nil {
		return nil, err
	}
	return FetchSignaturesForImageDigest(ctx, signedImgDesc.Descriptor.Digest, sigRepo, sigTagSuffix, registryOpts...)
}

func FetchSignaturesForImageDigest(ctx context.Context, signedImageDigest v1.Hash, sigRepo name.Repository, sigTagSuffix string, registryOpts ...remote.Option) ([]SignedPayload, error) {
	sigImgDesc, err := remote.Get(AttachedImageTag(sigRepo, signedImageDigest, sigTagSuffix), registryOpts...)
	if err != nil {
		return nil, errors.Wrap(err, "getting signature manifest")
	}
	sigImg, err := sigImgDesc.Image()
	if err != nil {
		return nil, errors.Wrap(err, "remote image")
	}

	m, err := sigImg.Manifest()
	if err != nil {
		return nil, errors.Wrap(err, "manifest")
	}

	g, ctx := errgroup.WithContext(ctx)
	signatures := make([]SignedPayload, len(m.Layers))
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))
	for i, desc := range m.Layers {
		i, desc := i, desc
		g.Go(func() error {
			if err := sem.Acquire(ctx, 1); err != nil {
				return err
			}
			defer sem.Release(1)
			base64sig, ok := desc.Annotations[sigkey]
			if !ok {
				return nil
			}
			l, err := sigImg.LayerByDigest(desc.Digest)
			if err != nil {
				return err
			}

			// Compressed is a misnomer here, we just want the raw bytes from the registry.
			r, err := l.Compressed()
			if err != nil {
				return err

			}
			payload, err := ioutil.ReadAll(r)
			if err != nil {
				return err
			}
			sp := SignedPayload{
				Payload:         payload,
				Base64Signature: base64sig,
			}
			// We may have a certificate and chain
			certPem := desc.Annotations[certkey]
			if certPem != "" {
				certs, err := cryptoutils.LoadCertificatesFromPEM(bytes.NewReader([]byte(certPem)))
				if err != nil {
					return err
				}
				sp.Cert = certs[0]
			}
			chainPem := desc.Annotations[chainkey]
			if chainPem != "" {
				certs, err := cryptoutils.LoadCertificatesFromPEM(bytes.NewReader([]byte(chainPem)))
				if err != nil {
					return err
				}
				sp.Chain = certs
			}

			bundle := desc.Annotations[BundleKey]
			if bundle != "" {
				var b cremote.Bundle
				if err := json.Unmarshal([]byte(bundle), &b); err != nil {
					return errors.Wrap(err, "unmarshaling bundle")
				}
				sp.Bundle = &b
			}

			signatures[i] = sp
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return signatures, nil

}
