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

package kind

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"sigs.k8s.io/kind/pkg/cluster/nodes"
	"sigs.k8s.io/kind/pkg/exec"
)

func TestWrite(t *testing.T) {
	ctx := context.Background()
	img, err := random.Image(1024, 1)
	if err != nil {
		t.Fatalf("random.Image() = %v", err)
	}

	tag, err := name.NewTag("kind.local/test:new")
	if err != nil {
		t.Fatalf("name.NewTag() = %v", err)
	}

	n1 := &fakeNode{}
	n2 := &fakeNode{}
	GetProvider = func() provider {
		return &fakeProvider{nodes: []nodes.Node{n1, n2}}
	}

	if err := Write(ctx, tag, img); err != nil {
		t.Fatalf("Write() = %v", err)
	}

	// Verify the respective command is executed on each node.
	for _, n := range []*fakeNode{n1, n2} {
		if got, want := len(n.cmds), 1; got != want {
			t.Fatalf("len(n.cmds) = %d, want %d", got, want)
		}
		c := n.cmds[0]

		if got, want := c.cmd, "ctr --namespace=k8s.io images import --all-platforms -"; got != want {
			t.Fatalf("c.cmd = %s, want %s", got, want)
		}
	}
}

func TestTag(t *testing.T) {
	ctx := context.Background()
	oldTag, err := name.NewTag("kind.local/test:test")
	if err != nil {
		t.Fatalf("name.NewTag() = %v", err)
	}

	newTag, err := name.NewTag("kind.local/test:new")
	if err != nil {
		t.Fatalf("name.NewTag() = %v", err)
	}

	n1 := &fakeNode{}
	n2 := &fakeNode{}
	GetProvider = func() provider {
		return &fakeProvider{nodes: []nodes.Node{n1, n2}}
	}

	if err := Tag(ctx, oldTag, newTag); err != nil {
		t.Fatalf("Tag() = %v", err)
	}

	// Verify the respective command is executed on each node.
	for _, n := range []*fakeNode{n1, n2} {
		if got, want := len(n.cmds), 1; got != want {
			t.Fatalf("len(n.cmds) = %d, want %d", got, want)
		}
		c := n.cmds[0]

		if got, want := c.cmd, fmt.Sprintf("ctr --namespace=k8s.io images tag --force %s %s", oldTag, newTag); got != want {
			t.Fatalf("c.cmd = %s, want %s", got, want)
		}
	}
}

func TestFailWithNoNodes(t *testing.T) {
	ctx := context.Background()
	img, err := random.Image(1024, 1)
	if err != nil {
		panic(err)
	}

	oldTag, err := name.NewTag("kind.local/test:test")
	if err != nil {
		t.Fatalf("name.NewTag() = %v", err)
	}

	newTag, err := name.NewTag("kind.local/test:new")
	if err != nil {
		t.Fatalf("name.NewTag() = %v", err)
	}

	GetProvider = func() provider {
		return &fakeProvider{}
	}

	if err := Write(ctx, newTag, img); err == nil {
		t.Fatal("Write() = nil, wanted an error")
	}
	if err := Tag(ctx, oldTag, newTag); err == nil {
		t.Fatal("Tag() = nil, wanted an error")
	}
}

func TestFailCommands(t *testing.T) {
	ctx := context.Background()
	img, err := random.Image(1024, 1)
	if err != nil {
		panic(err)
	}

	oldTag, err := name.NewTag("kind.local/test:test")
	if err != nil {
		t.Fatalf("name.NewTag() = %v", err)
	}

	newTag, err := name.NewTag("kind.local/test:new")
	if err != nil {
		t.Fatalf("name.NewTag() = %v", err)
	}

	errTest := errors.New("test")

	n1 := &fakeNode{err: errTest}
	n2 := &fakeNode{err: errTest}
	GetProvider = func() provider {
		return &fakeProvider{nodes: []nodes.Node{n1, n2}}
	}

	if err := Write(ctx, newTag, img); !errors.Is(err, errTest) {
		t.Fatalf("Write() = %v, want %v", err, errTest)
	}
	if err := Tag(ctx, oldTag, newTag); !errors.Is(err, errTest) {
		t.Fatalf("Write() = %v, want %v", err, errTest)
	}
}

// fakeProvider
type fakeProvider struct {
	nodes []nodes.Node
}

func (f *fakeProvider) ListInternalNodes(string) ([]nodes.Node, error) {
	return f.nodes, nil
}

type fakeNode struct {
	cmds []*fakeCmd
	err  error
}

func (f *fakeNode) CommandContext(_ context.Context, cmd string, args ...string) exec.Cmd {
	command := &fakeCmd{
		cmd: strings.Join(append([]string{cmd}, args...), " "),
		err: f.err,
	}
	f.cmds = append(f.cmds, command)
	return command
}

func (f *fakeNode) String() string {
	return "test"
}

// The following functions are not used by our code at all.
func (f *fakeNode) Command(string, ...string) exec.Cmd { return nil }
func (f *fakeNode) Role() (string, error)              { return "", nil }
func (f *fakeNode) IP() (string, string, error)        { return "", "", nil }
func (f *fakeNode) SerialLogs(io.Writer) error         { return nil }

type fakeCmd struct {
	cmd   string
	err   error
	stdin io.Reader
}

func (f *fakeCmd) Run() error {
	if f.stdin != nil {
		// Consume the entire stdin to move the image publish forward.
		io.ReadAll(f.stdin)
	}
	return f.err
}

func (f *fakeCmd) SetStdin(stdin io.Reader) exec.Cmd {
	f.stdin = stdin
	return f
}

// The following functions are not used by our code at all.
func (f *fakeCmd) SetEnv(...string) exec.Cmd    { return f }
func (f *fakeCmd) SetStdout(io.Writer) exec.Cmd { return f }
func (f *fakeCmd) SetStderr(io.Writer) exec.Cmd { return f }
