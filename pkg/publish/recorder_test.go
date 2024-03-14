// Copyright 2021 ko Build Authors All Rights Reserved.
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

package publish

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/ko/pkg/build"
	"github.com/sigstore/cosign/v2/pkg/oci/signed"
)

type cbPublish struct {
	cb func(context.Context, build.Result, string) (name.Reference, error)
}

var _ Interface = (*cbPublish)(nil)

func (sp *cbPublish) Publish(ctx context.Context, br build.Result, ref string) (name.Reference, error) {
	return sp.cb(ctx, br, ref)
}

func (sp *cbPublish) Close() error {
	return nil
}

func TestRecorder(t *testing.T) {
	repo := name.MustParseReference("docker.io/ubuntu:latest")
	inner := &cbPublish{cb: func(_ context.Context, b build.Result, _ string) (name.Reference, error) {
		h, err := b.Digest()
		if err != nil {
			return nil, err
		}
		return repo.Context().Digest(h.String()), nil
	}}

	buf := bytes.NewBuffer(nil)

	recorder, err := NewRecorder(inner, buf)
	if err != nil {
		t.Fatalf("NewRecorder() = %v", err)
	}

	img, err := random.Image(3, 3)
	if err != nil {
		t.Fatalf("random.Image() = %v", err)
	}
	si := signed.Image(img)

	index, err := random.Index(3, 3, 2)
	if err != nil {
		t.Fatalf("random.Image() = %v", err)
	}
	sii := signed.ImageIndex(index)

	if _, err := recorder.Publish(context.Background(), si, ""); err != nil {
		t.Errorf("recorder.Publish() = %v", err)
	}
	if _, err := recorder.Publish(context.Background(), si, ""); err != nil {
		t.Errorf("recorder.Publish() = %v", err)
	}
	if _, err := recorder.Publish(context.Background(), sii, ""); err != nil {
		t.Errorf("recorder.Publish() = %v", err)
	}
	if err := recorder.Close(); err != nil {
		t.Errorf("recorder.Close() = %v", err)
	}

	refs := strings.Split(strings.TrimSpace(buf.String()), "\n")

	if want, got := len(refs), 5; got != want {
		t.Errorf("len(refs) = %d, wanted %d", got, want)
	}

	for _, s := range refs {
		ref, err := name.ParseReference(s)
		if err != nil {
			t.Fatalf("name.ParseReference() = %v", err)
		}
		// Don't compare the digests, they are random.
		if want, got := repo.Context().String(), ref.Context().String(); want != got {
			t.Errorf("reference repo = %v, wanted %v", got, want)
		}
	}
}
