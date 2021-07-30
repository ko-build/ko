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

package commands

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/spf13/cobra"
)

// addDeps augments our CLI surface with deps.
func addDeps(topLevel *cobra.Command) {
	var platform string

	deps := &cobra.Command{
		Use:   "deps IMAGE",
		Short: "Print Go module dependency information about the ko-built binary in the image",
		Long: `This sub-command finds and extracts the executable binary in the image, assuming it was built by ko, and prints information about the Go module dependencies of that executable, as reported by "go version -m".

If the image was not built using ko, or if it was built without embedding dependency information, this command will fail.`,
		Example: `
  # Fetch and extract Go dependency information from an image:
  ko deps docker.io/my-user/my-image:v3`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := createCancellableContext()

			ref, err := name.ParseReference(args[0])
			if err != nil {
				return err
			}

			p, err := makePlatform(platform)
			if err != nil {
				return err
			}

			img, err := remote.Image(ref,
				remote.WithContext(ctx),
				remote.WithAuthFromKeychain(authn.DefaultKeychain),
				remote.WithUserAgent(ua()),
				remote.WithPlatform(p))
			if err != nil {
				return err
			}

			rc := mutate.Extract(img)
			defer rc.Close()
			tr := tar.NewReader(rc)
			for {
				// Stop reading if the context is cancelled.
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					// keep reading.
				}
				h, err := tr.Next()
				if err == io.EOF {
					return errors.New("no ko-built executable found")
				}
				if err != nil {
					return err
				}

				if strings.HasPrefix(h.Name, "/ko-app/") && h.Typeflag == tar.TypeReg {
					tmp, err := ioutil.TempFile("", filepath.Base(h.Name))
					if err != nil {
						return err
					}
					defer os.RemoveAll(tmp.Name()) // best effort: remove tmp file afterwards.
					defer tmp.Close()              // close it first.
					if _, err := io.Copy(tmp, tr); err != nil {
						return err
					}
					if err := os.Chmod(tmp.Name(), fs.FileMode(h.Mode)); err != nil {
						return err
					}
					cmd := exec.CommandContext(ctx, "go", "version", "-m", tmp.Name())
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					return cmd.Run()
				}
			}
			// unreachable
		},
	}
	deps.Flags().StringVar(&platform, "platform", "",
		"Which platform to use when pulling a multi-platform image. Format: <os>[/<arch>[/<variant>]][,platform]")
	topLevel.AddCommand(deps)
}

func makePlatform(platform string) (v1.Platform, error) {
	if platform == "" {
		platform = "linux/amd64"
	}
	if platform == "all" || strings.Contains("platform", ",") {
		return v1.Platform{}, errors.New("--platform cannot be 'all' or specify multiple platforms")
	}

	goos, goarch, goarm := os.Getenv("GOOS"), os.Getenv("GOARCH"), os.Getenv("GOARM")

	// Default to linux/amd64 unless GOOS and GOARCH are set.
	if goos != "" && goarch != "" {
		platform = path.Join(goos, goarch)
	}

	// Use GOARM for variant if it's set and GOARCH is arm.
	if strings.Contains(goarch, "arm") && goarm != "" {
		platform = path.Join(platform, "v"+goarm)
	}

	var p v1.Platform
	parts := strings.Split(platform, "/")
	if len(parts) > 0 {
		p.OS = parts[0]
	}
	if len(parts) > 1 {
		p.Architecture = parts[1]
	}
	if len(parts) > 2 {
		p.Variant = parts[2]
	}
	if len(parts) > 3 {
		return v1.Platform{}, fmt.Errorf("too many slashes in platform spec: %s", platform)
	}
	return p, nil
}
