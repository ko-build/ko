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

package apk

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"fmt"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"golang.org/x/sync/errgroup"

	"chainguard.dev/apko/pkg/build/types"
	"chainguard.dev/apko/pkg/exec"
	"chainguard.dev/apko/pkg/options"
	"chainguard.dev/apko/pkg/sbom"
)

// Programmatic wrapper around apk-tools.  For now, this is done with os.Exec(),
// but this has been designed so that we can port it easily to use libapk-go once
// it is ready.

const (
	DefaultKeyRingPath       = "/etc/apk/keys"
	DefaultSystemKeyRingPath = "/usr/share/apk/keys/"
)

type APK struct {
	impl     apkImplementation
	executor *exec.Executor
	Options  options.Options
}

func New() *APK {
	return NewWithOptions(options.Default)
}

func NewWithOptions(o options.Options) *APK {
	a := &APK{
		Options: o,
		impl:    &apkDefaultImplementation{},
	}
	_ = a.buildExecutor()
	return a
}

type Option func(*APK) error

func (a *APK) buildExecutor() error {
	hostArch := types.ParseArchitecture(runtime.GOARCH)
	execOpts := []exec.Option{exec.WithProot(a.Options.UseProot)}
	if a.Options.UseProot && !a.Options.Arch.Compatible(hostArch) {
		a.Options.Log.Printf("%q requires QEMU (not compatible with %q)", a.Options.Arch, hostArch)
		execOpts = append(execOpts, exec.WithQemu(a.Options.Arch.ToQEmu()))
	}

	executor, err := exec.New(a.Options.WorkDir, a.Options.Logger(), execOpts...)
	if err != nil {
		return fmt.Errorf("building executor: %w", err)
	}
	a.executor = executor
	return nil
}

// Builds the image in Context.WorkDir according to the image configuration
func (a *APK) Initialize(ic *types.ImageConfiguration) error {
	// initialize apk
	if err := a.impl.InitDB(&a.Options, *a.executor); err != nil {
		return fmt.Errorf("failed to initialize apk database: %w", err)
	}

	var eg errgroup.Group

	eg.Go(func() error {
		if err := a.impl.InitKeyring(&a.Options, ic); err != nil {
			return fmt.Errorf("failed to initialize apk keyring: %w", err)
		}
		return nil
	})

	eg.Go(func() error {
		if err := a.impl.InitRepositories(&a.Options, ic); err != nil {
			return fmt.Errorf("failed to initialize apk repositories: %w", err)
		}
		return nil
	})

	eg.Go(func() error {
		if err := a.impl.InitWorld(&a.Options, ic); err != nil {
			return fmt.Errorf("failed to initialize apk world: %w", err)
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return err
	}

	// sync reality with desired apk world
	if err := a.impl.FixateWorld(&a.Options, a.executor); err != nil {
		return fmt.Errorf("failed to fixate apk world: %w", err)
	}

	eg.Go(func() error {
		if err := a.impl.NormalizeScriptsTar(&a.Options); err != nil {
			return fmt.Errorf("failed to normalize scripts.tar: %w", err)
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

// AdditionalTags is a helper function used in conjunction with the --package-version-tag flag
// If --package-version-tag is set to a package name (e.g. go), then this function
// returns a list of all images that should be published with the associated version of that package tagged (e.g. 1.18)
func AdditionalTags(opts options.Options) ([]string, error) {
	if opts.PackageVersionTag == "" {
		return nil, nil
	}
	dbPath := filepath.Join(opts.WorkDir, "lib/apk/db/installed")
	pkgs, err := sbom.ReadPackageIndex(&sbom.DefaultOptions, dbPath)
	if err != nil {
		return nil, err
	}
	for _, pkg := range pkgs {
		if pkg.Name != opts.PackageVersionTag {
			continue
		}
		version := pkg.Version
		if version == "" {
			opts.Log.Warningf("Version for package %s is empty", pkg.Name)
			continue
		}
		if opts.TagSuffix != "" {
			version += opts.TagSuffix
		}
		opts.Log.Debugf("Found version, images will be tagged with %s", version)

		additionalTags, err := appendTag(opts, fmt.Sprintf("%s%s", opts.PackageVersionTagPrefix, version))
		if err != nil {
			return nil, err
		}

		if opts.PackageVersionTagStem && len(additionalTags) > 0 {
			opts.Log.Debugf("Adding stemmed version tags")
			stemmedTags, err := getStemmedVersionTags(opts, additionalTags[0], version)
			if err != nil {
				return nil, err
			}
			additionalTags = append(additionalTags, stemmedTags...)
		}

		opts.Log.Infof("Returning additional tags %v", additionalTags)
		return additionalTags, nil
	}
	opts.Log.Warnf("No version info found for package %s, skipping additional tagging", opts.PackageVersionTag)
	return nil, nil
}

func appendTag(opts options.Options, newTag string) ([]string, error) {
	newTags := make([]string, len(opts.Tags))
	for i, t := range opts.Tags {
		nt, err := replaceTag(t, newTag)
		if err != nil {
			return nil, err
		}
		newTags[i] = nt
	}
	return newTags, nil
}

func replaceTag(img, newTag string) (string, error) {
	ref, err := name.ParseReference(img)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s", ref.Context(), newTag), nil
}

// TODO: use version parser from https://gitlab.alpinelinux.org/alpine/go/-/tree/master/version
func getStemmedVersionTags(opts options.Options, origRef string, version string) ([]string, error) {
	tags := []string{}
	re := regexp.MustCompile("[.]+")
	tmp := []string{}
	for _, part := range re.Split(version, -1) {
		tmp = append(tmp, part)
		additionalTag := strings.Join(tmp, ".")
		if additionalTag == version {
			tmp := strings.Split(version, "-")
			additionalTag = strings.Join(tmp[:len(tmp)-1], "-")
		}
		additionalTag, err := replaceTag(origRef,
			fmt.Sprintf("%s%s", opts.PackageVersionTagPrefix, additionalTag))
		if err != nil {
			return nil, err
		}
		tags = append(tags, additionalTag)
	}
	sort.Slice(tags, func(i, j int) bool {
		return tags[j] < tags[i]
	})
	return tags, nil
}

func (a *APK) SetImplementation(impl apkImplementation) {
	a.impl = impl
}
