// Copyright 2026 ko Build Authors All Rights Reserved.
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
	"testing"
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

func TestStarterDocumentCreatedIsUTC(t *testing.T) {
	// SOURCE_DATE_EPOCH=0 observed from a UTC-8 time zone.
	date := time.Unix(0, 0).In(time.FixedZone("UTC-8", -8*60*60))

	doc, _ := starterDocument("test", date, v1.Hash{Algorithm: "sha256", Hex: "deadbeef"})

	if want, got := "1970-01-01T00:00:00Z", doc.CreationInfo.Created; got != want {
		t.Errorf("Created = %q, wanted %q", got, want)
	}
}
