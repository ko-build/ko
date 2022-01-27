// This file is part of CycloneDX GoMod
//
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
// Copyright (c) OWASP Foundation. All Rights Reserved.

// Package generate exposes cyclonedx-gomod's SBOM generation capabilities.
package generate

import cdx "github.com/CycloneDX/cyclonedx-go"

// Generator is the interface that provides abstraction for multiple BOM generation strategies.
//
// Generators MUST only consider facts. Results MUST be reproducible.
// The returned BOM MUST NOT include any of the following elements:
//  - SerialNumber
//  - Metadata.Timestamp
//  - Metadata.Tools
//  - Metadata.Authors
//  - Metadata.Manufacture
//  - Metadata.Supplier
//  - Metadata.Licenses
// It is the responsibility of the caller to ensure that these elements
// are set according to their context and use case.
//
// At the very least, callers SHOULD set the SerialNumber and Metadata.Timestamp
// elements before publishing or forwarding the generated BOM.
type Generator interface {
	Generate() (*cdx.BOM, error)
}
