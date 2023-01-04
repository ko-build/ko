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
	"crypto"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/digitorus/timestamp"
	cbundle "github.com/sigstore/cosign/v2/pkg/cosign/bundle"
	"github.com/sigstore/sigstore/pkg/tuf"

	"github.com/sigstore/cosign/v2/pkg/blob"
	"github.com/sigstore/cosign/v2/pkg/oci/static"
	"github.com/sigstore/cosign/v2/pkg/types"

	"github.com/cyberphone/json-canonicalization/go/src/webpki.org/jsoncanonicalizer"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"

	ssldsse "github.com/secure-systems-lab/go-securesystemslib/dsse"
	"github.com/sigstore/cosign/v2/pkg/oci"
	"github.com/sigstore/cosign/v2/pkg/oci/layout"
	ociremote "github.com/sigstore/cosign/v2/pkg/oci/remote"
	"github.com/sigstore/rekor/pkg/generated/client"
	"github.com/sigstore/rekor/pkg/generated/models"
	"github.com/sigstore/sigstore/pkg/cryptoutils"
	"github.com/sigstore/sigstore/pkg/signature"
	"github.com/sigstore/sigstore/pkg/signature/dsse"
	"github.com/sigstore/sigstore/pkg/signature/options"
	sigPayload "github.com/sigstore/sigstore/pkg/signature/payload"
	tsaverification "github.com/sigstore/timestamp-authority/pkg/verification"
)

// Identity specifies an issuer/subject to verify a signature against.
// Both IssuerRegExp/SubjectRegExp support regexp while Issuer/Subject are for
// strict matching.
type Identity struct {
	Issuer        string
	Subject       string
	IssuerRegExp  string
	SubjectRegExp string
}

// CheckOpts are the options for checking signatures.
type CheckOpts struct {
	// RegistryClientOpts are the options for interacting with the container registry.
	RegistryClientOpts []ociremote.Option

	// Annotations optionally specifies image signature annotations to verify.
	Annotations map[string]interface{}

	// ClaimVerifier, if provided, verifies claims present in the oci.Signature.
	ClaimVerifier func(sig oci.Signature, imageDigest v1.Hash, annotations map[string]interface{}) error

	// RekorClient, if set, is used to make online tlog calls use to verify signatures and public keys.
	RekorClient *client.Rekor
	// RekorPubKeys, if set, is used to validate signatures on log entries from
	// Rekor. It is a map from LogID to crypto.PublicKey. LogID is
	// derived from the PublicKey (see RFC 6962 S3.2).
	// Note that even though the type is of crypto.PublicKey, Rekor only allows
	// for ecdsa.PublicKey: https://github.com/sigstore/cosign/issues/2540
	RekorPubKeys *TrustedTransparencyLogPubKeys

	// SigVerifier is used to verify signatures.
	SigVerifier signature.Verifier
	// PKOpts are the options provided to `SigVerifier.PublicKey()`.
	PKOpts []signature.PublicKeyOption

	// RootCerts are the root CA certs used to verify a signature's chained certificate.
	RootCerts *x509.CertPool
	// IntermediateCerts are the optional intermediate CA certs used to verify a certificate chain.
	IntermediateCerts *x509.CertPool

	// CertGithubWorkflowTrigger is the GitHub Workflow Trigger name expected for a certificate to be valid. The empty string means any certificate can be valid.
	CertGithubWorkflowTrigger string
	// CertGithubWorkflowSha is the GitHub Workflow SHA expected for a certificate to be valid. The empty string means any certificate can be valid.
	CertGithubWorkflowSha string
	// CertGithubWorkflowName is the GitHub Workflow Name expected for a certificate to be valid. The empty string means any certificate can be valid.
	CertGithubWorkflowName string
	// CertGithubWorkflowRepository is the GitHub Workflow Repository  expected for a certificate to be valid. The empty string means any certificate can be valid.
	CertGithubWorkflowRepository string
	// CertGithubWorkflowRef is the GitHub Workflow Ref expected for a certificate to be valid. The empty string means any certificate can be valid.
	CertGithubWorkflowRef string

	// IgnoreSCT requires that a certificate contain an embedded SCT during verification. An SCT is proof of inclusion in a
	// certificate transparency log.
	IgnoreSCT bool
	// Detached SCT. Optional, as the SCT is usually embedded in the certificate.
	SCT []byte
	// CTLogPubKeys, if set, is used to validate SCTs against those keys.
	// It is a map from log id to LogIDMetadata. It is a map from LogID to crypto.PublicKey. LogID is derived from the PublicKey (see RFC 6962 S3.2).
	CTLogPubKeys *TrustedTransparencyLogPubKeys

	// SignatureRef is the reference to the signature file
	SignatureRef string

	// Identities is an array of Identity (Subject, Issuer) matchers that have
	// to be met for the signature to ve valid.
	Identities []Identity

	// Force offline verification of the signature
	Offline bool

	// Set of flags to verify an RFC3161 timestamp used for trusted timestamping
	// TSACertificate is the certificate used to sign the timestamp. Optional, if provided in the timestamp
	TSACertificate *x509.Certificate
	// TSARootCertificates are the set of roots to verify the TSA certificate
	TSARootCertificates []*x509.Certificate
	// TSAIntermediateCertificates are the set of intermediates for chain building
	TSAIntermediateCertificates []*x509.Certificate

	// SkipTlogVerify skip tlog verification
	SkipTlogVerify bool
}

