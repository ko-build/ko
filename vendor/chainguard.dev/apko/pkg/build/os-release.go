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

	"chainguard.dev/apko/pkg/build/types"
	"chainguard.dev/apko/pkg/options"
)

func maybeGenerateVendorReleaseFile(
	o *options.Options, ic *types.ImageConfiguration,
) error {
	if ic.OSRelease.ID == "" || ic.OSRelease.VersionID == "" {
		return nil
	}

	path := filepath.Join(o.WorkDir, "etc", fmt.Sprintf("%s-release", ic.OSRelease.ID))

	_, err := os.Stat(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	w, err := os.Create(path)
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = fmt.Fprintf(w, "%s\n", ic.OSRelease.VersionID)
	if err != nil {
		return err
	}

	return nil
}

func (di *defaultBuildImplementation) GenerateOSRelease(
	o *options.Options, ic *types.ImageConfiguration,
) error {
	path := filepath.Join(o.WorkDir, "etc", "os-release")

	osReleaseExists := true
	if _, err := os.Stat(path); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		o.Logger().Warnf("did not find /etc/os-release at %s", path)
		osReleaseExists = false
	}

	// If /etc/os-release does not exist, return an error that it already exists.
	// However, if the user is requesting an override, write over it anyway.
	// TODO: better than checking for "apko-generated image"
	if osReleaseExists && ic.OSRelease.Name == "apko-generated image" {
		return ErrOSReleaseAlreadyPresent
	}

	w, err := os.Create(path)
	if err != nil {
		return err
	}
	defer w.Close()

	if ic.OSRelease.ID != "" {
		_, err := fmt.Fprintf(w, "ID=%s\n", ic.OSRelease.ID)
		if err != nil {
			return err
		}
	}

	if ic.OSRelease.Name != "" {
		_, err := fmt.Fprintf(w, "NAME=\"%s\"\n", ic.OSRelease.Name)
		if err != nil {
			return err
		}
	}

	if ic.OSRelease.PrettyName != "" {
		_, err := fmt.Fprintf(w, "PRETTY_NAME=\"%s\"\n", ic.OSRelease.PrettyName)
		if err != nil {
			return err
		}
	}

	if ic.OSRelease.VersionID != "" {
		_, err := fmt.Fprintf(w, "VERSION_ID=%s\n", ic.OSRelease.VersionID)
		if err != nil {
			return err
		}
	}

	if ic.OSRelease.HomeURL != "" {
		_, err := fmt.Fprintf(w, "HOME_URL=\"%s\"\n", ic.OSRelease.HomeURL)
		if err != nil {
			return err
		}
	}

	if ic.OSRelease.BugReportURL != "" {
		_, err := fmt.Fprintf(w, "BUG_REPORT_URL=\"%s\"\n", ic.OSRelease.BugReportURL)
		if err != nil {
			return err
		}
	}

	if err := maybeGenerateVendorReleaseFile(o, ic); err != nil {
		return err
	}

	return nil
}
