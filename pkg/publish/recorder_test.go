// Copyright 2021 Google LLC All Rights Reserved.
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
	"fmt"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/ko/pkg/build"
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
	num := 0
	inner := &cbPublish{cb: func(c context.Context, b build.Result, s string) (name.Reference, error) {
		num++
		return name.ParseReference(fmt.Sprintf("ubuntu:%d", num))
	}}

	buf := bytes.NewBuffer(nil)

	recorder, err := NewRecorder(inner, buf)
	if err != nil {
		t.Fatalf("NewRecorder() = %v", err)
	}

	if _, err := recorder.Publish(context.Background(), nil, ""); err != nil {
		t.Errorf("recorder.Publish() = %v", err)
	}
	if _, err := recorder.Publish(context.Background(), nil, ""); err != nil {
		t.Errorf("recorder.Publish() = %v", err)
	}
	if _, err := recorder.Publish(context.Background(), nil, ""); err != nil {
		t.Errorf("recorder.Publish() = %v", err)
	}
	if err := recorder.Close(); err != nil {
		t.Errorf("recorder.Close() = %v", err)
	}

	want, got := "ubuntu:1\nubuntu:2\nubuntu:3\n", buf.String()
	if got != want {
		t.Errorf("buf.String() = %s, wanted %s", got, want)
	}
}