// This is a substitutable signature verification function that can be used for verifying
// attestations of blobs.
type signatureVerificationFn func(
	ctx context.Context, verifier signature.Verifier, sig payloader) error

// For unit testing
type payloader interface {
	// no-op for attestations
	Base64Signature() (string, error)
	Payload() ([]byte, error)
}

func verifyOCIAttestation(_ context.Context, verifier signature.Verifier, att payloader) error {
	payload, err := att.Payload()
	if err != nil {
		return err
	}

	env := ssldsse.Envelope{}
	if err := json.Unmarshal(payload, &env); err != nil {
		return err
	}

	if env.PayloadType != types.IntotoPayloadType {
		return NewVerificationError("invalid payloadType %s on envelope. Expected %s", env.PayloadType, types.IntotoPayloadType)
	}
	dssev, err := ssldsse.NewEnvelopeVerifier(&dsse.VerifierAdapter{SignatureVerifier: verifier})
	if err != nil {
		return err
	}
	_, err = dssev.Verify(&env)
	return err
}

func verifyOCISignature(ctx context.Context, verifier signature.Verifier, sig payloader) error {
	b64sig, err := sig.Base64Signature()
	if err != nil {
		return err
	}
	signature, err := base64.StdEncoding.DecodeString(b64sig)
	if err != nil {
		return err
	}
	payload, err := sig.Payload()
	if err != nil {
		return err
	}
	return verifier.VerifySignature(bytes.NewReader(signature), bytes.NewReader(payload), options.WithContext(ctx))
}

// ValidateAndUnpackCert creates a Verifier from a certificate. Veries that the certificate
// chains up to a trusted root. Optionally verifies the subject and issuer of the certificate.
func ValidateAndUnpackCert(cert *x509.Certificate, co *CheckOpts) (signature.Verifier, error) {
	verifier, err := signature.LoadVerifier(cert.PublicKey, crypto.SHA256)
	if err != nil {
		return nil, fmt.Errorf("invalid certificate found on signature: %w", err)
	}

	// Handle certificates where the Subject Alternative Name is not set to a supported
	// GeneralName (RFC 5280 4.2.1.6). Go only supports DNS, IP addresses, email addresses,
	// or URIs as SANs. Fulcio can issue a certificate with an OtherName GeneralName, so
	// remove the unhandled critical SAN extension before verifying.
	if len(cert.UnhandledCriticalExtensions) > 0 {
		var unhandledExts []asn1.ObjectIdentifier
		for _, oid := range cert.UnhandledCriticalExtensions {
			if !oid.Equal(cryptoutils.SANOID) {
				unhandledExts = append(unhandledExts, oid)
			}
		}
		cert.UnhandledCriticalExtensions = unhandledExts
	}

	// Now verify the cert, then the signature.
	chains, err := TrustedCert(cert, co.RootCerts, co.IntermediateCerts)
	if err != nil {
		return nil, err
	}

	err = CheckCertificatePolicy(cert, co)
	if err != nil {
		return nil, err
	}

	// If IgnoreSCT is set, skip the SCT check
	if co.IgnoreSCT {
		return verifier, nil
	}
	contains, err := ContainsSCT(cert.Raw)
	if err != nil {
		return nil, err
	}
	if !contains && len(co.SCT) == 0 {
		return nil, &VerificationError{"certificate does not include required embedded SCT and no detached SCT was set"}
	}
	// handle if chains has more than one chain - grab first and print message
	if len(chains) > 1 {
		fmt.Fprintf(os.Stderr, "**Info** Multiple valid certificate chains found. Selecting the first to verify the SCT.\n")
	}
	if contains {
		if err := VerifyEmbeddedSCT(context.Background(), chains[0], co.CTLogPubKeys); err != nil {
			return nil, err
		}
	} else {
		chain := chains[0]
		if len(chain) < 2 {
			return nil, errors.New("certificate chain must contain at least a certificate and its issuer")
		}
		certPEM, err := cryptoutils.MarshalCertificateToPEM(chain[0])
		if err != nil {
			return nil, err
		}
		chainPEM, err := cryptoutils.MarshalCertificatesToPEM(chain[1:])
		if err != nil {
			return nil, err
		}
		if err := VerifySCT(context.Background(), certPEM, chainPEM, co.SCT, co.CTLogPubKeys); err != nil {
			return nil, err
		}
	}

	return verifier, nil
}

// CheckCertificatePolicy checks that the certificate subject and issuer match
// the expected values.
func CheckCertificatePolicy(cert *x509.Certificate, co *CheckOpts) error {
	ce := CertExtensions{Cert: cert}

	if err := validateCertExtensions(ce, co); err != nil {
		return err
	}
	oidcIssuer := ce.GetIssuer()
	// If there are identities given, go through them and if one of them
	// matches, call that good, otherwise, return an error.
	if len(co.Identities) > 0 {
		for _, identity := range co.Identities {
			issuerMatches := false
			switch {
			// Check the issuer first
			case identity.IssuerRegExp != "":
				if regex, err := regexp.Compile(identity.IssuerRegExp); err != nil {
					return fmt.Errorf("malformed issuer in identity: %s : %w", identity.IssuerRegExp, err)
				} else if regex.MatchString(oidcIssuer) {
					issuerMatches = true
				}
			case identity.Issuer != "":
				if identity.Issuer == oidcIssuer {
					issuerMatches = true
				}
			default:
				// No issuer constraint on this identity, so checks out
				issuerMatches = true
			}

			// Then the subject
			subjectMatches := false
			switch {
			case identity.SubjectRegExp != "":
				regex, err := regexp.Compile(identity.SubjectRegExp)
				if err != nil {
					return fmt.Errorf("malformed subject in identity: %s : %w", identity.SubjectRegExp, err)
				}
				for _, san := range getSubjectAlternateNames(cert) {
					if regex.MatchString(san) {
						subjectMatches = true
						break
					}
				}
			case identity.Subject != "":
				for _, san := range getSubjectAlternateNames(cert) {
					if san == identity.Subject {
						subjectMatches = true
						break
					}
				}
			default:
				// No subject constraint on this identity, so checks out
				subjectMatches = true
			}
			if subjectMatches && issuerMatches {
				// If both issuer / subject match, return verifier
				return nil
			}
		}
		return &VerificationError{"none of the expected identities matched what was in the certificate"}
	}
	return nil
}

