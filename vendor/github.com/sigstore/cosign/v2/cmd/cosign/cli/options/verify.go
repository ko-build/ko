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

package options

import (
	"github.com/spf13/cobra"
)

type CommonVerifyOptions struct {
	Offline          bool // Force offline verification
	TSACertChainPath string
	SkipTlogVerify   bool
}

func (o *CommonVerifyOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.Offline, "offline", false,
		"only allow offline verification")

	cmd.Flags().StringVar(&o.TSACertChainPath, "timestamp-certificate-chain", "",
		"path to PEM-encoded certificate chain file for the RFC3161 timestamp authority. Must contain the root CA certificate. "+
			"Optionally may contain intermediate CA certificates, and may contain the leaf TSA certificate if not present in the timestamp")

	cmd.Flags().BoolVar(&o.SkipTlogVerify, "insecure-skip-tlog-verify", false,
		"skip transparency log verification, to be used when an artifact signature has not been uploaded to the transparency log. Artifacts "+
			"cannot be publicly verified when not included in a log")
}

// VerifyOptions is the top level wrapper for the `verify` command.
type VerifyOptions struct {
	Key          string
	CheckClaims  bool
	Attachment   string
	Output       string
	SignatureRef string
	LocalImage   bool

	CommonVerifyOptions CommonVerifyOptions
	SecurityKey         SecurityKeyOptions
	CertVerify          CertVerifyOptions
	Rekor               RekorOptions
	Registry            RegistryOptions
	SignatureDigest     SignatureDigestOptions

	AnnotationOptions
}

var _ Interface = (*VerifyOptions)(nil)

// AddFlags implements Interface
func (o *VerifyOptions) AddFlags(cmd *cobra.Command) {
	o.SecurityKey.AddFlags(cmd)
	o.Rekor.AddFlags(cmd)
	o.CertVerify.AddFlags(cmd)
	o.Registry.AddFlags(cmd)
	o.SignatureDigest.AddFlags(cmd)
	o.AnnotationOptions.AddFlags(cmd)
	o.CommonVerifyOptions.AddFlags(cmd)

	cmd.Flags().StringVar(&o.Key, "key", "",
		"path to the public key file, KMS URI or Kubernetes Secret")
	_ = cmd.Flags().SetAnnotation("key", cobra.BashCompFilenameExt, []string{})

	cmd.Flags().BoolVar(&o.CheckClaims, "check-claims", true,
		"whether to check the claims found")

	cmd.Flags().StringVar(&o.Attachment, "attachment", "",
		"related image attachment to verify (sbom), default none")

	cmd.Flags().StringVarP(&o.Output, "output", "o", "json",
		"output format for the signing image information (json|text)")

	cmd.Flags().StringVar(&o.SignatureRef, "signature", "",
		"signature content or path or remote URL")

	cmd.Flags().BoolVar(&o.LocalImage, "local-image", false,
		"whether the specified image is a path to an image saved locally via 'cosign save'")
}

// VerifyAttestationOptions is the top level wrapper for the `verify attestation` command.
type VerifyAttestationOptions struct {
	Key         string
	CheckClaims bool
	Output      string

	CommonVerifyOptions CommonVerifyOptions
	SecurityKey         SecurityKeyOptions
	Rekor               RekorOptions
	CertVerify          CertVerifyOptions
	Registry            RegistryOptions
	Predicate           PredicateRemoteOptions
	Policies            []string
	LocalImage          bool
}

var _ Interface = (*VerifyAttestationOptions)(nil)

// AddFlags implements Interface
func (o *VerifyAttestationOptions) AddFlags(cmd *cobra.Command) {
	o.SecurityKey.AddFlags(cmd)
	o.Rekor.AddFlags(cmd)
	o.CertVerify.AddFlags(cmd)
	o.Registry.AddFlags(cmd)
	o.Predicate.AddFlags(cmd)
	o.CommonVerifyOptions.AddFlags(cmd)

	cmd.Flags().StringVar(&o.Key, "key", "",
		"path to the public key file, KMS URI or Kubernetes Secret")

	cmd.Flags().BoolVar(&o.CheckClaims, "check-claims", true,
		"whether to check the claims found")

	cmd.Flags().StringSliceVar(&o.Policies, "policy", nil,
		"specify CUE or Rego files will be using for validation")

	cmd.Flags().StringVarP(&o.Output, "output", "o", "json",
		"output format for the signing image information (json|text)")

	cmd.Flags().BoolVar(&o.LocalImage, "local-image", false,
		"whether the specified image is a path to an image saved locally via 'cosign save'")
}

// VerifyBlobOptions is the top level wrapper for the `verify blob` command.
type VerifyBlobOptions struct {
	Key        string
	Signature  string
	BundlePath string

	SecurityKey         SecurityKeyOptions
	CertVerify          CertVerifyOptions
	Rekor               RekorOptions
	Registry            RegistryOptions
	CommonVerifyOptions CommonVerifyOptions

	RFC3161TimestampPath string
}

var _ Interface = (*VerifyBlobOptions)(nil)

// AddFlags implements Interface
func (o *VerifyBlobOptions) AddFlags(cmd *cobra.Command) {
	o.SecurityKey.AddFlags(cmd)
	o.Rekor.AddFlags(cmd)
	o.CertVerify.AddFlags(cmd)
	o.Registry.AddFlags(cmd)
	o.CommonVerifyOptions.AddFlags(cmd)

	cmd.Flags().StringVar(&o.Key, "key", "",
		"path to the public key file, KMS URI or Kubernetes Secret")

	cmd.Flags().StringVar(&o.Signature, "signature", "",
		"signature content or path or remote URL")

	cmd.Flags().StringVar(&o.BundlePath, "bundle", "",
		"path to bundle FILE")

	cmd.Flags().StringVar(&o.RFC3161TimestampPath, "rfc3161-timestamp", "",
		"path to RFC3161 timestamp FILE")
}

// VerifyDockerfileOptions is the top level wrapper for the `dockerfile verify` command.
type VerifyDockerfileOptions struct {
	VerifyOptions
	BaseImageOnly bool
}

var _ Interface = (*VerifyDockerfileOptions)(nil)

// AddFlags implements Interface
func (o *VerifyDockerfileOptions) AddFlags(cmd *cobra.Command) {
	o.VerifyOptions.AddFlags(cmd)

	cmd.Flags().BoolVar(&o.BaseImageOnly, "base-image-only", false,
		"only verify the base image (the last FROM image in the Dockerfile)")
}

// VerifyBlobAttestationOptions is the top level wrapper for the `verify-blob-attestation` command.
type VerifyBlobAttestationOptions struct {
	Key           string
	SignaturePath string
	PredicateOptions
}

var _ Interface = (*VerifyBlobOptions)(nil)

// AddFlags implements Interface
func (o *VerifyBlobAttestationOptions) AddFlags(cmd *cobra.Command) {
	o.PredicateOptions.AddFlags(cmd)

	cmd.Flags().StringVar(&o.Key, "key", "",
		"path to the public key file, KMS URI or Kubernetes Secret")

	cmd.Flags().StringVar(&o.SignaturePath, "signature", "",
		"path to base64-encoded signature over attestation in DSSE format")
}
