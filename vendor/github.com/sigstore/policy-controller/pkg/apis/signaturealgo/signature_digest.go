//
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

package signaturealgo

import (
	"crypto"
	"fmt"
	"strings"
)

var DefaultSignatureAlgorithm = "sha256"

// supportedSignatureAlgorithms sets a list of support signature algorithms that is similar to the list supported by cosign
var supportedSignatureAlgorithms = map[string]crypto.Hash{
	"sha224": crypto.SHA224,
	"sha256": crypto.SHA256,
	"sha384": crypto.SHA384,
	"sha512": crypto.SHA512,
}

// HashAlgorithm returns a crypto.Hash code using an algorithm name as input parameter
func HashAlgorithm(algorithmName string) (crypto.Hash, error) {
	if algorithmName == "" {
		return crypto.SHA256, nil
	}
	normalizedAlgo := strings.ToLower(strings.TrimSpace(algorithmName))

	algo, exists := supportedSignatureAlgorithms[normalizedAlgo]
	if !exists {
		return crypto.SHA256, fmt.Errorf("unknown digest algorithm: %s", algorithmName)
	}

	return algo, nil
}
