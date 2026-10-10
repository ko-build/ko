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

package commands

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRunInterruptibleSendsInterrupt(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("os.Interrupt is SIGINT on unix; Windows process signaling is different")
	}
	if os.Getenv("KO_TEST_INTERRUPT_HELPER") == "1" {
		interruptHelper()
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := exec.Command(os.Args[0], "-test.run=^TestRunInterruptibleSendsInterrupt$")
	cmd.Env = append(os.Environ(), "KO_TEST_INTERRUPT_HELPER=1")
	out := &syncBuffer{}
	cmd.Stdout = out
	cmd.Stderr = out

	errCh := make(chan error, 1)
	go func() {
		errCh <- runInterruptible(ctx, cmd)
	}()

	waitForOutput(t, out, "ready", 5*time.Second)
	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("runInterruptible() = %v, want nil after SIGINT", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for helper after interrupt; child was likely SIGKILLed")
	}

	if !strings.Contains(out.String(), "got interrupt") {
		t.Fatalf("helper did not receive SIGINT, output:\n%s", out.String())
	}
}

func interruptHelper() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	fmt.Println("ready")
	os.Stdout.Sync()
	<-ch
	fmt.Println("got interrupt")
}

type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func waitForOutput(t *testing.T, out *syncBuffer, want string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if strings.Contains(out.String(), want) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %q, output:\n%s", want, out.String())
}
