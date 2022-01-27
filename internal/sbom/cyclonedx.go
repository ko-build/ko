// Copyright 2022 Google LLC All Rights Reserved.
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
	"bytes"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/CycloneDX/cyclonedx-gomod/pkg/generate/bin"
	"github.com/rs/zerolog"
)

func GenerateCycloneDX(binaryPath string) ([]byte, error) {
	gen, err := bin.NewGenerator(binaryPath, bin.WithLogger(zerolog.Nop()))
	if err != nil {
		return nil, err
	}
	bom, err := gen.Generate()
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	enc := cdx.NewBOMEncoder(&buf, cdx.BOMFileFormatJSON)
	enc.SetPretty(true)
	if err := enc.Encode(bom); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
