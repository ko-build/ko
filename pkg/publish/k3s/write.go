// Copyright 2020 Google LLC All Rights Reserved.
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

package k3s

import (
	"bytes"
	"context"
	"fmt"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"golang.org/x/sync/errgroup"
	"io"
	"log"
	"os"
	"os/exec"
)

const (
	limaInstanceEnvKey             = "LIMA_INSTANCE"
	rancherDesktopLimaInstanceName = "0"
)

// Tag adds a tag to an already existent image.
func Tag(ctx context.Context, src, dest name.Tag) error {
	nerdctl, li, env := commandWithEnv()

	var stdErr bytes.Buffer
	cmd := exec.CommandContext(ctx, nerdctl, "--namespace=k8s.io", "tag", src.String(), dest.String())
	cmd.Env = env
	cmd.Stderr = &stdErr

	if err := cmd.Run(); err != nil {
		log.Printf("Error while excuting command %s %s", cmd.String(), stdErr.String())
		return fmt.Errorf("failed to tag image to instance %q: %w", li, err)
	}

	return nil
}

// Write saves the image into the k3s nodes as the given tag.
func Write(ctx context.Context, tag name.Tag, img v1.Image) error {
	pr, pw := io.Pipe()

	grp := errgroup.Group{}
	grp.Go(func() error {
		return pw.CloseWithError(tarball.Write(tag, img, pw))
	})

	nerdctl, li, env := commandWithEnv()

	var stdErr bytes.Buffer
	cmd := exec.CommandContext(ctx, nerdctl, "--namespace=k8s.io", "load")
	cmd.Stdin = pr
	cmd.Env = env
	cmd.Stderr = &stdErr

	if err := cmd.Run(); err != nil {
		log.Printf("Error while excuting command %s %s", cmd.String(), stdErr.String())
		return fmt.Errorf("failed to load image to instance %q: %w", li, err)
	}

	if err := grp.Wait(); err != nil {
		return fmt.Errorf("failed to write intermediate tarball representation: %w", err)
	}

	return nil
}

//commandWithEnv build the nerdctl command with correct instance name and required environment variables
func commandWithEnv() (string, string, []string) {
	nerdctl := "nerdctl.lima"
	env := os.Environ()
	// If no LIMA_INSTANCE env is defined it defaults to Rancher Desktop "0"
	li, ok := os.LookupEnv(limaInstanceEnvKey)
	if !ok {
		nerdctl = "nerdctl"
		li = rancherDesktopLimaInstanceName
		env = append(env,
			fmt.Sprintf("LIMA_INSTANCE=%s", li))
	}

	return nerdctl, li, env
}
