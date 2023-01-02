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

package kubernetes

import (
	"fmt"

	"k8s.io/client-go/kubernetes"

	utilversion "k8s.io/apimachinery/pkg/util/version"
	// Initialize all known client auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func client() (kubernetes.Interface, error) {
	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(), nil).ClientConfig()
	if clientcmd.IsEmptyConfig(err) {
		cfg, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("error creating REST client config in-cluster: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("error creating REST client config: %w", err)
	}
	return kubernetes.NewForConfig(cfg)
}

func checkImmutableSecretSupported(client kubernetes.Interface) (bool, error) {
	k8sVer, err := client.Discovery().ServerVersion()
	if err != nil {
		return false, err
	}
	semVer, err := utilversion.ParseSemantic(k8sVer.String())
	if err != nil {
		return false, err
	}
	// https://kubernetes.io/docs/concepts/configuration/secret/#secret-immutable
	return semVer.Major() >= 1 && semVer.Minor() >= 21, nil
}
