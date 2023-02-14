// Copyright 2022 Chainguard, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package build

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/hashicorp/go-multierror"
	coci "github.com/sigstore/cosign/pkg/oci"
	"github.com/sirupsen/logrus"

	"chainguard.dev/apko/pkg/build/types"
	"chainguard.dev/apko/pkg/exec"
	"chainguard.dev/apko/pkg/options"
	"chainguard.dev/apko/pkg/s6"
)

// Context contains all of the information necessary to build an
// OCI image. Includes the configurationfor the build,
// the path to the config file, the executor for root jails and
// architecture emulation, the s6 supervisor to add to the image,
// build options, and the `buildImplementation`, which handles the actual build.
type Context struct {
	impl buildImplementation
	// ImageConfiguration instructions to use for the build, normally from an apko.yaml file, but can be set directly.
	ImageConfiguration types.ImageConfiguration
	// ImageConfigFile path to the config file used, if any, to load the ImageConfiguration
	ImageConfigFile string
	executor        *exec.Executor
	s6              *s6.Context
	Assertions      []Assertion
	Options         options.Options
}

func (bc *Context) Summarize() {
	bc.Logger().Printf("build context:")
	bc.Options.Summarize(bc.Logger())
	bc.ImageConfiguration.Summarize(bc.Logger())
}

// BuildTarball calls the underlying implementation's BuildTarball
// which takes the fully populated working directory and saves it to
// an OCI image layer tar.gz file.
func (bc *Context) BuildTarball() (string, error) {
	return bc.impl.BuildTarball(&bc.Options)
}

func (bc *Context) GenerateImageSBOM(arch types.Architecture, img coci.SignedImage) error {
	opts := bc.Options
	opts.Arch = arch
	return bc.impl.GenerateImageSBOM(&opts, &bc.ImageConfiguration, img)
}

func (bc *Context) GenerateIndexSBOM(indexDigest name.Digest, imgs map[types.Architecture]coci.SignedImage) error {
	return bc.impl.GenerateIndexSBOM(&bc.Options, &bc.ImageConfiguration, indexDigest, imgs)
}

func (bc *Context) GenerateSBOM() error {
	return bc.impl.GenerateSBOM(&bc.Options, &bc.ImageConfiguration)
}

func (bc *Context) BuildImage() error {
	// TODO(puerco): Point to final interface (see comment on buildImage fn)
	return buildImage(bc.impl, &bc.Options, &bc.ImageConfiguration, bc.executor, bc.s6)
}

func (bc *Context) Logger() *logrus.Entry {
	return bc.Options.Logger()
}

// BuildLayer given the context set up, including
// build configuration and working directory,
// lays out all of the packages in the working directory,
// sets up the necessary user accounts and groups,
// and sets everything up in the directory. Then
// packages it all up into a standard OCI image layer
// tar.gz file.
func (bc *Context) BuildLayer() (string, error) {
	bc.Summarize()

	// build image filesystem
	if err := bc.BuildImage(); err != nil {
		return "", err
	}

	// run any assertions defined
	if err := bc.runAssertions(); err != nil {
		return "", err
	}

	// build layer tarball
	layerTarGZ, err := bc.BuildTarball()
	if err != nil {
		return "", err
	}

	// generate SBOM
	if bc.Options.WantSBOM {
		if err := bc.GenerateSBOM(); err != nil {
			return "", fmt.Errorf("generating SBOMs: %w", err)
		}
	} else {
		bc.Logger().Debug("Not generating SBOMs (WantSBOM = false)")
	}

	return layerTarGZ, nil
}

func (bc *Context) runAssertions() error {
	var eg multierror.Group

	for _, a := range bc.Assertions {
		a := a
		eg.Go(func() error { return a(bc) })
	}

	return eg.Wait().ErrorOrNil()
}

// New creates a build context.
// The SOURCE_DATE_EPOCH env variable is supported and will
// overwrite the provided timestamp if present.
func New(workDir string, opts ...Option) (*Context, error) {
	bc := Context{
		Options: options.Default,
		impl:    &defaultBuildImplementation{},
	}
	bc.Options.WorkDir = workDir

	for _, opt := range opts {
		if err := opt(&bc); err != nil {
			return nil, err
		}
	}

	// SOURCE_DATE_EPOCH will always overwrite the build flag
	if v, ok := os.LookupEnv("SOURCE_DATE_EPOCH"); ok {
		// The value MUST be an ASCII representation of an integer
		// with no fractional component, identical to the output
		// format of date +%s.
		sec, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			// If the value is malformed, the build process
			// SHOULD exit with a non-zero error code.
			return nil, fmt.Errorf("failed to parse SOURCE_DATE_EPOCH: %w", err)
		}

		bc.Options.SourceDateEpoch = time.Unix(sec, 0)
	}

	// if arch is missing default to the running program's arch
	zeroArch := types.Architecture{}
	if bc.Options.Arch == zeroArch {
		bc.Options.Arch = types.ParseArchitecture(runtime.GOARCH)
	}

	if bc.Options.WithVCS && bc.ImageConfiguration.VCSUrl == "" {
		bc.ImageConfiguration.ProbeVCSUrl(bc.ImageConfigFile, bc.Logger())
	}

	return &bc, nil
}

// Refresh initializes the build process by calling the underlying implementation's
// Refresh(), which includes getting the chroot/proot jailed process executor (and
// possibly architecture emulator), sets those on the Context, and returns.
func (bc *Context) Refresh() error {
	s6, executor, err := bc.impl.Refresh(&bc.Options)
	if err != nil {
		return err
	}

	bc.executor = executor
	bc.s6 = s6

	return nil
}

func (bc *Context) SetImplementation(i buildImplementation) {
	bc.impl = i
}
