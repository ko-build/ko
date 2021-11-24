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
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/ko/pkg/build"
)

const (
	// OpenShiftDomain is a sentinel "registry" that represents loading images into
	// OpenShift's internal registry.
	OpenShiftDomain = "ocp.local"

	// This line is printed by the tunnel command when it tells us which port it uses.
	portPrefix = "Forwarding from 127.0.0.1:"
)

type ocpPublisher struct {
	inner  Interface
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
	registryHost, registryPort, err := net.SplitHostPort(registryHostPort)
	if err != nil {
		return nil, fmt.Errorf("failed to split host and port of registry: %w", err)
	}
	registryHostParts := strings.SplitN(registryHost, ".", 3)
	registrySvc := registryHostParts[0]
	registryNs := registryHostParts[1]

	// Setup a tunnel to the registry.
	// TODO: Should we generalize this into a ko feature for a tunneled registry?
	tunnel := exec.Command("oc", "port-forward", "-n", registryNs, "svc/"+registrySvc, ":"+registryPort)
	tunnel.Stderr = os.Stderr
	tunnelLogs, err := tunnel.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get tunnel output: %w", err)
	}
	if err := tunnel.Start(); err != nil {
		return nil, fmt.Errorf("failed to launch tunnel: %w", err)
	}

	// Wait for the "Forwarding from" logline to appear.
	var tunnelPort string
	scanner := bufio.NewScanner(tunnelLogs)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, portPrefix) {
			tunnelPort = strings.TrimPrefix(strings.TrimSuffix(line, " -> 5000"), portPrefix)
			break
		}
	}
	// Drop all remaining stdout logs into a black hole so that the process can continue.
	go io.Copy(io.Discard, tunnelLogs)
	log.Printf("Using local tunnel with port %q", tunnelPort)

	transport := http.DefaultTransport.(*http.Transport).Clone()
	dial := transport.DialContext
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		// Force connect to the local tunnel.
		// TODO: Should we generalize this into the default handler so people can use
		//       existing tunnels in the default publisher?
		return dial(ctx, "tcp", "127.0.0.1:"+tunnelPort)
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

	inner, err := NewDefault(fmt.Sprintf("%s.%s.svc:%s/%s", registrySvc, registryNs, registryPort, targetNamespace),
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
		tunnel: tunnel.Process,
	}, nil
}

// Publish implements publish.Interface.
func (t *ocpPublisher) Publish(ctx context.Context, br build.Result, s string) (name.Reference, error) {
	return t.inner.Publish(ctx, br, s)
}

func (t *ocpPublisher) Close() error {
	if err := t.tunnel.Kill(); err != nil {
		return fmt.Errorf("failed to kill tunnel process: %w", err)
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
