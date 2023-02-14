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

package types

import (
	"fmt"
	"os"

	"github.com/jinzhu/copier"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"chainguard.dev/apko/pkg/fetch"
	"chainguard.dev/apko/pkg/vcs"
)

// Attempt to probe an upstream VCS URL if known.
func (ic *ImageConfiguration) ProbeVCSUrl(imageConfigPath string, logger *logrus.Entry) {
	url, err := vcs.ProbeDirFromPath(imageConfigPath)
	if err != nil {
		logger.Debugf("failed to probe VCS URL: %v", err)
		return
	}

	if url != "" {
		ic.VCSUrl = url
		logger.Printf("detected %s as VCS URL", ic.VCSUrl)
	}
}

// Parse a configuration blob into an ImageConfiguration struct.
func (ic *ImageConfiguration) parse(configData []byte, logger *logrus.Entry) error {
	if err := yaml.Unmarshal(configData, ic); err != nil {
		return fmt.Errorf("failed to parse image configuration: %w", err)
	}

	if ic.Include != "" {
		logger.Printf("including %s for configuration", ic.Include)

		baseIc := ImageConfiguration{}

		if err := baseIc.Load(ic.Include, logger); err != nil {
			return fmt.Errorf("failed to read include file: %w", err)
		}

		mergedIc := ImageConfiguration{}

		// Copy the base configuration...
		if err := copier.Copy(&mergedIc, &baseIc); err != nil {
			return fmt.Errorf("failed to copy base configuration: %w", err)
		}

		// ... and then overlay the local configuration on top.
		if err := copier.CopyWithOption(&mergedIc, ic, copier.Option{IgnoreEmpty: true}); err != nil {
			return fmt.Errorf("failed to overlay specific configuration: %w", err)
		}

		// Now copy the merged configuration back to ic.
		if err := copier.Copy(ic, &mergedIc); err != nil {
			return fmt.Errorf("failed to copy merged configuration: %w", err)
		}

		// Merge packages, repositories and keyrings.
		keyring := append([]string{}, baseIc.Contents.Keyring...)
		keyring = append(keyring, mergedIc.Contents.Keyring...)
		ic.Contents.Keyring = keyring

		repos := append([]string{}, baseIc.Contents.Repositories...)
		repos = append(repos, mergedIc.Contents.Repositories...)
		ic.Contents.Repositories = repos

		pkgs := append([]string{}, baseIc.Contents.Packages...)
		pkgs = append(pkgs, mergedIc.Contents.Packages...)
		ic.Contents.Packages = pkgs
	}

	return nil
}

// Loads an image configuration given a configuration file path.
func (ic *ImageConfiguration) Load(imageConfigPath string, logger *logrus.Entry) error {
	data, err := os.ReadFile(imageConfigPath)
	if err == nil {
		return ic.parse(data, logger)
	}

	// At this point, we're doing a remote config file.
	logger.Warnf("remote configurations are an experimental feature and subject to change.")

	data, err = fetch.Fetch(imageConfigPath)
	if err != nil {
		return fmt.Errorf("unable to fetch remote include from git: %w", err)
	}

	return ic.parse(data, logger)
}

// Do preflight checks and mutations on an image configuration.
func (ic *ImageConfiguration) Validate() error {
	if ic.Entrypoint.Type == "service-bundle" {
		if err := ic.ValidateServiceBundle(); err != nil {
			return err
		}
	}

	for _, u := range ic.Accounts.Users {
		if u.UserName == "" {
			return fmt.Errorf("configured user %v has no configured user name", u)
		}

		if u.UID == 0 {
			return fmt.Errorf("configured user %v has UID 0", u)
		}
	}

	for _, g := range ic.Accounts.Groups {
		if g.GroupName == "" {
			return fmt.Errorf("configured group %v has no configured group name", g)
		}

		if g.GID == 0 {
			return fmt.Errorf("configured group %v has GID 0", g)
		}
	}

	if ic.OSRelease.ID == "" {
		ic.OSRelease.ID = "unknown"
	}

	if ic.OSRelease.Name == "" {
		ic.OSRelease.Name = "apko-generated image"
		ic.OSRelease.PrettyName = "apko-generated image"
	}

	if ic.OSRelease.VersionID == "" {
		ic.OSRelease.VersionID = "unknown"
	}

	if ic.OSRelease.HomeURL == "" {
		ic.OSRelease.HomeURL = "https://github.com/chainguard-dev/apko"
	}

	return nil
}

// Do preflight checks and mutations on an image configured to manage
// a service bundle.
func (ic *ImageConfiguration) ValidateServiceBundle() error {
	ic.Entrypoint.Command = "/bin/s6-svscan /sv"

	// It's harmless to have a duplicate entry in /etc/apk/world,
	// apk will fix it up when the fixate op happens.
	ic.Contents.Packages = append(ic.Contents.Packages, "s6")

	return nil
}

func (ic *ImageConfiguration) Summarize(logger *logrus.Entry) {
	logger.Printf("image configuration:")
	logger.Printf("  contents:")
	logger.Printf("    repositories: %v", ic.Contents.Repositories)
	logger.Printf("    keyring:      %v", ic.Contents.Keyring)
	logger.Printf("    packages:     %v", ic.Contents.Packages)
	if ic.Entrypoint.Type != "" || ic.Entrypoint.Command != "" || len(ic.Entrypoint.Services) != 0 {
		logger.Printf("  entrypoint:")
		logger.Printf("    type:    %s", ic.Entrypoint.Type)
		logger.Printf("    command:     %s", ic.Entrypoint.Command)
		logger.Printf("    service: %v", ic.Entrypoint.Services)
		logger.Printf("    shell fragment: %v", ic.Entrypoint.ShellFragment)
	}
	if ic.Cmd != "" {
		logger.Printf("  cmd: %s", ic.Cmd)
	}
	if ic.Accounts.RunAs != "" || len(ic.Accounts.Users) != 0 || len(ic.Accounts.Groups) != 0 {
		logger.Printf("  accounts:")
		logger.Printf("    runas:  %s", ic.Accounts.RunAs)
		logger.Printf("    users:")
		for _, u := range ic.Accounts.Users {
			logger.Printf("      - uid=%d(%s) gid=%d", u.UID, u.UserName, u.GID)
		}
		logger.Printf("    groups:")
		for _, g := range ic.Accounts.Groups {
			logger.Printf("      - gid=%d(%s) members=%v", g.GID, g.GroupName, g.Members)
		}
	}
	if len(ic.Annotations) > 0 {
		logger.Printf("    annotations:")
		for k, v := range ic.Annotations {
			logger.Printf("      %s: %s", k, v)
		}
	}
}
