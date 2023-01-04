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
	"bytes"
	"context"
	"crypto/x509"
	"encoding/json"

	"github.com/sigstore/policy-controller/pkg/tuf"
	"github.com/sigstore/sigstore/pkg/cryptoutils"

	"knative.dev/pkg/apis"
	"knative.dev/pkg/logging"
)

// By default the TUF repo contains this prefix, so if it's there, remove
// it.
const DefaultTUFRepoPrefix = "/repository/"

// Validate implements apis.Validatable
func (c *TrustRoot) Validate(ctx context.Context) *apis.FieldError {
	return c.Spec.Validate(ctx).ViaField("spec")
}

func (spec *TrustRootSpec) Validate(ctx context.Context) (errors *apis.FieldError) {
	if spec.Repository == nil && spec.Remote == nil && spec.SigstoreKeys == nil {
		return apis.ErrMissingOneOf("repository", "remote", "sigstoreKeys")
	}
	if spec.Repository != nil {
		if spec.Remote != nil || spec.SigstoreKeys != nil {
			return apis.ErrMultipleOneOf("repository", "remote", "sigstoreKeys")
		}
		return spec.Repository.Validate(ctx).ViaField("repository")
	}
	if spec.Remote != nil {
		if spec.Repository != nil || spec.SigstoreKeys != nil {
			return apis.ErrMultipleOneOf("repository", "remote", "sigstoreKeys")
		}
		return spec.Remote.Validate(ctx).ViaField("remote")
	}
	if spec.SigstoreKeys != nil {
		if spec.Remote != nil || spec.Repository != nil {
			return apis.ErrMultipleOneOf("repository", "remote", "sigstoreKeys")
		}
		return spec.SigstoreKeys.Validate(ctx).ViaField("sigstoreKeys")
	}
	return
}

func (repo *Repository) Validate(ctx context.Context) (errors *apis.FieldError) {
	if repo.Targets == "" {
		errors = errors.Also(apis.ErrMissingField("targets"))
	}

	errors = errors.Also(ValidateRoot(ctx, repo.Root))

	if len(repo.MirrorFS) == 0 {
		errors = errors.Also(apis.ErrMissingField("repository"))
	} else {
		if errors != nil {
			// We return here in case there in case there are errors. This is
			// because we do not want to pollute the error message, because
			// with any of the above errors, the TUF init will fail and it will
			// not be a meaningful error without fixing the above errors.
			return
		}
		// Make sure we can construct a TUF client out of it.
		c, err := tuf.ClientFromSerializedMirror(ctx, repo.MirrorFS, repo.Root, repo.Targets, DefaultTUFRepoPrefix)
		if err != nil {
			errors = errors.Also(apis.ErrInvalidValue("failed to construct a TUF client", "mirrorFS", err.Error()))
		} else {
			targetFiles, err := c.Targets()
			if err != nil {
				errors = errors.Also(apis.ErrInvalidValue("failed to get targets from a TUF client", "mirrorFS", err.Error()))
			}
			logging.FromContext(ctx).Debugf("FS uncompressed ok, have %d valid targets", len(targetFiles))
		}
	}
	return
}

func (remote *Remote) Validate(ctx context.Context) (errors *apis.FieldError) {
	if remote.Mirror.String() == "" {
		errors = errors.Also(apis.ErrMissingField("mirror"))
	}
	errors = errors.Also(ValidateRoot(ctx, remote.Root))
	return
}

func (sigstoreKeys *SigstoreKeys) Validate(ctx context.Context) (errors *apis.FieldError) {
	if len(sigstoreKeys.CertificateAuthorities) == 0 && len(sigstoreKeys.TimeStampAuthorities) == 0 {
		errors = errors.Also(apis.ErrMissingOneOf("certificateAuthority", "timestampAuthorities"))
	} else {
		for i, ca := range sigstoreKeys.CertificateAuthorities {
			errors = ValidateCertificateAuthority(ctx, ca).ViaFieldIndex("certificateAuthority", i)
		}
	}

	// These are optionals, so we just validate them if they are there and do
	// not report them as missing.
	for i, tsa := range sigstoreKeys.TimeStampAuthorities {
		errors = ValidateTimeStampAuthority(ctx, tsa).ViaFieldIndex("timestampAuthorities", i)
	}
	for i, ctl := range sigstoreKeys.CTLogs {
		errors = ValidateTransparencyLogInstance(ctx, ctl).ViaFieldIndex("ctLogs", i)
	}
	for i, tl := range sigstoreKeys.TLogs {
		errors = ValidateTransparencyLogInstance(ctx, tl).ViaFieldIndex("tLogs", i)
	}
	return
}

