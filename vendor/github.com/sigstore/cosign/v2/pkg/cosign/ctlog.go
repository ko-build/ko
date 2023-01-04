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
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/sigstore/cosign/v2/pkg/cosign/env"
	"github.com/sigstore/sigstore/pkg/tuf"
)

// This is the CT log public key target name
var ctPublicKeyStr = `ctfe.pub`

// GetCTLogPubs retrieves trusted CTLog public keys from the embedded or cached
// TUF root. If expired, makes a network call to retrieve the updated targets.
// By default the public keys comes from TUF, but you can override this for test
// purposes by using an env variable `SIGSTORE_CT_LOG_PUBLIC_KEY_FILE`. If using
// an alternate, the file can be PEM, or DER format.
func GetCTLogPubs(ctx context.Context) (*TrustedTransparencyLogPubKeys, error) {
	publicKeys := NewTrustedTransparencyLogPubKeys()
	altCTLogPub := env.Getenv(env.VariableSigstoreCTLogPublicKeyFile)

	if altCTLogPub != "" {
		raw, err := os.ReadFile(altCTLogPub)
		if err != nil {
			return nil, fmt.Errorf("error reading alternate CTLog public key file: %w", err)
		}
		if err := publicKeys.AddTransparencyLogPubKey(raw, tuf.Active); err != nil {
			return nil, fmt.Errorf("AddCTLogPubKey: %w", err)
		}
	} else {
		tufClient, err := tuf.NewFromEnv(ctx)
		if err != nil {
			return nil, err
		}
		targets, err := tufClient.GetTargetsByMeta(tuf.CTFE, []string{ctPublicKeyStr})
		if err != nil {
			return nil, err
		}
		for _, t := range targets {
			if err := publicKeys.AddTransparencyLogPubKey(t.Target, t.Status); err != nil {
				return nil, fmt.Errorf("AddCTLogPubKey: %w", err)
			}
		}
	}

	if len(publicKeys.Keys) == 0 {
		return nil, errors.New("none of the CTLog public keys have been found")
	}

	return &publicKeys, nil
}
