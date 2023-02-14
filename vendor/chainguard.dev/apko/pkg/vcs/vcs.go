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

package vcs

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

const defaultRemoteName = "origin"

// Given a starting directory and toplevel directory, work backwards
// to the toplevel directory and probe for a Git repository, returning
// the origin URI if known.
func ProbeDirForVCSUrl(startDir, toplevelDir string) (string, error) {
	repo, err := OpenRepository(startDir, toplevelDir)
	if err != nil {
		return "", fmt.Errorf("opening git repository: %w", err)
	}

	vcsURL, err := getRemoteURL(repo, defaultRemoteName)
	if err != nil {
		return "", fmt.Errorf("fetching remote URL: %w", err)
	}

	hash, err := resolveGitRevision(repo, "HEAD")
	if err != nil {
		return "", fmt.Errorf("resolving repo HEAD: %w", err)
	}
	if hash != "" {
		vcsURL = fmt.Sprintf("%s@%s", vcsURL, hash)
	}

	return vcsURL, nil
}

// Given a starting directory, work backwards to the current working
// directory and probe for a Git repository, returning the origin URI
// if known.
func ProbeDirFromPath(startingPath string) (string, error) {
	toplevelDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot find working directory: %w", err)
	}

	startingPath, err = filepath.Abs(startingPath)
	if err != nil {
		return "", fmt.Errorf("cannot dereference relative path %s: %w", startingPath, err)
	}

	fi, err := os.Stat(startingPath)
	if err != nil {
		return "", fmt.Errorf("cannot check start directory: %w", err)
	}

	// If starting path is not a directory, get the parent directory.
	// This way we can just pass things like "foo/apko.yaml" as the
	// input here.
	if !fi.IsDir() {
		startingPath = filepath.Dir(startingPath)
	}

	return ProbeDirForVCSUrl(startingPath, toplevelDir)
}

func getRemoteURL(repo *git.Repository, remoteName string) (string, error) {
	if remoteName == "" {
		remoteName = defaultRemoteName
	}
	remote, err := repo.Remote(remoteName)
	if err != nil {
		return "", fmt.Errorf("unable to determine upstream git vcs url: %w", err)
	}

	remoteConfig := remote.Config()
	remoteURL := remoteConfig.URLs[0]

	normalizedURL, err := url.Parse(remoteURL)
	if err != nil {
		// URL is most likely a git+ssh:// type URL, represented
		// in the way git itself does so.

		// Take the user@host:repo and turn it into user@host/repo.
		remoteURL = strings.Replace(remoteURL, ":", "/", 1)
		remoteURL = fmt.Sprintf("git+ssh://%s", remoteURL)

		normalizedURL, err = url.Parse(remoteURL)
		if err != nil {
			return "", fmt.Errorf("unable to parse %s as a git vcs url: %w", remoteURL, err)
		}
	}

	// sanitize any user authentication data from the VCS URL
	normalizedURL.User = nil

	return normalizedURL.String(), nil
}

// resolveGitRevision resolves a git revision
func resolveGitRevision(repo *git.Repository, revString string) (string, error) {
	rev := plumbing.Revision(revString)
	hash, err := repo.ResolveRevision(rev)
	if err != nil {
		return "", fmt.Errorf("resolving revision: %w", err)
	}

	return hash.String(), nil
}

// OpenRepository opens a repository in startDir checking every
// parent dir until topLevelDir
func OpenRepository(startDir, toplevelDir string) (repo *git.Repository, err error) {
	if toplevelDir == "" {
		toplevelDir = "/"
	}

	for l, d := range map[string]string{
		"start": startDir, "top-level": toplevelDir,
	} {
		d, err := filepath.Abs(d)
		if err != nil {
			return nil, fmt.Errorf("resolving %s directory: %w", l, err)
		}

		fi, err := os.Stat(d)
		if err != nil {
			return nil, fmt.Errorf("cannot check %s directory: %w", l, err)
		}

		if !fi.IsDir() {
			return nil, fmt.Errorf("%s path %s is not a directory", l, d)
		}

		if l == "start" {
			startDir = d
			continue
		}
		toplevelDir = d
	}

	searchPath := startDir

	for {
		if !strings.HasPrefix(searchPath, toplevelDir) {
			return nil, fmt.Errorf("path %s is not contained by %s", searchPath, toplevelDir)
		}

		repo, err := git.PlainOpen(searchPath)
		if err != nil {
			if searchPath == "/" || searchPath == toplevelDir {
				return nil, fmt.Errorf("directory %s is not in a git repository", startDir)
			}
			searchPath = filepath.Dir(searchPath)
			continue
		}
		return repo, nil
	}
}
