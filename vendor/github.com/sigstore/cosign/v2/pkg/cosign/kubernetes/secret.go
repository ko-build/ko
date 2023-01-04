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
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"k8s.io/utils/pointer"

	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sigstore/cosign/v2/pkg/cosign"
)

const (
	KeyReference = "k8s://"
)

func GetKeyPairSecret(ctx context.Context, k8sRef string) (*v1.Secret, error) {
	namespace, name, err := parseRef(k8sRef)
	if err != nil {
		return nil, err
	}

	client, err := client()
	if err != nil {
		return nil, fmt.Errorf("new for config: %w", err)
	}

	var s *v1.Secret
	if s, err = client.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{}); err != nil {
		return nil, fmt.Errorf("checking if secret exists: %w", err)
	}

	return s, nil
}

func KeyPairSecret(ctx context.Context, k8sRef string, pf cosign.PassFunc) error {
	namespace, name, err := parseRef(k8sRef)
	if err != nil {
		return err
	}
	// now, generate the key in memory
	keys, err := cosign.GenerateKeyPair(pf)
	if err != nil {
		return fmt.Errorf("generating key pair: %w", err)
	}

	// create the k8s client
	client, err := client()
	if err != nil {
		return fmt.Errorf("new for config: %w", err)
	}
	immutable, err := checkImmutableSecretSupported(client)
	if err != nil {
		return fmt.Errorf("check immutable: %w", err)
	}
	var s *v1.Secret
	if s, err = client.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{}); err != nil {
		if k8serrors.IsNotFound(err) {
			s, err = client.CoreV1().Secrets(namespace).Create(ctx, secret(keys, namespace, name, nil, immutable), metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("creating secret %s in ns %s: %w", name, namespace, err)
			}
		} else {
			return fmt.Errorf("checking if secret exists: %w", err)
		}
	} else { // Update the existing secret
		s, err = client.CoreV1().Secrets(namespace).Update(ctx, secret(keys, namespace, name, s.Data, immutable), metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("updating secret %s in ns %s: %w", name, namespace, err)
		}
	}

	fmt.Fprintf(os.Stderr, "Successfully created secret %s in namespace %s\n", s.Name, s.Namespace)
	if err := os.WriteFile("cosign.pub", keys.PublicBytes, 0600); err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "Public key written to cosign.pub")
	return nil
}

// creates a secret with the following data:
// * cosign.key
// * cosign.pub
// * cosign.password
func secret(keys *cosign.KeysBytes, namespace, name string, data map[string][]byte, immutable bool) *v1.Secret {
	if data == nil {
		data = map[string][]byte{}
	}
	data["cosign.key"] = keys.PrivateBytes
	data["cosign.pub"] = keys.PublicBytes
	data["cosign.password"] = keys.Password()

	obj := metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
	}

	// For Kubernetes >= 1.21, set Immutable by default
	if immutable {
		return &v1.Secret{
			ObjectMeta: obj,
			Data:       data,
			Immutable:  pointer.Bool(true),
		}
	}

	return &v1.Secret{
		ObjectMeta: obj,
		Data:       data,
	}
}

// the reference should be formatted as <namespace>/<secret name>
func parseRef(k8sRef string) (string, string, error) {
	s := strings.Split(strings.TrimPrefix(k8sRef, KeyReference), "/")
	if len(s) != 2 {
		return "", "", errors.New("kubernetes specification should be in the format k8s://<namespace>/<secret>")
	}
	return s[0], s[1], nil
}