func ValidateRoot(ctx context.Context, rootJSON []byte) *apis.FieldError {
	if rootJSON == nil {
		return apis.ErrMissingField("root")
	}
	var root map[string]interface{}
	if err := json.Unmarshal(rootJSON, &root); err != nil {
		return apis.ErrInvalidValue("failed to unmarshal", "root", err.Error())
	}
	// TODO(vaikas): Tighten this validation to check for proper shape.
	if root["signatures"] == nil {
		return apis.ErrInvalidValue("missing signatures in root.json", "root", "no signatures")
	}
	return nil
}

func ValidateCertificateAuthority(ctx context.Context, ca CertificateAuthority) (errors *apis.FieldError) {
	errors = errors.Also(ValidateDistinguishedName(ctx, ca.Subject)).ViaField("subject")
	if ca.URI.String() == "" {
		errors = errors.Also(apis.ErrMissingField("uri"))
	}
	if len(ca.CertChain) == 0 {
		errors = errors.Also(apis.ErrMissingField("certchain"))
	}
	return
}

func ValidateTimeStampAuthority(ctx context.Context, ca CertificateAuthority) (errors *apis.FieldError) {
	errors = errors.Also(ValidateDistinguishedName(ctx, ca.Subject)).ViaField("subject")
	if ca.URI.String() == "" {
		errors = errors.Also(apis.ErrMissingField("uri"))
	}
	if len(ca.CertChain) == 0 {
		errors = errors.Also(apis.ErrMissingField("certchain"))
	}
	leaves, _, _, err := SplitPEMCertificateChain(ca.CertChain)
	if err != nil {
		errors = errors.Also(apis.ErrInvalidValue("error splitting the certificates", "certChain", err.Error()))
	}
	if len(leaves) > 1 {
		errors = errors.Also(apis.ErrInvalidValue("certificate chain must contain at most one TSA certificate", "certChain"))
	}
	return
}

func ValidateDistinguishedName(ctx context.Context, dn DistinguishedName) (errors *apis.FieldError) {
	if dn.Organization == "" {
		errors = errors.Also(apis.ErrMissingField("organization"))
	}
	if dn.CommonName == "" {
		errors = errors.Also(apis.ErrMissingField("commonName"))
	}
	return
}

func ValidateTransparencyLogInstance(ctx context.Context, tli TransparencyLogInstance) (errors *apis.FieldError) {
	if tli.BaseURL.String() == "" {
		errors = errors.Also(apis.ErrMissingField("baseURL"))
	}
	if tli.HashAlgorithm == "" {
		errors = errors.Also(apis.ErrMissingField("hashAlgorithm"))
	}
	if len(tli.PublicKey) == 0 {
		errors = errors.Also(apis.ErrMissingField("publicKey"))
	}
	return
}

// SplitPEMCertificateChain returns a list of leaf (non-CA) certificates, a certificate pool for
// intermediate CA certificates, and a certificate pool for root CA certificates
func SplitPEMCertificateChain(pem []byte) (leaves, intermediates, roots []*x509.Certificate, err error) {
	certs, err := cryptoutils.UnmarshalCertificatesFromPEM(pem)
	if err != nil {
		return nil, nil, nil, err
	}

	for _, cert := range certs {
		if !cert.IsCA {
			leaves = append(leaves, cert)
		} else {
			// root certificates are self-signed
			if bytes.Equal(cert.RawSubject, cert.RawIssuer) {
				roots = append(roots, cert)
			} else {
				intermediates = append(intermediates, cert)
			}
		}
	}

	return leaves, intermediates, roots, nil
}
