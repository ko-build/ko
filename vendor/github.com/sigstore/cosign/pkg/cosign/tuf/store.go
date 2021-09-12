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

package tuf

import (
	"context"
	"io"
	"path"

	"cloud.google.com/go/storage"
	"github.com/theupdateframework/go-tuf/client"
	"google.golang.org/api/option"
)

type GcsRemoteOptions struct {
	MetadataPath string
	TargetsPath  string
}

type gcsRemoteStore struct {
	bucket string
	ctx    context.Context
	client *storage.Client
	opts   *GcsRemoteOptions
}

// A remote store for TUF metadata on GCS.
func GcsRemoteStore(ctx context.Context, bucket string, opts *GcsRemoteOptions, client *storage.Client) (client.RemoteStore, error) {
	if opts == nil {
		opts = &GcsRemoteOptions{}
	}
	if opts.TargetsPath == "" {
		opts.TargetsPath = "targets"
	}
	store := gcsRemoteStore{ctx: ctx, bucket: bucket, opts: opts, client: client}
	if client == nil {
		var err error
		store.client, err = storage.NewClient(ctx, option.WithoutAuthentication())
		if err != nil {
			return nil, err
		}
	}
	return &store, nil
}

func (h *gcsRemoteStore) GetMeta(name string) (io.ReadCloser, int64, error) {
	return h.get(path.Join(h.opts.MetadataPath, name))
}

func (h *gcsRemoteStore) GetTarget(name string) (io.ReadCloser, int64, error) {
	return h.get(path.Join(h.opts.TargetsPath, name))
}

func (h *gcsRemoteStore) get(s string) (io.ReadCloser, int64, error) {
	obj := h.client.Bucket(h.bucket).Object(s)
	attrs, err := obj.Attrs(h.ctx)
	if err != nil {
		return nil, 0, err
	}
	rc, err := obj.NewReader(h.ctx)
	if err != nil {
		return nil, 0, err
	}
	return rc, attrs.Size, nil
}
