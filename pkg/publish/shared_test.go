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

package publish

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/ko/pkg/build"
)

type slowpublish struct {
	sleep time.Duration
}

// slowpublish implements Interface
var _ Interface = (*slowpublish)(nil)

func (sp *slowpublish) Publish(context.Context, build.Result, string) (name.Reference, error) {
	time.Sleep(sp.sleep)
	return makeRef()
}

func (sp *slowpublish) Close() error {
	return nil
}

func TestCaching(t *testing.T) {
	duration := 100 * time.Millisecond
	ref := "foo"

	sp := &slowpublish{duration}
	cb, _ := NewCaching(sp)

	previousDigest := "not-a-digest"
	// Each iteration, we test that the first publish is slow and subsequent
	// publishs are fast and return the same reference.  For each of these
	// iterations we use a new random image, which should invalidate the
	// cached reference from previous iterations.
	for range 3 {
		img, _ := random.Index(256, 8, 1)

		start := time.Now()
		ref1, err := cb.Publish(context.Background(), img, ref)
		if err != nil {
			t.Errorf("Publish() = %v", err)
		}
		end := time.Now()

		elapsed := end.Sub(start)
		if elapsed < duration {
			t.Errorf("Elapsed time %v, wanted >= %s", elapsed, duration)
		}
		d1 := ref1.String()

		if d1 == previousDigest {
			t.Errorf("Got same digest as previous iteration, wanted different: %v", d1)
		}
		previousDigest = d1

		start = time.Now()
		ref2, err := cb.Publish(context.Background(), img, ref)
		if err != nil {
			t.Errorf("Publish() = %v", err)
		}
		end = time.Now()

		elapsed = end.Sub(start)
		if elapsed >= duration {
			t.Errorf("Elapsed time %v, wanted < %s", elapsed, duration)
		}
		d2 := ref2.String()

		if d1 != d2 {
			t.Error("Got different references, wanted same")
		}
	}
}

// ctxpublish fails with the context's error if it is done, and succeeds otherwise.
type ctxpublish struct {
	calls atomic.Int32
}

// ctxpublish implements Interface
var _ Interface = (*ctxpublish)(nil)

func (cp *ctxpublish) Publish(ctx context.Context, _ build.Result, _ string) (name.Reference, error) {
	cp.calls.Add(1)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return makeRef()
}

func (cp *ctxpublish) Close() error {
	return nil
}

func TestCachingRetriesOtherCallersContextError(t *testing.T) {
	ref := "foo"
	inner := &ctxpublish{}
	cp, _ := NewCaching(inner)
	img, _ := random.Index(256, 1, 1)

	// The first caller's context is cancelled, so its publish fails.
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := cp.Publish(cancelled, img, ref); !errors.Is(err, context.Canceled) {
		t.Fatalf("Publish(cancelled) = %v, wanted %v", err, context.Canceled)
	}

	// A caller with a live context must not get the cached cancellation.
	ref1, err := cp.Publish(context.Background(), img, ref)
	if err != nil {
		t.Fatalf("Publish() = %v", err)
	}

	// And the successful result is cached from then on.
	ref2, err := cp.Publish(context.Background(), img, ref)
	if err != nil {
		t.Fatalf("Publish() = %v", err)
	}
	if ref1 != ref2 {
		t.Errorf("Publish() = %v, wanted the cached %v", ref2, ref1)
	}
	if got, want := inner.calls.Load(), int32(2); got != want {
		t.Errorf("inner publishes = %d, wanted %d", got, want)
	}
}
