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

package git_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/ko/pkg/internal/git"
	"github.com/google/ko/pkg/internal/gittesting"
	"github.com/stretchr/testify/require"
)

const fakeGitURL = "git@github.com:foo/bar.git"

func TestNotAGitFolder(t *testing.T) {
	dir := t.TempDir()
	i, err := git.GetInfo(context.TODO(), dir)
	require.ErrorIs(t, err, git.ErrNotRepository)

	tpl := i.TemplateValue()
	requireEmpty(t, tpl)
}

func TestSingleCommit(t *testing.T) {
	dir := t.TempDir()
	gittesting.GitInit(t, dir)
	gittesting.GitRemoteAdd(t, dir, fakeGitURL)
	gittesting.GitCommit(t, dir, "commit1")
	gittesting.GitTag(t, dir, "v0.0.1")
	i, err := git.GetInfo(context.TODO(), dir)
	require.NoError(t, err)

	tpl := i.TemplateValue()
	require.Equal(t, "main", tpl["Branch"])
	require.Equal(t, "v0.0.1", tpl["Tag"])
	require.NotEmpty(t, tpl["ShortCommit"].(string))
	require.NotEmpty(t, tpl["FullCommit"].(string))
	require.NotEmpty(t, tpl["CommitDate"].(string))
	require.NotZero(t, tpl["CommitTimestamp"].(int64))
	require.False(t, tpl["IsDirty"].(bool))
	require.True(t, tpl["IsClean"].(bool))
	require.Equal(t, "clean", tpl["TreeState"])
}

func TestBranch(t *testing.T) {
	dir := t.TempDir()
	gittesting.GitInit(t, dir)
	gittesting.GitRemoteAdd(t, dir, fakeGitURL)
	gittesting.GitCommit(t, dir, "test-branch-commit")
	gittesting.GitTag(t, dir, "test-branch-tag")
	gittesting.GitCheckoutBranch(t, dir, "test-branch")
	i, err := git.GetInfo(context.TODO(), dir)
	require.NoError(t, err)

	tpl := i.TemplateValue()
	require.Equal(t, "test-branch", tpl["Branch"])
	require.Equal(t, "test-branch-tag", tpl["Tag"])
	require.NotEmpty(t, tpl["ShortCommit"].(string))
	require.NotEmpty(t, tpl["FullCommit"].(string))
	require.NotEmpty(t, tpl["CommitDate"].(string))
	require.NotZero(t, tpl["CommitTimestamp"].(int64))
	require.False(t, tpl["IsDirty"].(bool))
	require.True(t, tpl["IsClean"].(bool))
	require.Equal(t, "clean", tpl["TreeState"])
}

func TestNewRepository(t *testing.T) {
	dir := t.TempDir()
	gittesting.GitInit(t, dir)
	i, err := git.GetInfo(context.TODO(), dir)
	// TODO: improve this error handling
	require.ErrorContains(t, err, `fatal: ambiguous argument 'HEAD'`)

	tpl := i.TemplateValue()
	requireEmpty(t, tpl)
}

func TestNoTags(t *testing.T) {
	dir := t.TempDir()
	gittesting.GitInit(t, dir)
	gittesting.GitRemoteAdd(t, dir, fakeGitURL)
	gittesting.GitCommit(t, dir, "first")
	i, err := git.GetInfo(context.TODO(), dir)
	require.ErrorIs(t, err, git.ErrNoTag)

	tpl := i.TemplateValue()
	require.Equal(t, "main", tpl["Branch"])
	require.Equal(t, "v0.0.0", tpl["Tag"])
	require.NotEmpty(t, tpl["ShortCommit"].(string))
	require.NotEmpty(t, tpl["FullCommit"].(string))
	require.NotEmpty(t, tpl["CommitDate"].(string))
	require.NotZero(t, tpl["CommitTimestamp"].(int64))
	require.False(t, tpl["IsDirty"].(bool))
	require.True(t, tpl["IsClean"].(bool))
	require.Equal(t, "clean", tpl["TreeState"])
}

func TestDirty(t *testing.T) {
	dir := t.TempDir()
	gittesting.GitInit(t, dir)
	gittesting.GitRemoteAdd(t, dir, fakeGitURL)
	testFile, err := os.Create(filepath.Join(dir, "testFile"))
	require.NoError(t, err)
	require.NoError(t, testFile.Close())
	gittesting.GitAdd(t, dir)
	gittesting.GitCommit(t, dir, "commit2")
	gittesting.GitTag(t, dir, "v0.0.1")
	require.NoError(t, os.WriteFile(testFile.Name(), []byte("lorem ipsum"), 0o644))
	i, err := git.GetInfo(context.TODO(), dir)
	require.ErrorContains(t, err, "git is in a dirty state")

	tpl := i.TemplateValue()
	require.Equal(t, "main", tpl["Branch"])
	require.Equal(t, "v0.0.1", tpl["Tag"])
	require.NotEmpty(t, tpl["ShortCommit"].(string))
	require.NotEmpty(t, tpl["FullCommit"].(string))
	require.NotEmpty(t, tpl["CommitDate"].(string))
	require.NotZero(t, tpl["CommitTimestamp"].(int64))
	require.True(t, tpl["IsDirty"].(bool))
	require.False(t, tpl["IsClean"].(bool))
	require.Equal(t, "dirty", tpl["TreeState"])
}

func TestValidState(t *testing.T) {
	dir := t.TempDir()
	gittesting.GitInit(t, dir)
	gittesting.GitRemoteAdd(t, dir, fakeGitURL)
	gittesting.GitCommit(t, dir, "commit3")
	gittesting.GitTag(t, dir, "v0.0.1")
	gittesting.GitTag(t, dir, "v0.0.2")
	gittesting.GitCommit(t, dir, "commit4")
	gittesting.GitTag(t, dir, "v0.0.3")
	i, err := git.GetInfo(context.TODO(), dir)
	require.NoError(t, err)
	require.Equal(t, "v0.0.3", i.Tag)
	require.False(t, i.Dirty)
}

func TestGitNotInPath(t *testing.T) {
	t.Setenv("PATH", "")
	i, err := git.GetInfo(context.TODO(), "")
	require.ErrorIs(t, err, git.ErrNoGit)

	tpl := i.TemplateValue()
	requireEmpty(t, tpl)
}

func requireEmpty(t *testing.T, tpl map[string]any) {
	require.Equal(t, "", tpl["Branch"])
	require.Equal(t, "", tpl["Tag"])
	require.Equal(t, "", tpl["ShortCommit"])
	require.Equal(t, "", tpl["FullCommit"])
	require.False(t, tpl["IsDirty"].(bool))
	require.True(t, tpl["IsClean"].(bool))
	require.Equal(t, "clean", tpl["TreeState"])
}