func validateCertExtensions(ce CertExtensions, co *CheckOpts) error {
	if co.CertGithubWorkflowTrigger != "" {
		if ce.GetCertExtensionGithubWorkflowTrigger() != co.CertGithubWorkflowTrigger {
			return &VerificationError{"expected GitHub Workflow Trigger not found in certificate"}
		}
	}

	if co.CertGithubWorkflowSha != "" {
		if ce.GetExtensionGithubWorkflowSha() != co.CertGithubWorkflowSha {
			return &VerificationError{"expected GitHub Workflow SHA not found in certificate"}
		}
	}

	if co.CertGithubWorkflowName != "" {
		if ce.GetCertExtensionGithubWorkflowName() != co.CertGithubWorkflowName {
			return &VerificationError{"expected GitHub Workflow Name not found in certificate"}
		}
	}

	if co.CertGithubWorkflowRepository != "" {
		if ce.GetCertExtensionGithubWorkflowRepository() != co.CertGithubWorkflowRepository {
			return &VerificationError{"expected GitHub Workflow Repository not found in certificate"}
		}
	}

	if co.CertGithubWorkflowRef != "" {
		if ce.GetCertExtensionGithubWorkflowRef() != co.CertGithubWorkflowRef {
			return &VerificationError{"expected GitHub Workflow Ref not found in certificate"}
		}
	}
	return nil
}

// getSubjectAlternateNames returns all of the following for a Certificate.
// DNSNames
// EmailAddresses
// IPAddresses
// URIs
func getSubjectAlternateNames(cert *x509.Certificate) []string {
	sans := []string{}
	sans = append(sans, cert.DNSNames...)
	sans = append(sans, cert.EmailAddresses...)
	for _, ip := range cert.IPAddresses {
		sans = append(sans, ip.String())
	}
	for _, uri := range cert.URIs {
		sans = append(sans, uri.String())
	}
	// ignore error if there's no OtherName SAN
	otherName, _ := cryptoutils.UnmarshalOtherNameSAN(cert.Extensions)
	if len(otherName) > 0 {
		sans = append(sans, otherName)
	}
	return sans
}

// ValidateAndUnpackCertWithChain creates a Verifier from a certificate. Verifies that the certificate
// chains up to the provided root. Chain should start with the parent of the certificate and end with the root.
// Optionally verifies the subject and issuer of the certificate.
func ValidateAndUnpackCertWithChain(cert *x509.Certificate, chain []*x509.Certificate, co *CheckOpts) (signature.Verifier, error) {
	if len(chain) == 0 {
		return nil, errors.New("no chain provided to validate certificate")
	}
	rootPool := x509.NewCertPool()
	rootPool.AddCert(chain[len(chain)-1])
	co.RootCerts = rootPool

	subPool := x509.NewCertPool()
	for _, c := range chain[:len(chain)-1] {
		subPool.AddCert(c)
	}
	co.IntermediateCerts = subPool

	return ValidateAndUnpackCert(cert, co)
}

func tlogValidateEntry(ctx context.Context, client *client.Rekor, rekorPubKeys *TrustedTransparencyLogPubKeys,
	sig oci.Signature, pem []byte) (*models.LogEntryAnon, error) {
	b64sig, err := sig.Base64Signature()
	if err != nil {
		return nil, err
	}
	payload, err := sig.Payload()
	if err != nil {
		return nil, err
	}
	tlogEntries, err := FindTlogEntry(ctx, client, b64sig, payload, pem)
	if err != nil {
		return nil, err
	}
	if len(tlogEntries) == 0 {
		return nil, fmt.Errorf("no valid tlog entries found with proposed entry")
	}
	// Always return the earliest integrated entry. That
	// always suffices for verification of signature time.
	var earliestLogEntry models.LogEntryAnon
	var earliestLogEntryTime *time.Time
	entryVerificationErrs := make([]string, 0)
	for _, e := range tlogEntries {
		entry := e
		if err := VerifyTLogEntryOffline(&entry, rekorPubKeys); err != nil {
			entryVerificationErrs = append(entryVerificationErrs, err.Error())
			continue
		}
		entryTime := time.Unix(*entry.IntegratedTime, 0)
		if earliestLogEntryTime == nil || entryTime.Before(*earliestLogEntryTime) {
			earliestLogEntryTime = &entryTime
			earliestLogEntry = entry
		}
	}
	if earliestLogEntryTime == nil {
		return nil, fmt.Errorf("no valid tlog entries found %s", strings.Join(entryVerificationErrs, ", "))
	}
	return &earliestLogEntry, nil
}

type fakeOCISignatures struct {
	oci.Signatures
	signatures []oci.Signature
}

