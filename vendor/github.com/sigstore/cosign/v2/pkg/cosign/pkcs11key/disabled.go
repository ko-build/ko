//go:build !pkcs11key
// +build !pkcs11key

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

package pkcs11key

import (
	"context"
	"crypto"
	"crypto/x509"
	"errors"
	"io"

	"github.com/sigstore/sigstore/pkg/signature"
)

// The empty struct is used so this file never imports piv-go which is
// dependent on cgo and will fail to build if imported.
type empty struct{} //nolint

type Key struct{}

func GetKeyWithURIConfig(config *Pkcs11UriConfig, askForPinIfNeeded bool) (*Key, error) {
	return nil, errors.New("unimplemented")
}

func (k *Key) Certificate() (*x509.Certificate, error) {
	return nil, errors.New("unimplemented")
}

func (k *Key) PublicKey(opts ...signature.PublicKeyOption) (crypto.PublicKey, error) {
	return nil, errors.New("unimplemented")
}

func (k *Key) VerifySignature(signature, message io.Reader, opts ...signature.VerifyOption) error {
	return errors.New("unimplemented")
}

func (k *Key) Verifier() (signature.Verifier, error) {
	return nil, errors.New("unimplemented")
}

func (k *Key) Sign(ctx context.Context, rawPayload []byte) ([]byte, []byte, error) {
	return nil, nil, errors.New("unimplemented")
}

func (k *Key) SignMessage(message io.Reader, opts ...signature.SignOption) ([]byte, error) {
	return nil, errors.New("unimplemented")
}

func (k *Key) SignerVerifier() (signature.SignerVerifier, error) {
	return nil, errors.New("unimplemented")
}

func (k *Key) Close() {
}
