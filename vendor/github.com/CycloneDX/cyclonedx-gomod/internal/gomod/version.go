// This file is part of CycloneDX GoMod
//
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
// Copyright (c) OWASP Foundation. All Rights Reserved.

package gomod

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/rs/zerolog"
	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
)

// GetModuleVersion attempts to detect a given module's version.
//
// If no Git repository is found in moduleDir, directories will be traversed
// upwards until the root directory is reached. This is done to accommodate
// for multi-module repositories, where modules are not placed in the repo root.
func GetModuleVersion(logger zerolog.Logger, moduleDir string) (string, error) {
	logger.Debug().
		Str("moduleDir", moduleDir).
		Msg("detecting module version")

	repoDir, err := filepath.Abs(moduleDir)
	if err != nil {
		return "", err
	}

	for {
		if tagVersion, err := GetVersionFromTag(logger, repoDir); err == nil {
			return tagVersion, nil
		} else {
			if errors.Is(err, git.ErrRepositoryNotExists) {
				if strings.HasSuffix(repoDir, string(filepath.Separator)) {
					// filepath.Abs and filepath.Dir both return paths
					// that do not end with separators, UNLESS it's the
					// root dir. We can't move up any further.
					return "", fmt.Errorf("no git repository found")
				}
				repoDir = filepath.Dir(repoDir) // Move to the parent dir
				continue
			}

			return "", err
		}
	}
}

// GetVersionFromTag checks if the HEAD commit is annotated with a tag and if it is, returns that tag's name.
// If the HEAD commit is not tagged, a pseudo version will be generated and returned instead.
func GetVersionFromTag(logger zerolog.Logger, moduleDir string) (string, error) {
	repo, err := git.PlainOpen(moduleDir)
	if err != nil {
		return "", err
	}

	headRef, err := repo.Head()
	if err != nil {
		return "", err
	}

	headCommit, err := repo.CommitObject(headRef.Hash())
	if err != nil {
		return "", err
	}

	latestTag, err := GetLatestTag(logger, repo, headCommit)
	if err != nil {
		if errors.Is(err, plumbing.ErrObjectNotFound) {
			return module.PseudoVersion("v0", "", headCommit.Author.When, headCommit.Hash.String()[:12]), nil
		}

		return "", err
	}

	if latestTag.commit.Hash.String() == headCommit.Hash.String() {
		return latestTag.name, nil
	}

	return module.PseudoVersion(
		semver.Major(latestTag.name),
		latestTag.name,
		latestTag.commit.Author.When,
		latestTag.commit.Hash.String()[:12],
	), nil
}

type tag struct {
	name   string
	commit *object.Commit
}

// GetLatestTag determines the latest tag relative to HEAD.
// Only tags with valid semver are considered.
func GetLatestTag(logger zerolog.Logger, repo *git.Repository, headCommit *object.Commit) (*tag, error) {
	logger.Debug().
		Str("headCommit", headCommit.Hash.String()).
		Msg("getting latest tag for head commit")

	tagRefs, err := repo.Tags()
	if err != nil {
		return nil, err
	}

	var latestTag tag

	err = tagRefs.ForEach(func(ref *plumbing.Reference) error {
		if semver.IsValid(ref.Name().Short()) {
			rev := plumbing.Revision(ref.Name().String())

			commitHash, err := repo.ResolveRevision(rev)
			if err != nil {
				return err
			}

			commit, err := repo.CommitObject(*commitHash)
			if err != nil {
				return err
			}

			isBeforeOrAtHead := commit.Committer.When.Before(headCommit.Author.When) ||
				commit.Committer.When.Equal(headCommit.Committer.When)

			if isBeforeOrAtHead && (latestTag.commit == nil || commit.Committer.When.After(latestTag.commit.Committer.When)) {
				latestTag.name = ref.Name().Short()
				latestTag.commit = commit
			}
		} else {
			logger.Debug().
				Str("tag", ref.Name().Short()).
				Str("hash", ref.Hash().String()).
				Str("reason", "not a valid semver").
				Msg("skipping tag")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if latestTag.commit == nil {
		return nil, plumbing.ErrObjectNotFound
	}

	return &latestTag, nil
}