func (fos *fakeOCISignatures) Get() ([]oci.Signature, error) {
	return fos.signatures, nil
}

// VerifyImageSignatures does all the main cosign checks in a loop, returning the verified signatures.
// If there were no valid signatures, we return an error.
func VerifyImageSignatures(ctx context.Context, signedImgRef name.Reference, co *CheckOpts) (checkedSignatures []oci.Signature, bundleVerified bool, err error) {
	// Enforce this up front.
	if co.RootCerts == nil && co.SigVerifier == nil {
		return nil, false, errors.New("one of verifier or root certs is required")
	}

	// This is a carefully optimized sequence for fetching the signatures of the
	// entity that minimizes registry requests when supplied with a digest input
	digest, err := ociremote.ResolveDigest(signedImgRef, co.RegistryClientOpts...)
	if err != nil {
		return nil, false, err
	}
	h, err := v1.NewHash(digest.Identifier())
	if err != nil {
		return nil, false, err
	}

	var sigs oci.Signatures
	sigRef := co.SignatureRef
	if sigRef == "" {
		st, err := ociremote.SignatureTag(digest, co.RegistryClientOpts...)
		if err != nil {
			return nil, false, err
		}
		sigs, err = ociremote.Signatures(st, co.RegistryClientOpts...)
		if err != nil {
			return nil, false, err
		}
	} else {
		sigs, err = loadSignatureFromFile(sigRef, signedImgRef, co)
		if err != nil {
			return nil, false, err
		}
	}

	return verifySignatures(ctx, sigs, h, co)
}

// VerifyLocalImageSignatures verifies signatures from a saved, local image, without any network calls, returning the verified signatures.
// If there were no valid signatures, we return an error.
func VerifyLocalImageSignatures(ctx context.Context, path string, co *CheckOpts) (checkedSignatures []oci.Signature, bundleVerified bool, err error) {
	// Enforce this up front.
	if co.RootCerts == nil && co.SigVerifier == nil {
		return nil, false, errors.New("one of verifier or root certs is required")
	}

	se, err := layout.SignedImageIndex(path)
	if err != nil {
		return nil, false, err
	}

	var h v1.Hash
	// Verify either an image index or image.
	ii, err := se.SignedImageIndex(v1.Hash{})
	if err != nil {
		return nil, false, err
	}
	i, err := se.SignedImage(v1.Hash{})
	if err != nil {
		return nil, false, err
	}
	switch {
	case ii != nil:
		h, err = ii.Digest()
		if err != nil {
			return nil, false, err
		}
	case i != nil:
		h, err = i.Digest()
		if err != nil {
			return nil, false, err
		}
	default:
		return nil, false, errors.New("must verify either an image index or image")
	}

	sigs, err := se.Signatures()
	if err != nil {
		return nil, false, err
	}

	return verifySignatures(ctx, sigs, h, co)
}

func verifySignatures(ctx context.Context, sigs oci.Signatures, h v1.Hash, co *CheckOpts) (checkedSignatures []oci.Signature, bundleVerified bool, err error) {
	sl, err := sigs.Get()
	if err != nil {
		return nil, false, err
	}

	validationErrs := []string{}

	for _, sig := range sl {
		sig, err := static.Copy(sig)
		if err != nil {
			validationErrs = append(validationErrs, err.Error())
			continue
		}
		verified, err := VerifyImageSignature(ctx, sig, h, co)
		bundleVerified = bundleVerified || verified
		if err != nil {
			validationErrs = append(validationErrs, err.Error())
			continue
		}

		// Phew, we made it.
		checkedSignatures = append(checkedSignatures, sig)
	}
	if len(checkedSignatures) == 0 {
		return nil, false, fmt.Errorf("%w:\n%s", ErrNoMatchingSignatures, strings.Join(validationErrs, "\n "))
	}
	return checkedSignatures, bundleVerified, nil
}

