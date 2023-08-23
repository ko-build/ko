// Copyright 2020 ko Build Authors All Rights Reserved.
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
	"context"
	"os"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/v1/random"
)

func TestLayout(t *testing.T) {
	img, err := random.Image(1024, 1)
	if err != nil {
		t.Fatalf("random.Image() = %v", err)
	}
	importpath := "github.com/Google/go-containerregistry/cmd/crane"

	tmp, err := os.MkdirTemp("/tmp", "ko")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	lp, err := NewLayout(tmp)
	if err != nil {
		t.Errorf("NewLayout() = %v", err)
	}
	if d, err := lp.Publish(context.Background(), img, importpath); err != nil {
		t.Errorf("Publish() = %v", err)
	} else if !strings.HasPrefix(d.String(), tmp) {
		t.Errorf("Publish() = %v, wanted prefix %v", d, tmp)
	}
}
