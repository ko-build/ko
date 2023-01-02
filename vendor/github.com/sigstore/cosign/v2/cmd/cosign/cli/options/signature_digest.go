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
	"crypto"
	"fmt"
	"sort"
	"strings"

	_ "crypto/sha256" // for sha224 + sha256
	_ "crypto/sha512" // for sha384 + sha512

	"github.com/spf13/cobra"
)

var supportedSignatureAlgorithms = map[string]crypto.Hash{
	"sha224": crypto.SHA224,
	"sha256": crypto.SHA256,
	"sha384": crypto.SHA384,
	"sha512": crypto.SHA512,
}

func supportedSignatureAlgorithmNames() []string {
	names := make([]string, 0, len(supportedSignatureAlgorithms))

	for name := range supportedSignatureAlgorithms {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}

// SignatureDigestOptions holds options for specifying which digest algorithm should
// be used when processing a signature.
type SignatureDigestOptions struct {
	AlgorithmName string
}

var _ Interface = (*SignatureDigestOptions)(nil)

// AddFlags implements Interface
func (o *SignatureDigestOptions) AddFlags(cmd *cobra.Command) {
	validSignatureDigestAlgorithms := strings.Join(supportedSignatureAlgorithmNames(), "|")

	cmd.Flags().StringVar(&o.AlgorithmName, "signature-digest-algorithm", "sha256",
		fmt.Sprintf("digest algorithm to use when processing a signature (%s)", validSignatureDigestAlgorithms))
}

// HashAlgorithm converts the algorithm's name - provided as a string - into a crypto.Hash algorithm.
// Returns an error if the algorithm name doesn't match a supported algorithm, and defaults to SHA256
// in the event that the given algorithm is invalid.
func (o *SignatureDigestOptions) HashAlgorithm() (crypto.Hash, error) {
	normalizedAlgo := strings.ToLower(strings.TrimSpace(o.AlgorithmName))

	if normalizedAlgo == "" {
		return crypto.SHA256, nil
	}

	algo, exists := supportedSignatureAlgorithms[normalizedAlgo]
	if !exists {
		return crypto.SHA256, fmt.Errorf("unknown digest algorithm: %s", o.AlgorithmName)
	}

	if !algo.Available() {
		return crypto.SHA256, fmt.Errorf("hash %q is not available on this platform", o.AlgorithmName)
	}

	return algo, nil
}