// verifyInternal holds the main verification flow for signatures and attestations.
//  1. Verifies the signature using the provided verifier.
//  2. Checks for transparency log entry presence:
//     a. Verifies the Rekor entry in the bundle, if provided. This works offline OR
//     b. If we don't have a Rekor entry retrieved via cert, do an online lookup (assuming
//     we are in experimental mode).
//  3. If a certificate is provided, check it's expiration using the transparency log timestamp.
func verifyInternal(ctx context.Context, sig oci.Signature, h v1.Hash,
	verifyFn signatureVerificationFn, co *CheckOpts) (
	bundleVerified bool, err error) {
	var acceptableRFC3161Time, acceptableRekorBundleTime *time.Time // Timestamps for the signature we accept, or nil if not applicable.

	if co.TSARootCertificates != nil {
		acceptableRFC3161Timestamp, err := VerifyRFC3161Timestamp(sig, co)
		if err != nil {
			return false, fmt.Errorf("unable to verify RFC3161 timestamp bundle: %w", err)
		}
		if acceptableRFC3161Timestamp != nil {
			acceptableRFC3161Time = &acceptableRFC3161Timestamp.Time
		}
	}

	if !co.SkipTlogVerify {
		bundleVerified, err = VerifyBundle(sig, co)
		if err != nil {
			return false, fmt.Errorf("error verifying bundle: %w", err)
		}

		if bundleVerified {
			// Update with the verified bundle's integrated time.
			t, err := getBundleIntegratedTime(sig)
			if err != nil {
				return false, fmt.Errorf("error getting bundle integrated time: %w", err)
			}
			acceptableRekorBundleTime = &t
		} else {
			// If the --offline flag was specified, fail here. bundleVerified returns false with
			// no error when there was no bundle provided.
			if co.Offline {
				return false, fmt.Errorf("offline verification failed")
			}

			// no Rekor client provided for an online lookup
			if co.RekorClient == nil {
				return false, fmt.Errorf("rekor client not provided for online verification")
			}

			pemBytes, err := keyBytes(sig, co)
			if err != nil {
				return false, err
			}

			e, err := tlogValidateEntry(ctx, co.RekorClient, co.RekorPubKeys, sig, pemBytes)
			if err != nil {
				return false, err
			}
			t := time.Unix(*e.IntegratedTime, 0)
			acceptableRekorBundleTime = &t
		}
	}

	verifier := co.SigVerifier
	if verifier == nil {
		// If we don't have a public key to check against, we can try a root cert.
		cert, err := sig.Cert()
		if err != nil {
			return false, err
		}
		if cert == nil {
			return false, &VerificationError{"no certificate found on signature"}
		}
		// Create a certificate pool for intermediate CA certificates, excluding the root
		chain, err := sig.Chain()
		if err != nil {
			return false, err
		}
		// If there is no chain annotation present, we preserve the pools set in the CheckOpts.
		if len(chain) > 0 {
			if len(chain) == 1 {
				co.IntermediateCerts = nil
			} else if co.IntermediateCerts == nil {
				// If the intermediate certs have not been loaded in by TUF
				pool := x509.NewCertPool()
				for _, cert := range chain[:len(chain)-1] {
					pool.AddCert(cert)
				}
				co.IntermediateCerts = pool
			}
		}
		verifier, err = ValidateAndUnpackCert(cert, co)
		if err != nil {
			return false, err
		}
	}

	// 1. Perform cryptographic verification of the signature using the certificate's public key.
	if err := verifyFn(ctx, verifier, sig); err != nil {
		return false, err
	}

	// We can't check annotations without claims, both require unmarshalling the payload.
	if co.ClaimVerifier != nil {
		if err := co.ClaimVerifier(sig, h, co.Annotations); err != nil {
			return false, err
		}
	}

	// 2. if a certificate was used, verify the certificate expiration against a time
	cert, err := sig.Cert()
	if err != nil {
		return false, err
	}
	if cert != nil {
		// use the provided Rekor bundle or RFC3161 timestamp to check certificate expiration
		expirationChecked := false

		if acceptableRFC3161Time != nil {
			// Verify the cert against the timestamp time.
			if err := CheckExpiry(cert, *acceptableRFC3161Time); err != nil {
				return false, fmt.Errorf("checking expiry on certificate with timestamp: %w", err)
			}
			expirationChecked = true
		}

		if acceptableRekorBundleTime != nil {
			if err := CheckExpiry(cert, *acceptableRekorBundleTime); err != nil {
				return false, fmt.Errorf("checking expiry on certificate with bundle: %w", err)
			}
			expirationChecked = true
		}

		// if no timestamp has been provided, use the current time
		if !expirationChecked {
			if err := CheckExpiry(cert, time.Now()); err != nil {
				return false, fmt.Errorf("checking expiry on certificate with bundle: %w", err)
			}
		}
	}

	return bundleVerified, nil
}

func keyBytes(sig oci.Signature, co *CheckOpts) ([]byte, error) {
	cert, err := sig.Cert()
	if err != nil {
		return nil, err
	}
	// We have a public key.
	if co.SigVerifier != nil {
		pub, err := co.SigVerifier.PublicKey(co.PKOpts...)
		if err != nil {
			return nil, err
		}
		return cryptoutils.MarshalPublicKeyToPEM(pub)
	}
	return cryptoutils.MarshalCertificateToPEM(cert)
}

// VerifyBlobSignature verifies a blob signature.
func VerifyBlobSignature(ctx context.Context, sig oci.Signature, co *CheckOpts) (bundleVerified bool, err error) {
	// The hash of the artifact is unused.
	return verifyInternal(ctx, sig, v1.Hash{}, verifyOCISignature, co)
}

// VerifyImageSignature verifies a signature
func VerifyImageSignature(ctx context.Context, sig oci.Signature, h v1.Hash, co *CheckOpts) (bundleVerified bool, err error) {
	return verifyInternal(ctx, sig, h, verifyOCISignature, co)
}

func loadSignatureFromFile(sigRef string, signedImgRef name.Reference, co *CheckOpts) (oci.Signatures, error) {
	var b64sig string
	targetSig, err := blob.LoadFileOrURL(sigRef)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		targetSig = []byte(sigRef)
	}

	_, err = base64.StdEncoding.DecodeString(string(targetSig))

	if err == nil {
		b64sig = string(targetSig)
	} else {
		b64sig = base64.StdEncoding.EncodeToString(targetSig)
	}

	digest, err := ociremote.ResolveDigest(signedImgRef, co.RegistryClientOpts...)
	if err != nil {
		return nil, err
	}

	payload, err := (&sigPayload.Cosign{Image: digest}).MarshalJSON()

	if err != nil {
		return nil, err
	}

	sig, err := static.NewSignature(payload, b64sig)
	if err != nil {
		return nil, err
	}
	return &fakeOCISignatures{
		signatures: []oci.Signature{sig},
	}, nil
}

