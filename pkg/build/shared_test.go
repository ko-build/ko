// Copyright 2018 ko Build Authors All Rights Reserved.
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
	"context"
	"testing"
	"time"

	"github.com/google/go-containerregistry/pkg/v1/random"
)

type slowbuild struct {
	sleep time.Duration
}

// slowbuild implements Interface
var _ Interface = (*slowbuild)(nil)

func (sb *slowbuild) QualifyImport(ip string) (string, error) { return ip, nil }

func (sb *slowbuild) IsSupportedReference(string) error { return nil }

func (sb *slowbuild) Build(context.Context, string) (Result, error) {
	time.Sleep(sb.sleep)
	return random.Index(256, 8, 3)
}

func TestCaching(t *testing.T) {
	duration := 100 * time.Millisecond
	ip := "foo"

	sb := &slowbuild{duration}
	cb, _ := NewCaching(sb)

	if err := cb.IsSupportedReference(ip); err != nil {
		t.Errorf("ISR(%q) = (%v), wanted nil", err, ip)
	}

	previousDigest := "not-a-digest"
	// Each iteration, we test that the first build is slow and subsequent
	// builds are fast and return the same image.  Then we invalidate the
	// cache and iterate.
	for idx := 0; idx < 3; idx++ {
		start := time.Now()
		img1, err := cb.Build(context.Background(), ip)
		if err != nil {
			t.Errorf("Build() = %v", err)
		}
		end := time.Now()

		elapsed := end.Sub(start)
		if elapsed < duration {
			t.Errorf("Elapsed time %v, wanted >= %s", elapsed, duration)
		}
		d1 := digest(t, img1)

		if d1 == previousDigest {
			t.Errorf("Got same digest as previous iteration, wanted different: %v", d1)
		}
		previousDigest = d1

		start = time.Now()
		img2, err := cb.Build(context.Background(), ip)
		if err != nil {
			t.Errorf("Build() = %v", err)
		}
		end = time.Now()

		elapsed = end.Sub(start)
		if elapsed >= duration {
			t.Errorf("Elapsed time %v, wanted < %s", elapsed, duration)
		}
		d2 := digest(t, img2)

		if d1 != d2 {
			t.Error("Got different images, wanted same")
		}

		cb.Invalidate(ip)
	}
}
