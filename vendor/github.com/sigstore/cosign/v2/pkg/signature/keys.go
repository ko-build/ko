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

package signature

import (
	"context"
	"crypto"
	"crypto/x509"
	"errors"
	"fmt"
	"strings"

	"github.com/sigstore/cosign/v2/pkg/blob"
	"github.com/sigstore/cosign/v2/pkg/cosign"
	"github.com/sigstore/cosign/v2/pkg/cosign/git"
	"github.com/sigstore/cosign/v2/pkg/cosign/git/gitlab"
	"github.com/sigstore/cosign/v2/pkg/cosign/kubernetes"
	"github.com/sigstore/cosign/v2/pkg/cosign/pkcs11key"
	"github.com/sigstore/sigstore/pkg/cryptoutils"
	"github.com/sigstore/sigstore/pkg/signature"

	"github.com/sigstore/sigstore/pkg/signature/kms"
)

// LoadPublicKey is a wrapper for VerifierForKeyRef, hardcoding SHA256 as the hash algorithm
func LoadPublicKey(ctx context.Context, keyRef string) (verifier signature.Verifier, err error) {
	return VerifierForKeyRef(ctx, keyRef, crypto.SHA256)
}

// VerifierForKeyRef parses the given keyRef, loads the key and returns an appropriate
// verifier using the provided hash algorithm
func VerifierForKeyRef(ctx context.Context, keyRef string, hashAlgorithm crypto.Hash) (verifier signature.Verifier, err error) {
	// The key could be plaintext, in a file, at a URL, or in KMS.
	var perr *kms.ProviderNotFoundError
	kmsKey, err := kms.Get(ctx, keyRef, hashAlgorithm)
	switch {
	case err == nil:
		// KMS specified
		return kmsKey, nil
	case errors.As(err, &perr):
		// We can ignore ProviderNotFoundError; that just means the keyRef
		// didn't match any of the KMS schemes.
	default:
		// But other errors indicate something more insidious; pass those
		// through.
		return nil, err
	}

	raw, err := blob.LoadFileOrURL(keyRef)

	if err != nil {
		return nil, err
	}

	// PEM encoded file.
	pubKey, err := cryptoutils.UnmarshalPEMToPublicKey(raw)
	if err != nil {
		return nil, fmt.Errorf("pem to public key: %w", err)
	}

	return signature.LoadVerifier(pubKey, hashAlgorithm)
}

func loadKey(keyPath string, pf cosign.PassFunc) (signature.SignerVerifier, error) {
	kb, err := blob.LoadFileOrURL(keyPath)
	if err != nil {
		return nil, err
	}
	pass := []byte{}
	if pf != nil {
		pass, err = pf(false)
		if err != nil {
			return nil, err
		}
	}
	return cosign.LoadPrivateKey(kb, pass)
}

// LoadPublicKeyRaw loads a verifier from a PEM-encoded public key
func LoadPublicKeyRaw(raw []byte, hashAlgorithm crypto.Hash) (signature.Verifier, error) {
	pub, err := cryptoutils.UnmarshalPEMToPublicKey(raw)
	if err != nil {
		return nil, err
	}
	return signature.LoadVerifier(pub, hashAlgorithm)
}

func SignerFromKeyRef(ctx context.Context, keyRef string, pf cosign.PassFunc) (signature.Signer, error) {
	return SignerVerifierFromKeyRef(ctx, keyRef, pf)
}