// VerifyImageAttestations does all the main cosign checks in a loop, returning the verified attestations.
// If there were no valid attestations, we return an error.
func VerifyImageAttestations(ctx context.Context, signedImgRef name.Reference, co *CheckOpts) (checkedAttestations []oci.Signature, bundleVerified bool, err error) {
	// Enforce this up front.
	if co.RootCerts == nil && co.SigVerifier == nil {
		return nil, false, errors.New("one of verifier or root certs is required")
	}

	// This is a carefully optimized sequence for fetching the attestations of
	// the entity that minimizes registry requests when supplied with a digest
	// input.
	digest, err := ociremote.ResolveDigest(signedImgRef, co.RegistryClientOpts...)
	if err != nil {
		return nil, false, err
	}
	h, err := v1.NewHash(digest.Identifier())
	if err != nil {
		return nil, false, err
	}
	st, err := ociremote.AttestationTag(digest, co.RegistryClientOpts...)
	if err != nil {
		return nil, false, err
	}
	atts, err := ociremote.Signatures(st, co.RegistryClientOpts...)
	if err != nil {
		return nil, false, err
	}

	return verifyImageAttestations(ctx, atts, h, co)
}

// VerifyLocalImageAttestations verifies attestations from a saved, local image, without any network calls,
// returning the verified attestations.
// If there were no valid signatures, we return an error.
func VerifyLocalImageAttestations(ctx context.Context, path string, co *CheckOpts) (checkedAttestations []oci.Signature, bundleVerified bool, err error) {
	// Enforce this up front.
	if co.RootCerts == nil && co.SigVerifier == nil {
		return nil, false, errors.New("one of verifier or root certs is required")
	}

	se, err := layout.SignedImageIndex(path)
	if err != nil {
		return nil, false, err
	}

	var h v1.Hash
	// Verify either an image index or image.
	ii, err := se.SignedImageIndex(v1.Hash{})
	if err != nil {
		return nil, false, err
	}
	i, err := se.SignedImage(v1.Hash{})
	if err != nil {
		return nil, false, err
	}
	switch {
	case ii != nil:
		h, err = ii.Digest()
		if err != nil {
			return nil, false, err
		}
	case i != nil:
		h, err = i.Digest()
		if err != nil {
			return nil, false, err
		}
	default:
		return nil, false, errors.New("must verify either an image index or image")
	}

	atts, err := se.Attestations()
	if err != nil {
		return nil, false, err
	}
	return verifyImageAttestations(ctx, atts, h, co)
}

func VerifyBlobAttestation(ctx context.Context, att oci.Signature, co *CheckOpts) (
	bool, error) {
	// A blob attestation does not have an associated artifact (currently) to check the claims against.
	// So we can safely add a nil hash.
	return verifyInternal(ctx, att, v1.Hash{}, verifyOCIAttestation, co)
}

func verifyImageAttestations(ctx context.Context, atts oci.Signatures, h v1.Hash, co *CheckOpts) (checkedAttestations []oci.Signature, bundleVerified bool, err error) {
	sl, err := atts.Get()
	if err != nil {
		return nil, false, err
	}

	validationErrs := []string{}
	for _, att := range sl {
		att, err := static.Copy(att)
		if err != nil {
			validationErrs = append(validationErrs, err.Error())
			continue
		}
		if err := func(att oci.Signature) error {
			verified, err := verifyInternal(ctx, att, h, verifyOCIAttestation, co)
			bundleVerified = bundleVerified || verified
			return err
		}(att); err != nil {
			validationErrs = append(validationErrs, err.Error())
			continue
		}

		// Phew, we made it.
		checkedAttestations = append(checkedAttestations, att)
	}
	if len(checkedAttestations) == 0 {
		return nil, false, fmt.Errorf("%w:\n%s", ErrNoMatchingAttestations, strings.Join(validationErrs, "\n "))
	}
	return checkedAttestations, bundleVerified, nil
}

// CheckExpiry confirms the time provided is within the valid period of the cert
func CheckExpiry(cert *x509.Certificate, it time.Time) error {
	ft := func(t time.Time) string {
		return t.Format(time.RFC3339)
	}
	if cert.NotAfter.Before(it) {
		return NewVerificationError("certificate expired before signatures were entered in log: %s is before %s",
			ft(cert.NotAfter), ft(it))
	}
	if cert.NotBefore.After(it) {
		return NewVerificationError("certificate was issued after signatures were entered in log: %s is after %s",
			ft(cert.NotAfter), ft(it))
	}
	return nil
}

func getBundleIntegratedTime(sig oci.Signature) (time.Time, error) {
	bundle, err := sig.Bundle()
	if err != nil {
		return time.Now(), err
	} else if bundle == nil {
		return time.Now(), nil
	}
	return time.Unix(bundle.Payload.IntegratedTime, 0), nil
}

