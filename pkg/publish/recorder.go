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
	"context"
	"io"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/ko/pkg/build"
)

// recorder wraps a publisher implementation in a layer that recordes the published
// references to an io.Writer.
type recorder struct {
	inner Interface
	wc    io.Writer
}

// recorder implements Interface
var _ Interface = (*recorder)(nil)

// NewRecorder wraps the provided publish.Interface in an implementation that
// records publish results to an io.Writer.
func NewRecorder(inner Interface, wc io.Writer) (Interface, error) {
	return &recorder{
		inner: inner,
		wc:    wc,
	}, nil
}

// Publish implements Interface
func (r *recorder) Publish(ctx context.Context, br build.Result, ref string) (name.Reference, error) {
	result, err := r.inner.Publish(ctx, br, ref)
	if err != nil {
		return nil, err
	}
	if _, err := r.wc.Write([]byte(result.String() + "\n")); err != nil {
		return nil, err
	}
	return result, nil
}

// Close implements Interface
func (r *recorder) Close() error {
	if err := r.inner.Close(); err != nil {
		return err
	}
	if c, ok := r.wc.(io.Closer); ok {
		return c.Close()
	}
	return nil
}
