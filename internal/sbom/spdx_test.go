// Copyright 2026 ko Build Authors All Rights Reserved.
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

package sbom

import (
	"encoding/json"
	"regexp"
	"testing"

	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/sigstore/cosign/v3/pkg/oci/signed"
)

func TestGenerateImageSPDXModuleIDs(t *testing.T) {
	modules := "app: go1.27.0\n" +
		"\tpath\texample.com/app\n" +
		"\tmod\texample.com/app\t(devel)\t\n" +
		"\tdep\texample.com/lib_name\tv1.2.3\t\n" +
		"\tdep\texample.com/lib-name\tv1.2.3\t\n" +
		"\tdep\texample.com/legacy\tv2.0.0+incompatible\t\n"
	data, err := GenerateImageSPDX("test", []byte(modules), signed.Image(empty.Image))
	if err != nil {
		t.Fatal(err)
	}

	var doc Document
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	idPattern := regexp.MustCompile(`^SPDXRef-[a-zA-Z0-9.-]+$`)
	ids := map[string]bool{doc.ID: true}
	packages := make(map[string]Package)
	for _, pkg := range doc.Packages {
		if !idPattern.MatchString(pkg.ID) {
			t.Errorf("invalid SPDX identifier %q", pkg.ID)
		}
		if ids[pkg.ID] {
			t.Errorf("duplicate SPDX identifier %q", pkg.ID)
		}
		ids[pkg.ID] = true
		packages[pkg.Name] = pkg
	}
	if got := packages["example.com/lib-name"].ID; got != "SPDXRef-Package-example.com.lib-name-v1.2.3" {
		t.Errorf("valid module identifier changed: %q", got)
	}
	if got := packages["example.com/legacy"].Version; got != "v2.0.0+incompatible" {
		t.Errorf("dependency version changed: %q", got)
	}
	for _, rel := range doc.Relationships {
		if !ids[rel.Element] || !ids[rel.Related] {
			t.Errorf("relationship references an unknown SPDX identifier: %+v", rel)
		}
	}
}