// This verifies an offline bundle contained in the sig against the trusted
// Rekor publicKeys.
func VerifyBundle(sig oci.Signature, co *CheckOpts) (bool, error) {
	bundle, err := sig.Bundle()
	if err != nil {
		return false, err
	} else if bundle == nil {
		return false, nil
	}

	if co.RekorPubKeys == nil || co.RekorPubKeys.Keys == nil {
		return false, errors.New("no trusted rekor public keys provided")
	}
	// Make sure all the rekorPubKeys are ecsda.PublicKeys
	for k, v := range co.RekorPubKeys.Keys {
		if _, ok := v.PubKey.(*ecdsa.PublicKey); !ok {
			return false, fmt.Errorf("rekor Public key for LogID %s is not type ecdsa.PublicKey", k)
		}
	}

	if err := compareSigs(bundle.Payload.Body.(string), sig); err != nil {
		return false, err
	}

	if err := comparePublicKey(bundle.Payload.Body.(string), sig, co); err != nil {
		return false, err
	}

	pubKey, ok := co.RekorPubKeys.Keys[bundle.Payload.LogID]
	if !ok {
		return false, &VerificationError{"verifying bundle: rekor log public key not found for payload"}
	}
	err = VerifySET(bundle.Payload, bundle.SignedEntryTimestamp, pubKey.PubKey.(*ecdsa.PublicKey))
	if err != nil {
		return false, err
	}
	if pubKey.Status != tuf.Active {
		fmt.Fprintf(os.Stderr, "**Info** Successfully verified Rekor entry using an expired verification key\n")
	}

	payload, err := sig.Payload()
	if err != nil {
		return false, fmt.Errorf("reading payload: %w", err)
	}
	signature, err := sig.Base64Signature()
	if err != nil {
		return false, fmt.Errorf("reading base64signature: %w", err)
	}

	alg, bundlehash, err := bundleHash(bundle.Payload.Body.(string), signature)
	h := sha256.Sum256(payload)
	payloadHash := hex.EncodeToString(h[:])

	if alg != "sha256" || bundlehash != payloadHash {
		return false, fmt.Errorf("matching bundle to payload: %w", err)
	}
	return true, nil
}

// VerifyRFC3161Timestamp verifies that the timestamp in sig is correctly signed, and if so,
// returns the timestamp value.
// It returns (nil, nil) if there is no timestamp, or (nil, err) if there is an invalid timestamp.
func VerifyRFC3161Timestamp(sig oci.Signature, co *CheckOpts) (*timestamp.Timestamp, error) {
	ts, err := sig.RFC3161Timestamp()
	if err != nil {
		return nil, err
	} else if ts == nil {
		return nil, nil
	}

	b64Sig, err := sig.Base64Signature()
	if err != nil {
		return nil, fmt.Errorf("reading base64signature: %w", err)
	}

	var tsBytes []byte
	if len(b64Sig) == 0 {
		// For attestations, the Base64Signature is not set, therefore we rely on the signed payload
		signedPayload, err := sig.Payload()
		if err != nil {
			return nil, fmt.Errorf("reading the payload: %w", err)
		}
		tsBytes = signedPayload
	} else {
		// create timestamp over raw bytes of signature
		rawSig, err := base64.StdEncoding.DecodeString(b64Sig)
		if err != nil {
			return nil, err
		}
		tsBytes = rawSig
	}

	return tsaverification.VerifyTimestampResponse(ts.SignedRFC3161Timestamp, bytes.NewReader(tsBytes),
		tsaverification.VerifyOpts{
			TSACertificate: co.TSACertificate,
			Intermediates:  co.TSAIntermediateCertificates,
			Roots:          co.TSARootCertificates,
		})
}

// compare bundle signature to the signature we are verifying
func compareSigs(bundleBody string, sig oci.Signature) error {
	// TODO(nsmith5): modify function signature to make it more clear _why_
	// we've returned nil (there are several reasons possible here).
	actualSig, err := sig.Base64Signature()
	if err != nil {
		return fmt.Errorf("base64 signature: %w", err)
	}
	if actualSig == "" {
		// NB: empty sig means this is an attestation
		return nil
	}
	bundleSignature, err := bundleSig(bundleBody)
	if err != nil {
		return fmt.Errorf("failed to extract signature from bundle: %w", err)
	}
	if bundleSignature == "" {
		return nil
	}
	if bundleSignature != actualSig {
		return &VerificationError{"signature in bundle does not match signature being verified"}
	}
	return nil
}

func comparePublicKey(bundleBody string, sig oci.Signature, co *CheckOpts) error {
	pemBytes, err := keyBytes(sig, co)
	if err != nil {
		return err
	}

	bundleKey, err := bundleKey(bundleBody)
	if err != nil {
		return fmt.Errorf("failed to extract key from bundle: %w", err)
	}

	decodeSecond, err := base64.StdEncoding.DecodeString(bundleKey)
	if err != nil {
		return fmt.Errorf("decoding base64 string %s", bundleKey)
	}

	// Compare the PEM bytes, to ignore spurious newlines in the public key bytes.
	pemFirst, rest := pem.Decode(pemBytes)
	if len(rest) > 0 {
		return fmt.Errorf("unexpected PEM block: %s", rest)
	}
	pemSecond, rest := pem.Decode(decodeSecond)
	if len(rest) > 0 {
		return fmt.Errorf("unexpected PEM block: %s", rest)
	}

	if !bytes.Equal(pemFirst.Bytes, pemSecond.Bytes) {
		return fmt.Errorf("comparing public key PEMs, expected %s, got %s",
			pemBytes, decodeSecond)
	}

	return nil
}

