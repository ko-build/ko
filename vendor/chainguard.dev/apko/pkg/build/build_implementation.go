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

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	ggcrtypes "github.com/google/go-containerregistry/pkg/v1/types"
	coci "github.com/sigstore/cosign/pkg/oci"
	"sigs.k8s.io/release-utils/hash"

	chainguardAPK "chainguard.dev/apko/pkg/apk"
	"chainguard.dev/apko/pkg/build/types"
	"chainguard.dev/apko/pkg/exec"
	apkofs "chainguard.dev/apko/pkg/fs"
	"chainguard.dev/apko/pkg/options"
	"chainguard.dev/apko/pkg/s6"
	"chainguard.dev/apko/pkg/sbom"
	soptions "chainguard.dev/apko/pkg/sbom/options"
	"chainguard.dev/apko/pkg/tarball"
)

//counterfeiter:generate . buildImplementation

type buildImplementation interface {
	// Refresh initialize build, set options, and get a jail and emulation executor and s6 supervisor config
	Refresh(*options.Options) (*s6.Context, *exec.Executor, error)
	// BuildTarball build from the layout in a working directory to an OCI image layer tarball
	BuildTarball(*options.Options) (string, error)
	// GenerateSBOM generate a software-bill-of-materials for the image
	GenerateSBOM(*options.Options, *types.ImageConfiguration) error
	// InstallBusyboxSymlinks add the symlinks for busybox
	InstallBusyboxSymlinks(*options.Options, *exec.Executor) error
	// UpdateLdconfig run ldconfig to create links and cache for shared libraries, see https://man7.org/linux/man-pages/man8/ldconfig.8.html
	UpdateLdconfig(*options.Options, *exec.Executor) error
	// InitializeApk do all of the steps to set up apk for installing packages in the working directory
	InitializeApk(*options.Options, *types.ImageConfiguration) error
	// MutateAccounts set up the user accounts and groups in the working directory
	MutateAccounts(*options.Options, *types.ImageConfiguration) error
	// MutatePaths set permissions and ownership on files based on the ImageConfiguration
	MutatePaths(*options.Options, *types.ImageConfiguration) error
	// GenerateOSRelase generate /etc/os-release in the working directory
	GenerateOSRelease(*options.Options, *types.ImageConfiguration) error
	// ValidateImageConfiguration check that the supplied ImageConfiguration is valid
	ValidateImageConfiguration(*types.ImageConfiguration) error
	// BuildImage based on the ImageConfiguration, run all of the steps to generate the laid out paths in the working directory
	BuildImage(*options.Options, *types.ImageConfiguration, *exec.Executor, *s6.Context) error
	// WriteSupervisionTree insert the configuration files and binaries in the working directory for s6 to operate
	WriteSupervisionTree(*s6.Context, *types.ImageConfiguration) error
	// GenerateIndexSBOM generate an SBOM for the index
	GenerateIndexSBOM(*options.Options, *types.ImageConfiguration, name.Digest, map[types.Architecture]coci.SignedImage) error
	// GenerateImageSBOM generate an SBOM for the image contents
	GenerateImageSBOM(*options.Options, *types.ImageConfiguration, coci.SignedImage) error
	// AdditionalTags generate additional tags for apk packages
	AdditionalTags(*options.Options) error
}

type defaultBuildImplementation struct{}

func (di *defaultBuildImplementation) Refresh(o *options.Options) (*s6.Context, *exec.Executor, error) {
	o.TarballPath = ""

	hostArch := types.ParseArchitecture(runtime.GOARCH)

	execOpts := []exec.Option{exec.WithProot(o.UseProot)}
	if o.UseProot && !o.Arch.Compatible(hostArch) {
		o.Logger().Debugf("%q requires QEMU (not compatible with %q)", o.Arch, hostArch)
		execOpts = append(execOpts, exec.WithQemu(o.Arch.ToQEmu()))
	}

	executor, err := exec.New(o.WorkDir, o.Logger(), execOpts...)
	if err != nil {
		return nil, nil, err
	}

	return s6.New(o.WorkDir, o.Logger()), executor, nil
}

