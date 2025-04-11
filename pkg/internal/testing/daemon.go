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

package testing

import (
	"context"
	"io"
	"strings"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
)

type MockDaemon struct {
	daemon.Client
	Tags []string

	inspectErr  error
	inspectResp image.InspectResponse
	inspectBody []byte
}

func (m *MockDaemon) NegotiateAPIVersion(context.Context) {}
func (m *MockDaemon) ImageLoad(context.Context, io.Reader, ...client.ImageLoadOption) (image.LoadResponse, error) {
	return image.LoadResponse{
		Body: io.NopCloser(strings.NewReader("Loaded")),
	}, nil
}

func (m *MockDaemon) ImageTag(_ context.Context, _ string, tag string) error {
	if m.Tags == nil {
		m.Tags = []string{}
	}
	m.Tags = append(m.Tags, tag)
	return nil
}

func (m *MockDaemon) ImageInspectWithRaw(_ context.Context, _ string) (image.InspectResponse, []byte, error) {
	return m.inspectResp, m.inspectBody, m.inspectErr
}
