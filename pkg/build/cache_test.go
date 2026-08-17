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

package build

import (
	"bytes"
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetBuildIDLogIncludesFilename(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing-binary")
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	_, err := getBuildID(context.Background(), missing)
	if err == nil {
		t.Fatal("getBuildID() = nil, want error for missing file")
	}
	got := buf.String()
	if !strings.Contains(got, missing) {
		t.Fatalf("log %q does not contain filename %q", got, missing)
	}
	// The swapped-arg bug put the exec error in the %s filename slot.
	if strings.Contains(got, `go tool buildid exec:`) || strings.Contains(got, `go tool buildid exit status`) {
		t.Fatalf("log put the error where the filename belongs: %q", got)
	}
}
