// Copyright 2021 Google LLC All Rights Reserved.
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
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/ko/pkg/build"
	"k8s.io/apimachinery/pkg/util/rand"
)

const (
	// OpenShiftDomain is a sentinel "registry" that represents loading images into
	// OpenShift's internal registry.
	OpenShiftDomain = "ocp.local"
)

type ocpPublisher struct {
	inner  Interface
	stdin  io.Closer
	tunnel *os.Process
}

// NewOpenShiftPublisher returns a new publish.Interface that loads images into
// OpenShift's internal registry.
func NewOpenShiftPublisher(namer Namer, tags []string) (Interface, error) {
	// Check if oc is installed.
	if _, err := exec.LookPath("oc"); err != nil {
		return nil, fmt.Errorf("failed to find oc, is it installed? : %w", err)
	}

	// Login to the registry.
	if _, err := runOc("registry", "login"); err != nil {
		return nil, fmt.Errorf("failed to login to the registry: %w", err)
	}
	log.Print("Logged into Openshift registry")

	registryHostPort, err := runOc("registry", "info", "--internal")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch internal registry host: %w", err)
	}

	// Setup a tunnel to the registry.
	// Note: port-forward is a privileged operation on Openshift, so we build an pipe
	//       tunnel using a socat pod and attaching to its pipes.
	//nolint:gosec // Launching this command with the output of the registry call above.
	tunnel := exec.Command("oc", "run", "registry-tunnel-"+rand.String(10), "--rm", "-i", "--image", "alpine/socat", "--", "-", "TCP4:"+registryHostPort)
	in, err := tunnel.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get tunnel stdin: %w", err)
	}
	out, err := tunnel.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get tunnel stdin: %w", err)
	}
	if err := tunnel.Start(); err != nil {
		return nil, fmt.Errorf("failed to launch tunnel: %w", err)
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		// Force connect to the local tunnel.
		// Note: This is only safe as long as the "connection" is always used sequentially.
		//       `ko` seems to behave this way currently. If this changes, we either have
		//       to launch a pod per connection or `oc exec` into a singular pod that just
		//       sleeps to create truly multiple connections.
		return pipeConn{in: in, out: out}, nil
	}

	// Determine the namespace to push the image to by assuming the user wants to push to
	// the active project.
	// Note: This is fine for single-namespace deployments but will fail if the manifest
	//       contains deployments in different namespaces.
	targetNamespace, err := runOc("project", "-q")
	if err != nil {
		return nil, fmt.Errorf("failed to determine active project: %w", err)
	}
	log.Printf("Pushing images to namespace %q", targetNamespace)

	inner, err := NewDefault(fmt.Sprintf("%s/%s", registryHostPort, targetNamespace),
		WithTransport(transport),
		WithAuthFromKeychain(authn.DefaultKeychain),
		WithNamer(namer),
		WithTags(tags),
		Insecure(true))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize inner publisher: %w", err)
	}

	return &ocpPublisher{
		inner:  inner,
		stdin:  in,
		tunnel: tunnel.Process,
	}, nil
}

// Publish implements publish.Interface.
func (t *ocpPublisher) Publish(ctx context.Context, br build.Result, s string) (name.Reference, error) {
	return t.inner.Publish(ctx, br, s)
}

func (t *ocpPublisher) Close() error {
	if err := t.stdin.Close(); err != nil {
		return fmt.Errorf("failed to close the tunnel: %w", err)
	}
	if _, err := t.tunnel.Wait(); err != nil {
		return fmt.Errorf("tunnel process didn't exit cleanly: %w", err)
	}
	return t.inner.Close()
}

// runOc is a helper that runs the given command with `oc` and returns the result as a
// trimmed string.
func runOc(args ...string) (string, error) {
	raw, err := exec.Command("oc", args...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(raw)), nil
}

// pipeConn acts as a TCP connection over a given input and output stream, that's
// supposed to be STDIN and STDOUT of the remote `socat` process.
type pipeConn struct {
	net.Conn
	out io.Reader
	in  io.Writer
}

func (c pipeConn) Read(b []byte) (n int, err error) {
	return c.out.Read(b)
}

func (c pipeConn) Write(b []byte) (n int, err error) {
	return c.in.Write(b)
}

func (c pipeConn) Close() error {
	// We don't close STDIN here as we want to reuse the process among many connections.
	return nil
}

func (c pipeConn) SetDeadline(t time.Time) error {
	return nil
}

func (c pipeConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c pipeConn) SetWriteDeadline(t time.Time) error {
	return nil
}
