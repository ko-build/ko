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

package common

import (
	"strings"

	registryfuncs "github.com/google/go-containerregistry/pkg/name"
)

const (
	ociRepoDelimiter = "/"
)

func ValidateOCI(oci string) error {
	// We want to validate both registry uris only or registry with valid repository names
	parts := strings.SplitN(oci, ociRepoDelimiter, 2)
	if len(parts) == 2 && (strings.ContainsRune(parts[0], '.') || strings.ContainsRune(parts[0], ':')) {
		_, err := registryfuncs.NewRepository(oci, registryfuncs.StrictValidation)
		if err != nil {
			return err
		}
		return nil
	}
	_, err := registryfuncs.NewRegistry(oci, registryfuncs.StrictValidation)
	if err != nil {
		return err
	}
	return nil
}