func (di *defaultBuildImplementation) BuildTarball(o *options.Options) (string, error) {
	var outfile *os.File
	var err error

	if o.TarballPath != "" {
		outfile, err = os.Create(o.TarballPath)
	} else {
		outfile, err = os.Create(filepath.Join(o.TempDir(), o.TarballFileName()))
	}
	if err != nil {
		return "", fmt.Errorf("opening the build context tarball path failed: %w", err)
	}
	o.TarballPath = outfile.Name()
	defer outfile.Close()

	tw, err := tarball.NewContext(tarball.WithSourceDateEpoch(o.SourceDateEpoch))
	if err != nil {
		return "", fmt.Errorf("failed to construct tarball build context: %w", err)
	}

	if err := tw.WriteArchive(outfile, apkofs.DirFS(o.WorkDir)); err != nil {
		return "", fmt.Errorf("failed to generate tarball for image: %w", err)
	}

	o.Logger().Infof("built image layer tarball as %s", outfile.Name())
	return outfile.Name(), nil
}

// GenerateImageSBOM generates an sbom for an image
func (di *defaultBuildImplementation) GenerateImageSBOM(o *options.Options, ic *types.ImageConfiguration, img coci.SignedImage) error {
	if len(o.SBOMFormats) == 0 {
		o.Logger().Warnf("skipping SBOM generation")
		return nil
	}

	s := newSBOM(o, ic)

	if err := s.ReadLayerTarball(o.TarballPath); err != nil {
		return fmt.Errorf("reading layer tar: %w", err)
	}

	if err := s.ReadPackageIndex(); err != nil {
		return fmt.Errorf("getting installed packages from sbom: %w", err)
	}

	// Get the image digest
	h, err := img.Digest()
	if err != nil {
		return fmt.Errorf("getting %s image digest: %w", o.Arch, err)
	}

	s.Options.ImageInfo.ImageDigest = h.String()
	s.Options.ImageInfo.Arch = o.Arch

	if _, err := s.Generate(); err != nil {
		return fmt.Errorf("generating SBOMs: %w", err)
	}

	return nil
}

// GenerateSBOM generates an SBOM for an apko layer
func (di *defaultBuildImplementation) GenerateSBOM(o *options.Options, ic *types.ImageConfiguration) error {
	if len(o.SBOMFormats) == 0 {
		o.Logger().Warnf("skipping SBOM generation")
		return nil
	}

	s := newSBOM(o, ic)

	if err := s.ReadLayerTarball(o.TarballPath); err != nil {
		return fmt.Errorf("reading layer tar: %w", err)
	}

	if err := s.ReadReleaseData(); err != nil {
		return fmt.Errorf("getting os-release: %w", err)
	}

	if err := s.ReadPackageIndex(); err != nil {
		return fmt.Errorf("getting installed packages from sbom: %w", err)
	}

	s.Options.ImageInfo.Arch = o.Arch

	if _, err := s.Generate(); err != nil {
		return fmt.Errorf("generating SBOMs: %w", err)
	}

	return nil
}

func (di *defaultBuildImplementation) InitializeApk(o *options.Options, ic *types.ImageConfiguration) error {
	apk := chainguardAPK.NewWithOptions(*o)
	return apk.Initialize(ic)
}

func (di *defaultBuildImplementation) AdditionalTags(o *options.Options) error {
	at, err := chainguardAPK.AdditionalTags(*o)
	if err != nil {
		return err
	}
	if at == nil {
		return nil
	}
	o.Tags = append(o.Tags, at...)
	return nil
}

func (di *defaultBuildImplementation) BuildImage(
	o *options.Options, ic *types.ImageConfiguration, e *exec.Executor, s6context *s6.Context,
) error {
	return buildImage(di, o, ic, e, s6context)
}

