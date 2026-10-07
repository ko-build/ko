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

import "testing"

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
