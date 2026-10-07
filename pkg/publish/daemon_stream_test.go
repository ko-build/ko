// Copyright 2023 ko Build Authors All Rights Reserved.
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
	archivetar "archive/tar"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

func TestDockerStreamError(t *testing.T) {
	resp := `{"stream":"Loading layer\n"}{"errorDetail":{"message":"cannot load windows image"},"error":"cannot load windows image"}`
	err := dockerStreamError(resp)
	if err == nil || err.Error() != "cannot load windows image" {
		t.Fatalf("dockerStreamError() = %v", err)
	}
	if err := dockerStreamError(""); err != nil {
		t.Fatalf("empty stream error = %v", err)
	}
	if err := dockerStreamError("Loaded"); err != nil {
		t.Fatalf("non-json stream error = %v", err)
	}
}

func TestDaemonLoadDroppedTag(t *testing.T) {
	const tag = "ko.local/github.com/google/ko:abc"
	if daemonLoadDroppedTag("", tag) {
		t.Fatal("empty response is the already-tagged fast path")
	}
	kept := `{"stream":"Loaded image: ` + tag + `\n"}`
	if daemonLoadDroppedTag(kept, tag) {
		t.Fatal("response names the tag")
	}
	dropped := `{"stream":"Loaded image ID: sha256:abc\n"}`
	if !daemonLoadDroppedTag(dropped, tag) {
		t.Fatal("ID-only load dropped the tag")
	}
}

func TestDaemonTagStatusRetryStillDropped(t *testing.T) {
	const tag = "ko.local/github.com/google/ko:abc"
	dropped := `{"stream":"Loaded image ID: sha256:abc\n"}`
	applied, err := daemonTagStatus(dropped, tag)
	if err != nil {
		t.Fatal(err)
	}
	if applied {
		t.Fatal("ID-only load must not count as tagged")
	}
	applied, err = daemonTagStatus("", tag)
	if err != nil || !applied {
		t.Fatalf("fast path applied=%v err=%v", applied, err)
	}
}

func TestSanitizeDockerArchiveStripsConfigColon(t *testing.T) {
	img, err := random.Image(1024, 1)
	if err != nil {
		t.Fatal(err)
	}
	tag, err := name.NewTag("ko.local/test:latest")
	if err != nil {
		t.Fatal(err)
	}
	var raw bytes.Buffer
	if err := tarball.Write(tag, img, &raw); err != nil {
		t.Fatal(err)
	}
	var sanitized bytes.Buffer
	if err := rewriteDockerArchive(bytes.NewReader(raw.Bytes()), &sanitized); err != nil {
		t.Fatal(err)
	}

	tr := archivetar.NewReader(bytes.NewReader(sanitized.Bytes()))
	var sawManifest bool
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(hdr.Name, ":") {
			t.Fatalf("tar entry %q contains a colon", hdr.Name)
		}
		if hdr.Name != "manifest.json" {
			continue
		}
		body, err := io.ReadAll(tr)
		if err != nil {
			t.Fatal(err)
		}
		sawManifest = true
		var manifests []struct {
			Config string
			Layers []string
		}
		if err := json.Unmarshal(body, &manifests); err != nil {
			t.Fatal(err)
		}
		if len(manifests) != 1 {
			t.Fatalf("manifests = %#v", manifests)
		}
		if strings.Contains(manifests[0].Config, ":") {
			t.Fatalf("config name %q still has a colon", manifests[0].Config)
		}
		cfgName, err := img.ConfigName()
		if err != nil {
			t.Fatal(err)
		}
		if manifests[0].Config != cfgName.Hex {
			t.Fatalf("config name = %q, want %q", manifests[0].Config, cfgName.Hex)
		}
	}
	if !sawManifest {
		t.Fatal("sanitized archive has no manifest.json")
	}
}