// buildImage is a temporary function to make the fakes work.
// TODO(puerco): In order to have a structure we can mock, we need to split
// image building to its own interface or split out to its own package.
func buildImage(
	di buildImplementation, o *options.Options, ic *types.ImageConfiguration,
	e *exec.Executor, s6context *s6.Context,
) error {
	o.Logger().Infof("doing pre-flight checks")
	if err := di.ValidateImageConfiguration(ic); err != nil {
		return fmt.Errorf("failed to validate configuration: %w", err)
	}

	o.Logger().Infof("building image fileystem in %s", o.WorkDir)

	if err := di.InitializeApk(o, ic); err != nil {
		return fmt.Errorf("initializing apk: %w", err)
	}

	if err := di.AdditionalTags(o); err != nil {
		return fmt.Errorf("adding additional tags: %w", err)
	}

	if err := di.MutateAccounts(o, ic); err != nil {
		return fmt.Errorf("failed to mutate accounts: %w", err)
	}

	if err := di.MutatePaths(o, ic); err != nil {
		return fmt.Errorf("failed to mutate paths: %w", err)
	}

	// maybe install busybox symlinks
	if err := di.InstallBusyboxSymlinks(o, e); err != nil {
		return fmt.Errorf("failed to install busybox symlinks: %w", err)
	}

	// maybe run ldconfig to update it (e.g. on glibc)
	if err := di.UpdateLdconfig(o, e); err != nil {
		return fmt.Errorf("failed to update ldconfig: %w", err)
	}

	if err := di.GenerateOSRelease(o, ic); err != nil {
		if errors.Is(err, ErrOSReleaseAlreadyPresent) {
			o.Logger().Warnf("did not generate /etc/os-release: %v", err)
		} else {
			return fmt.Errorf("failed to generate /etc/os-release: %w", err)
		}
	}

	if err := di.WriteSupervisionTree(s6context, ic); err != nil {
		return fmt.Errorf("failed to write supervision tree: %w", err)
	}

	o.Logger().Infof("finished building filesystem in %s", o.WorkDir)

	return nil
}

func newSBOM(o *options.Options, ic *types.ImageConfiguration) *sbom.SBOM {
	s := sbom.NewWithWorkDir(o.WorkDir, o.Arch)
	// Parse the image reference
	if len(o.Tags) > 0 {
		tag, err := name.NewTag(o.Tags[0])
		if err == nil {
			s.Options.ImageInfo.Tag = tag.TagStr()
			s.Options.ImageInfo.Name = tag.String()
		} else {
			o.Logger().Errorf("%s parsing tag %s, ignoring", o.Tags[0], err)
		}
	}

	s.Options.ImageInfo.SourceDateEpoch = o.SourceDateEpoch
	s.Options.Formats = o.SBOMFormats
	s.Options.ImageInfo.VCSUrl = ic.VCSUrl

	if o.UseDockerMediaTypes {
		s.Options.ImageInfo.ImageMediaType = ggcrtypes.DockerManifestSchema2
	} else {
		s.Options.ImageInfo.ImageMediaType = ggcrtypes.OCIManifestSchema1
	}

	s.Options.OutputDir = o.TempDir()
	if o.SBOMPath != "" {
		s.Options.OutputDir = o.SBOMPath
	}

	return s
}

func (di *defaultBuildImplementation) GenerateIndexSBOM(
	o *options.Options, ic *types.ImageConfiguration,
	indexDigest name.Digest, imgs map[types.Architecture]coci.SignedImage,
) error {
	if len(o.SBOMFormats) == 0 {
		o.Logger().Warnf("skipping index SBOM generation")
		return nil
	}

	s := newSBOM(o, ic)
	o.Logger().Infof("Generating index SBOM")

	// Add the image digest
	h, err := v1.NewHash(indexDigest.DigestStr())
	if err != nil {
		return errors.New("getting index hash")
	}
	s.Options.ImageInfo.IndexDigest = h

	s.Options.ImageInfo.IndexMediaType = ggcrtypes.OCIImageIndex
	if o.UseDockerMediaTypes {
		s.Options.ImageInfo.IndexMediaType = ggcrtypes.DockerManifestList
	}
	var ext string
	switch o.SBOMFormats[0] {
	case "spdx":
		ext = "spdx.json"
	case "cyclonedx":
		ext = "cdx"
	case "idb":
		ext = "idb"
	}

	// Load the images data into the SBOM generator options
	for arch, i := range imgs {
		sbomHash, err := hash.SHA256ForFile(filepath.Join(s.Options.OutputDir, fmt.Sprintf("sbom-%s.%s", arch.ToAPK(), ext)))
		if err != nil {
			return fmt.Errorf("checksumming %s SBOM: %w", arch, err)
		}

		d, err := i.Digest()
		if err != nil {
			return fmt.Errorf("getting arch image digest: %w", err)
		}

		s.Options.ImageInfo.Images = append(
			s.Options.ImageInfo.Images,
			soptions.ArchImageInfo{
				Digest:     d,
				Arch:       arch,
				SBOMDigest: sbomHash,
			})
	}

	if _, err := s.GenerateIndex(); err != nil {
		return fmt.Errorf("generting index SBOM: %w", err)
	}

	return nil
}
