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
	"context"
	"crypto/tls"
	"io"
	"net/http"

	ecr "github.com/awslabs/amazon-ecr-credential-helper/ecr-login"
	"github.com/chrismellard/docker-credential-acr-env/pkg/credhelper"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/authn/github"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/google"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	alibabaacr "github.com/mozillazg/docker-credential-acr-helper/pkg/credhelper"
	ociremote "github.com/sigstore/cosign/v2/pkg/oci/remote"
	"github.com/spf13/cobra"
)

// Keychain is an alias of authn.Keychain to expose this configuration option to consumers of this lib
type Keychain = authn.Keychain

// RegistryOptions is the wrapper for the registry options.
type RegistryOptions struct {
	AllowInsecure      bool
	AllowHTTPRegistry  bool
	KubernetesKeychain bool
	RefOpts            ReferenceOptions
	Keychain           Keychain
}

var _ Interface = (*RegistryOptions)(nil)

// AddFlags implements Interface
func (o *RegistryOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.AllowInsecure, "allow-insecure-registry", false,
		"whether to allow insecure connections to registries (e.g., with expired or self-signed TLS certificates). Don't use this for anything but testing")

	cmd.Flags().BoolVar(&o.AllowHTTPRegistry, "allow-http-registry", false,
		"whether to allow using HTTP protocol while connecting to registries. Don't use this for anything but testing")

	cmd.Flags().BoolVar(&o.KubernetesKeychain, "k8s-keychain", false,
		"whether to use the kubernetes keychain instead of the default keychain (supports workload identity).")

	o.RefOpts.AddFlags(cmd)
}

func (o *RegistryOptions) ClientOpts(ctx context.Context) ([]ociremote.Option, error) {
	opts := []ociremote.Option{ociremote.WithRemoteOptions(o.GetRegistryClientOpts(ctx)...)}
	if o.RefOpts.TagPrefix != "" {
		opts = append(opts, ociremote.WithPrefix(o.RefOpts.TagPrefix))
	}
	targetRepoOverride, err := ociremote.GetEnvTargetRepository()
	if err != nil {
		return nil, err
	}
	if (targetRepoOverride != name.Repository{}) {
		opts = append(opts, ociremote.WithTargetRepository(targetRepoOverride))
	}
	return opts, nil
}

func (o *RegistryOptions) NameOptions() []name.Option {
	var nameOpts []name.Option
	if o.AllowHTTPRegistry {
		nameOpts = append(nameOpts, name.Insecure)
	}
	return nameOpts
}

func (o *RegistryOptions) GetRegistryClientOpts(ctx context.Context) []remote.Option {
	opts := []remote.Option{
		remote.WithContext(ctx),
		remote.WithUserAgent(UserAgent()),
	}

	switch {
	case o.Keychain != nil:
		opts = append(opts, remote.WithAuthFromKeychain(o.Keychain))
	case o.KubernetesKeychain:
		kc := authn.NewMultiKeychain(
			authn.DefaultKeychain,
			google.Keychain,
			authn.NewKeychainFromHelper(ecr.NewECRHelper(ecr.WithLogger(io.Discard))),
			authn.NewKeychainFromHelper(credhelper.NewACRCredentialsHelper()),
			authn.NewKeychainFromHelper(alibabaacr.NewACRHelper().WithLoggerOut(io.Discard)),
			github.Keychain,
		)
		opts = append(opts, remote.WithAuthFromKeychain(kc))
	default:
		opts = append(opts, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	}

	if o.AllowInsecure {
		opts = append(opts, remote.WithTransport(&http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}})) // #nosec G402
	}
	return opts
}
