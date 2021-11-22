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

package mutate

import "github.com/sigstore/cosign/pkg/oci"

// DupeDetector scans a list of signatures looking for a duplicate.
type DupeDetector interface {
	Find(oci.Signatures, oci.Signature) (oci.Signature, error)
}

type ReplaceOp interface {
	Replace(oci.Signatures, oci.Signature) (oci.Signatures, error)
}

type SignOption func(*signOpts)

type signOpts struct {
	dd DupeDetector
	ro ReplaceOp
}

func makeSignOpts(opts ...SignOption) *signOpts {
	so := &signOpts{}
	for _, opt := range opts {
		opt(so)
	}
	return so
}

// WithDupeDetector configures Sign* to use the following DupeDetector
// to avoid attaching duplicate signatures.
func WithDupeDetector(dd DupeDetector) SignOption {
	return func(so *signOpts) {
		so.dd = dd
	}
}

func WithReplaceOp(ro ReplaceOp) SignOption {
	return func(so *signOpts) {
		so.ro = ro
	}
}
