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

	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/moby/moby/client"
)

type MockDaemon struct {
	daemon.Client
	Tags []string

	inspectErr  error
	inspectResp client.ImageInspectResult
}

func (m *MockDaemon) Ping(context.Context, client.PingOptions) (client.PingResult, error) {
	return client.PingResult{}, nil
}
func (m *MockDaemon) ImageLoad(context.Context, io.Reader, ...client.ImageLoadOption) (client.ImageLoadResult, error) {
	return io.NopCloser(strings.NewReader("Loaded")), nil
}

func (m *MockDaemon) ImageTag(_ context.Context, opt client.ImageTagOptions) (client.ImageTagResult, error) {
	if m.Tags == nil {
		m.Tags = []string{}
	}
	m.Tags = append(m.Tags, opt.Target)
	return client.ImageTagResult{}, nil
}

func (m *MockDaemon) ImageInspect(context.Context, string, ...client.ImageInspectOption) (client.ImageInspectResult, error) {
	return m.inspectResp, m.inspectErr
}
