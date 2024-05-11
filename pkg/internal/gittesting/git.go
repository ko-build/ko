// Copyright 2024 ko Build Authors All Rights Reserved.
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

// MIT License
//
// Copyright (c) 2016-2022 Carlos Alexandro Becker
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package gittesting

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
	"testing"
)

// GitInit inits a new git project.
func GitInit(t *testing.T, dir string) {
	t.Helper()
	out, err := fakeGit(dir, "init")
	requireNoError(t, err)
	requireContains(t, out, "Initialized empty Git repository", "")
	requireNoError(t, err)
	GitCheckoutBranch(t, dir, "main")
	_, _ = fakeGit("branch", "-D", "master")
}

// GitRemoteAdd adds the given url as remote.
func GitRemoteAdd(t *testing.T, dir, url string) {
	t.Helper()
	out, err := fakeGit(dir, "remote", "add", "origin", url)
	requireNoError(t, err)
	requireEmpty(t, out)
}

// GitCommit creates a git commits.
func GitCommit(t *testing.T, dir, msg string) {
	t.Helper()
	out, err := fakeGit(dir, "commit", "--allow-empty", "-m", msg)
	requireNoError(t, err)
	requireContains(t, out, "main", msg)
}

// GitTag creates a git tag.
func GitTag(t *testing.T, dir, tag string) {
	t.Helper()
	out, err := fakeGit(dir, "tag", tag)
	requireNoError(t, err)
	requireEmpty(t, out)
}

// GitAnnotatedTag creates an annotated tag.
func GitAnnotatedTag(t *testing.T, dir, tag, message string) {
	t.Helper()
	out, err := fakeGit(dir, "tag", "-a", tag, "-m", message)
	requireNoError(t, err)
	requireEmpty(t, out)
}

// GitAdd adds all files to stage.
func GitAdd(t *testing.T, dir string) {
	t.Helper()
	out, err := fakeGit(dir, "add", "-A")
	requireNoError(t, err)
	requireEmpty(t, out)
}

func fakeGit(dir string, args ...string) (string, error) {
	allArgs := []string{
		"-c", "user.name='GoReleaser'",
		"-c", "user.email='test@goreleaser.github.com'",
		"-c", "commit.gpgSign=false",
		"-c", "tag.gpgSign=false",
		"-c", "log.showSignature=false",
	}
	allArgs = append(allArgs, args...)
	return gitRun(dir, allArgs...)
}

// GitCheckoutBranch allows us to change the active branch that we're using.
func GitCheckoutBranch(t *testing.T, dir, name string) {
	t.Helper()
	out, err := fakeGit(dir, "checkout", "-b", name)
	requireNoError(t, err)
	requireEmpty(t, out)
}

func gitRun(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if err != nil {
		return "", errors.New(stderr.String())
	}

	return stdout.String(), nil
}

func requireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}

func requireContains(t *testing.T, val, expected, msg string) {
	t.Helper()
	if !strings.Contains(val, expected) {
		if len(msg) > 0 {
			t.Fatalf("%s: expected value %s missing from value %s", msg, expected, val)
		} else {
			t.Fatalf("expected value %s missing from value %s", expected, val)
		}
	}
}

func requireEmpty(t *testing.T, val string) {
	t.Helper()
	if len(val) > 0 {
		t.Fatalf("%s: expected empty string", val)
	}
}
