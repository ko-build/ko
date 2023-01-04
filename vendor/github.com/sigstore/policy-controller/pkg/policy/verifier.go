// Copyright 2023 The Sigstore Authors.
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

package policy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	ociremote "github.com/sigstore/cosign/v2/pkg/oci/remote"
	"github.com/sigstore/policy-controller/pkg/apis/config"
	"github.com/sigstore/policy-controller/pkg/webhook"
	webhookcip "github.com/sigstore/policy-controller/pkg/webhook/clusterimagepolicy"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
)

// Verifier is the interface for checking that a given image digest satisfies
// the policies backing this interface.
type Verifier interface {
	// Verify checks that the provided reference satisfies the backing policies.
	Verify(context.Context, name.Reference, authn.Keychain) error
}

// WarningWriter is used to surface warning messages in a manner that
// is customizable by callers that's suitable for their execution
// environment.
type WarningWriter func(string, ...interface{})

// Compile turns a Verification into an executable Verifier.
// Any compilation errors are returned here.
func Compile(ctx context.Context, v Verification, ww WarningWriter) (Verifier, error) {
	if err := v.Validate(ctx); err != nil {
		return nil, err
	}

	ipc, err := gather(ctx, v, ww)
	if err != nil {
		// This should never hit for validated policies.
		return nil, err
	}

	return &impl{
		verification: v,
		ipc:          ipc,
		ww:           ww,
	}, nil
}

func gather(ctx context.Context, v Verification, ww WarningWriter) (*config.ImagePolicyConfig, error) {
	pol := *v.Policies
	ipc := &config.ImagePolicyConfig{
		Policies: make(map[string]webhookcip.ClusterImagePolicy, len(pol)),
	}

	for i, p := range pol {
		var content string
		switch {
		case p.Data != "":
			content = p.Data

		case p.Path != "":
			raw, err := os.ReadFile(p.Path)
			if err != nil {
				return nil, err
			}
			content = string(raw)

		case p.URL != "":
			resp, err := http.Get(p.URL)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()
			raw, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			content = string(raw)

		default:
			// This should never happen for a validated policy.
			return nil, fmt.Errorf("unsupported policy shape: %v", p)
		}

		l, warns, err := ParseClusterImagePolicies(ctx, content)
		if err != nil {
			// This path should be unreachable, since we already parse
			// things during compilation.
			return nil, fmt.Errorf("parsing policies: %w", err)
		}
		if warns != nil {
			ww("policy %d: %v", i, warns)
		}

		// TODO(mattmoor): Add additional checks for unsupported things,
		// like Match, IncludeSpec, etc.

		for _, cip := range l {
			cip.SetDefaults(ctx)
			if _, ok := ipc.Policies[cip.Name]; ok {
				ww("duplicate policy named %q, skipping", cip.Name)
				continue
			}
			// We need to roundtrip the policy through JSON here because
			// the compiled policy expects to be decoded from JSON and only
			// sets up certain fields when being unmarshalled from JSON, so
			// things like keyful verification only work when we roundtrip
			// through JSON.
			var compiled webhookcip.ClusterImagePolicy
			if err := convert(webhookcip.ConvertClusterImagePolicyV1alpha1ToWebhook(cip), &compiled); err != nil {
				ww("roundtripping policy %v", err)
				continue
			}
			ipc.Policies[cip.Name] = compiled
		}
	}

	return ipc, nil
}

type impl struct {
	verification Verification

	ipc *config.ImagePolicyConfig
	ww  WarningWriter
}

// Check that impl implements Verifier
var _ Verifier = (*impl)(nil)

// Verify implements Verifier
func (i *impl) Verify(ctx context.Context, ref name.Reference, kc authn.Keychain) error {
	tm := getTypeMeta(ctx)
	om := getObjectMeta(ctx)
	matches, err := i.ipc.GetMatchingPolicies(ref.Name(), tm.Kind, tm.APIVersion, om.Labels)
	if err != nil {
		return err
	}

	if len(matches) == 0 {
		switch i.verification.NoMatchPolicy {
		case "allow":
			return nil
		case "warn":
			i.ww("%s is uncovered by policy", ref)
		case "deny":
			return fmt.Errorf("%s is uncovered by policy", ref)
		default:
			// This is unreachable for a validated Verification.
			return fmt.Errorf("unsupported noMatchPolicy: %q", i.verification.NoMatchPolicy)
		}
	}

	for _, p := range matches {
		_, errs := webhook.ValidatePolicy(ctx, "" /* namespace */, ref, p,
			kc, ociremote.WithRemoteOptions(remote.WithAuthFromKeychain(kc)))
		for _, err := range errs {
			var fe *apis.FieldError
			if errors.As(err, &fe) {
				if warnFE := fe.Filter(apis.WarningLevel); warnFE != nil {
					i.ww("%v", warnFE)
				}
				if errorFE := fe.Filter(apis.ErrorLevel); errorFE != nil {
					return errorFE
				}
			} else {
				return err
			}
		}
	}

	return nil
}

func getTypeMeta(ctx context.Context) (tm metav1.TypeMeta) {
	raw := webhook.GetIncludeTypeMeta(ctx)
	if raw == nil {
		return
	}
	_ = convert(raw, &tm)
	return
}

func getObjectMeta(ctx context.Context) (om metav1.ObjectMeta) {
	raw := webhook.GetIncludeObjectMeta(ctx)
	if raw == nil {
		return
	}
	_ = convert(raw, &om)
	return
}
