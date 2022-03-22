//go:build go1.18
// +build go1.18

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

package sbom

import (
	"fmt"
	"runtime/debug"
)

type BuildInfo debug.BuildInfo

func ParseBuildInfo(data string) (*BuildInfo, error) {
	dbi, err := debug.ParseBuildInfo(data)
	if err != nil {
		return nil, fmt.Errorf("parsing build info: %w", err)
	}
	bi := BuildInfo(*dbi)
	return &bi, nil
}