func SignerVerifierFromKeyRef(ctx context.Context, keyRef string, pf cosign.PassFunc) (signature.SignerVerifier, error) {
	switch {
	case strings.HasPrefix(keyRef, pkcs11key.ReferenceScheme):
		pkcs11UriConfig := pkcs11key.NewPkcs11UriConfig()
		err := pkcs11UriConfig.Parse(keyRef)
		if err != nil {
			return nil, fmt.Errorf("parsing pkcs11 uri: %w", err)
		}

		// Since we'll be signing, we need to set askForPinIsNeeded to true
		// because we need access to the private key.
		sk, err := pkcs11key.GetKeyWithURIConfig(pkcs11UriConfig, true)
		if err != nil {
			return nil, fmt.Errorf("opening pkcs11 token key: %w", err)
		}

		sv, err := sk.SignerVerifier()
		if err != nil {
			return nil, fmt.Errorf("initializing pkcs11 token signer verifier: %w", err)
		}

		return sv, nil
	case strings.HasPrefix(keyRef, kubernetes.KeyReference):
		s, err := kubernetes.GetKeyPairSecret(ctx, keyRef)
		if err != nil {
			return nil, err
		}

		if len(s.Data) > 0 {
			return cosign.LoadPrivateKey(s.Data["cosign.key"], s.Data["cosign.password"])
		}
	case strings.HasPrefix(keyRef, gitlab.ReferenceScheme):
		split := strings.Split(keyRef, "://")

		if len(split) < 2 {
			return nil, errors.New("could not parse scheme, use <scheme>://<ref> format")
		}

		provider, targetRef := split[0], split[1]

		pk, err := git.GetProvider(provider).GetSecret(ctx, targetRef, "COSIGN_PRIVATE_KEY")
		if err != nil {
			return nil, err
		}

		pass, err := git.GetProvider(provider).GetSecret(ctx, targetRef, "COSIGN_PASSWORD")
		if err != nil {
			return nil, err
		}

		return cosign.LoadPrivateKey([]byte(pk), []byte(pass))
	}

	if strings.Contains(keyRef, "://") {
		sv, err := kms.Get(ctx, keyRef, crypto.SHA256)
		if err == nil {
			return sv, nil
		}
		var e *kms.ProviderNotFoundError
		if !errors.As(err, &e) {
			return nil, fmt.Errorf("kms get: %w", err)
		}
		// ProviderNotFoundError is okay; loadKey handles other URL schemes
	}

	return loadKey(keyRef, pf)
}

func PublicKeyFromKeyRef(ctx context.Context, keyRef string) (signature.Verifier, error) {
	return PublicKeyFromKeyRefWithHashAlgo(ctx, keyRef, crypto.SHA256)
}

func PublicKeyFromKeyRefWithHashAlgo(ctx context.Context, keyRef string, hashAlgorithm crypto.Hash) (signature.Verifier, error) {
	if strings.HasPrefix(keyRef, kubernetes.KeyReference) {
		s, err := kubernetes.GetKeyPairSecret(ctx, keyRef)
		if err != nil {
			return nil, err
		}

		if len(s.Data) > 0 {
			return LoadPublicKeyRaw(s.Data["cosign.pub"], hashAlgorithm)
		}
	}

	if strings.HasPrefix(keyRef, pkcs11key.ReferenceScheme) {
		pkcs11UriConfig := pkcs11key.NewPkcs11UriConfig()
		err := pkcs11UriConfig.Parse(keyRef)
		if err != nil {
			return nil, fmt.Errorf("parsing pkcs11 uri): %w", err)
		}

		// Since we'll be verifying a signature, we do not need to set askForPinIsNeeded to true
		// because we only need access to the public key.
		sk, err := pkcs11key.GetKeyWithURIConfig(pkcs11UriConfig, false)
		if err != nil {
			return nil, fmt.Errorf("opening pkcs11 token key: %w", err)
		}

		v, err := sk.Verifier()
		if err != nil {
			return nil, fmt.Errorf("initializing pkcs11 token verifier: %w", err)
		}

		return v, nil
	} else if strings.HasPrefix(keyRef, gitlab.ReferenceScheme) {
		split := strings.Split(keyRef, "://")

		if len(split) < 2 {
			return nil, errors.New("could not parse scheme, use <scheme>://<ref> format")
		}

		provider, targetRef := split[0], split[1]

		pubKey, err := git.GetProvider(provider).GetSecret(ctx, targetRef, "COSIGN_PUBLIC_KEY")
		if err != nil {
			return nil, err
		}

		if len(pubKey) > 0 {
			return LoadPublicKeyRaw([]byte(pubKey), hashAlgorithm)
		}
	}

	return VerifierForKeyRef(ctx, keyRef, hashAlgorithm)
}

func PublicKeyPem(key signature.PublicKeyProvider, pkOpts ...signature.PublicKeyOption) ([]byte, error) {
	pub, err := key.PublicKey(pkOpts...)
	if err != nil {
		return nil, err
	}
	return cryptoutils.MarshalPublicKeyToPEM(pub)
}

func CertSubject(c *x509.Certificate) string {
	switch {
	case c.EmailAddresses != nil:
		return c.EmailAddresses[0]
	case c.URIs != nil:
		return c.URIs[0].String()
	}
	// ignore error if there's no OtherName SAN
	otherName, _ := cryptoutils.UnmarshalOtherNameSAN(c.Extensions)
	if len(otherName) > 0 {
		return otherName
	}
	return ""
}