func bundleHash(bundleBody, signature string) (string, string, error) {
	var toto models.Intoto
	var rekord models.Rekord
	var hrekord models.Hashedrekord
	var intotoObj models.IntotoV001Schema
	var rekordObj models.RekordV001Schema
	var hrekordObj models.HashedrekordV001Schema

	bodyDecoded, err := base64.StdEncoding.DecodeString(bundleBody)
	if err != nil {
		return "", "", err
	}

	// The fact that there's no signature (or empty rather), implies
	// that this is an Attestation that we're verifying.
	if len(signature) == 0 {
		err = json.Unmarshal(bodyDecoded, &toto)
		if err != nil {
			return "", "", err
		}

		specMarshal, err := json.Marshal(toto.Spec)
		if err != nil {
			return "", "", err
		}
		err = json.Unmarshal(specMarshal, &intotoObj)
		if err != nil {
			return "", "", err
		}

		return *intotoObj.Content.Hash.Algorithm, *intotoObj.Content.Hash.Value, nil
	}

	if err := json.Unmarshal(bodyDecoded, &rekord); err == nil {
		specMarshal, err := json.Marshal(rekord.Spec)
		if err != nil {
			return "", "", err
		}
		err = json.Unmarshal(specMarshal, &rekordObj)
		if err != nil {
			return "", "", err
		}
		return *rekordObj.Data.Hash.Algorithm, *rekordObj.Data.Hash.Value, nil
	}

	// Try hashedRekordObj
	err = json.Unmarshal(bodyDecoded, &hrekord)
	if err != nil {
		return "", "", err
	}
	specMarshal, err := json.Marshal(hrekord.Spec)
	if err != nil {
		return "", "", err
	}
	err = json.Unmarshal(specMarshal, &hrekordObj)
	if err != nil {
		return "", "", err
	}
	return *hrekordObj.Data.Hash.Algorithm, *hrekordObj.Data.Hash.Value, nil
}

// bundleSig extracts the signature from the rekor bundle body
func bundleSig(bundleBody string) (string, error) {
	var rekord models.Rekord
	var hrekord models.Hashedrekord
	var rekordObj models.RekordV001Schema
	var hrekordObj models.HashedrekordV001Schema

	bodyDecoded, err := base64.StdEncoding.DecodeString(bundleBody)
	if err != nil {
		return "", fmt.Errorf("decoding bundleBody: %w", err)
	}

	// Try Rekord
	if err := json.Unmarshal(bodyDecoded, &rekord); err == nil {
		specMarshal, err := json.Marshal(rekord.Spec)
		if err != nil {
			return "", err
		}
		if err := json.Unmarshal(specMarshal, &rekordObj); err != nil {
			return "", err
		}
		return rekordObj.Signature.Content.String(), nil
	}

	// Try hashedRekordObj
	if err := json.Unmarshal(bodyDecoded, &hrekord); err != nil {
		return "", err
	}
	specMarshal, err := json.Marshal(hrekord.Spec)
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(specMarshal, &hrekordObj); err != nil {
		return "", err
	}
	return hrekordObj.Signature.Content.String(), nil
}

// bundleKey extracts the key from the rekor bundle body
func bundleKey(bundleBody string) (string, error) {
	var rekord models.Rekord
	var hrekord models.Hashedrekord
	var intotod models.Intoto
	var rekordObj models.RekordV001Schema
	var hrekordObj models.HashedrekordV001Schema
	var intotodObj models.IntotoV001Schema

	bodyDecoded, err := base64.StdEncoding.DecodeString(bundleBody)
	if err != nil {
		return "", fmt.Errorf("decoding bundleBody: %w", err)
	}

	// Try Rekord
	if err := json.Unmarshal(bodyDecoded, &rekord); err == nil {
		specMarshal, err := json.Marshal(rekord.Spec)
		if err != nil {
			return "", err
		}
		if err := json.Unmarshal(specMarshal, &rekordObj); err != nil {
			return "", err
		}
		return rekordObj.Signature.PublicKey.Content.String(), nil
	}

	// Try hashedRekordObj
	if err := json.Unmarshal(bodyDecoded, &hrekord); err == nil {
		specMarshal, err := json.Marshal(hrekord.Spec)
		if err != nil {
			return "", err
		}
		if err := json.Unmarshal(specMarshal, &hrekordObj); err != nil {
			return "", err
		}
		return hrekordObj.Signature.PublicKey.Content.String(), nil
	}

	// Try Intoto
	if err := json.Unmarshal(bodyDecoded, &intotod); err != nil {
		return "", err
	}
	specMarshal, err := json.Marshal(intotod.Spec)
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(specMarshal, &intotodObj); err != nil {
		return "", err
	}
	return intotodObj.PublicKey.String(), nil
}

func VerifySET(bundlePayload cbundle.RekorPayload, signature []byte, pub *ecdsa.PublicKey) error {
	contents, err := json.Marshal(bundlePayload)
	if err != nil {
		return fmt.Errorf("marshaling: %w", err)
	}
	canonicalized, err := jsoncanonicalizer.Transform(contents)
	if err != nil {
		return fmt.Errorf("canonicalizing: %w", err)
	}

	// verify the SET against the public key
	hash := sha256.Sum256(canonicalized)
	if !ecdsa.VerifyASN1(pub, hash[:], signature) {
		return &VerificationError{"unable to verify SET"}
	}
	return nil
}

func TrustedCert(cert *x509.Certificate, roots *x509.CertPool, intermediates *x509.CertPool) ([][]*x509.Certificate, error) {
	chains, err := cert.Verify(x509.VerifyOptions{
		// THIS IS IMPORTANT: WE DO NOT CHECK TIMES HERE
		// THE CERTIFICATE IS TREATED AS TRUSTED FOREVER
		// WE CHECK THAT THE SIGNATURES WERE CREATED DURING THIS WINDOW
		CurrentTime:   cert.NotBefore,
		Roots:         roots,
		Intermediates: intermediates,
		KeyUsages: []x509.ExtKeyUsage{
			x509.ExtKeyUsageCodeSigning,
		},
	})
	if err != nil {
		return nil, err
	}
	return chains, nil
}

func correctAnnotations(wanted, have map[string]interface{}) bool {
	for k, v := range wanted {
		if have[k] != v {
			return false
		}
	}
	return true
}
