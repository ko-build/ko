// Copyright 2019 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package legacy provides methods for interacting with legacy image formats.
package legacy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/logs"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
)

// CopySchema1 allows `[g]crane cp` to work with old images without adding
// full support for schema 1 images to this package.
func CopySchema1(desc *remote.Descriptor, srcRef, dstRef name.Reference, srcAuth, dstAuth authn.Authenticator) error {
	m := schema1{}
	if err := json.NewDecoder(bytes.NewReader(desc.Manifest)).Decode(&m); err != nil {
		return err
	}

	for _, layer := range m.FSLayers {
		src := srcRef.Context().Digest(layer.BlobSum)
		dst := dstRef.Context().Digest(layer.BlobSum)

		blob, err := remote.Layer(src, remote.WithAuth(srcAuth))
		if err != nil {
			return err
		}

		if err := remote.WriteLayer(dst.Context(), blob, remote.WithAuth(dstAuth)); err != nil {
			return err
		}
	}

	return putManifest(desc, dstRef, dstAuth)
}

// TODO: perhaps expose this in remote?
func putManifest(desc *remote.Descriptor, dstRef name.Reference, dstAuth authn.Authenticator) error {
	reg := dstRef.Context().Registry
	scopes := []string{dstRef.Scope(transport.PushScope)}

	// TODO(jonjohnsonjr): Use NewWithContext.
	tr, err := transport.New(reg, dstAuth, http.DefaultTransport, scopes)
	if err != nil {
		return err
	}
	client := &http.Client{Transport: tr}

	u := url.URL{
		Scheme: dstRef.Context().Registry.Scheme(),
		Host:   dstRef.Context().RegistryStr(),
		Path:   fmt.Sprintf("/v2/%s/manifests/%s", dstRef.Context().RepositoryStr(), dstRef.Identifier()),
	}

	req, err := http.NewRequest(http.MethodPut, u.String(), bytes.NewBuffer(desc.Manifest))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", string(desc.MediaType))

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := transport.CheckError(resp, http.StatusOK, http.StatusCreated, http.StatusAccepted); err != nil {
		return err
	}

	// The image was successfully pushed!
	logs.Progress.Printf("%v: digest: %v size: %d", dstRef, desc.Digest, len(desc.Manifest))
	return nil
}

type fslayer struct {
	BlobSum string `json:"blobSum"`
}

type schema1 struct {
	FSLayers []fslayer `json:"fsLayers"`
}
