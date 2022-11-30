// Copyright 2023 Google LLC All Rights Reserved.
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

package provenance

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	intoto "github.com/in-toto/in-toto-golang/in_toto"
	slsa02common "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/common"
	slsa02 "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/v0.2"
)

func GenerateSLSA(ctx context.Context, koVersion, name, file, command string, args, env []string, workingDir string, buildStart, buildFinish time.Time, koParameters []string) ([]byte, error) {
	digest, err := calculateDigest(file)
	if err != nil {
		return nil, err
	}

	koConfigPath := os.Getenv("KO_CONFIG_PATH")
	if koConfigPath == "" {
		koConfigPath = filepath.Join(workingDir, ".ko.yaml")
		if _, err := os.Stat(koConfigPath); os.IsNotExist(err) {
			koConfigPath = ""
		}
	}
	var koConfigDigest string
	if koConfigPath != "" {
		koConfigDigest, err = calculateDigest(koConfigPath)
		if err != nil {
			return nil, err
		}
	}

	buildStartRFC3339, err := time.Parse(time.RFC3339, buildStart.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}

	buildFinishRFC3339, err := time.Parse(time.RFC3339, buildFinish.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}

	var koEnv []string
	for _, v := range env {
		if strings.HasPrefix(v, "KO_") {
			koEnv = append(koEnv, v)
		}
	}

	goversionmEnv, err := goversionm(ctx, file)
	if err != nil {
		return nil, err
	}

	gitRemoteUrl, err := extractRepoRemoteGitURL(ctx)
	if err != nil {
		return nil, err
	}

	commitSHA, err := extractLastCommitSHA(ctx)
	if err != nil {
		return nil, err
	}

	type (
		step struct {
			WorkingDir string   `json:"workingDir"`
			Command    []string `json:"command"`
			Info       []string `json:"info"`
		}
		buildConfig struct {
			Steps   []step `json:"steps"`
			Version int    `json:"version"`
		}
	)

	provenance := intoto.ProvenanceStatementSLSA02{
		StatementHeader: intoto.StatementHeader{
			Type:          intoto.StatementInTotoV01,
			PredicateType: slsa02.PredicateSLSAProvenance,
			Subject: []intoto.Subject{
				{
					Name: name,
					Digest: slsa02common.DigestSet{
						"sha256": digest,
					},
				},
			},
		},
		Predicate: slsa02.ProvenancePredicate{
			Builder: slsa02common.ProvenanceBuilder{
				ID: "https://github.com/ko-build/ko@" + koVersion,
			},
			BuildType: "",
			Invocation: slsa02.ProvenanceInvocation{
				ConfigSource: slsa02.ConfigSource{
					URI:        gitRemoteUrl,
					Digest:     slsa02common.DigestSet{"sha256": koConfigDigest},
					EntryPoint: koConfigPath,
				},
				Parameters:  koParameters,
				Environment: koEnv,
			},
			BuildConfig: buildConfig{
				Steps:   []step{{WorkingDir: workingDir, Command: append([]string{command}, args...), Info: goversionmEnv}},
				Version: 0,
			},
			Metadata: &slsa02.ProvenanceMetadata{
				BuildStartedOn:  &buildStartRFC3339,
				BuildFinishedOn: &buildFinishRFC3339,
				Completeness: slsa02.ProvenanceComplete{
					Parameters:  true,
					Environment: true,
					Materials:   false,
				},
				Reproducible: false,
			},
			Materials: []slsa02common.ProvenanceMaterial{
				{
					URI: gitRemoteUrl,
					Digest: slsa02common.DigestSet{
						"sha1": commitSHA,
					},
				},
			},
		},
	}

	return json.Marshal(provenance)
}

func calculateDigest(file string) (string, error) {
	binary, err := os.OpenFile(file, os.O_RDONLY, 0)
	if err != nil {
		return "", err
	}
	defer binary.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, binary); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func isGitRepo(ctx context.Context) bool {
	out, err := run(ctx, "rev-parse", "--is-inside-work-tree")
	return err == nil && strings.TrimSpace(out) == "true"
}

// Run runs a git command and returns its output or errors.
func run(ctx context.Context, args ...string) (string, error) {
	return runWithEnv(ctx, []string{}, args...)
}

func runWithEnv(ctx context.Context, env []string, args ...string) (string, error) {
	extraArgs := []string{
		"-c", "log.showSignature=false",
	}
	args = append(extraArgs, args...)
	/* #nosec */
	cmd := exec.CommandContext(ctx, "git", args...)

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = append(cmd.Env, env...)

	err := cmd.Run()
	if err != nil {
		return "", errors.New(stderr.String())
	}

	return stdout.String(), nil
}

func extractLastCommitSHA(ctx context.Context) (string, error) {
	var commitSHA string
	if !isGitRepo(ctx) {
		return "", nil
	}
	out, err := run(ctx, "rev-parse", "head")
	if err != nil {
		return "", nil
	}

	commitSHA = strings.ReplaceAll(strings.Split(out, "\n")[0], "'", "")
	if err != nil {
		return "", nil
	}
	return commitSHA, nil
}

// ExtractRepoFromConfig gets the repo name from the Git config.
func extractRepoRemoteGitURL(ctx context.Context) (string, error) {
	if !isGitRepo(ctx) {
		return "", nil
	}
	out, err := run(ctx, "ls-remote", "--get-url")
	if err != nil {
		return "", nil
	}

	out = strings.ReplaceAll(strings.Split(out, "\n")[0], "'", "")
	if err != nil {
		return "", nil
	}

	var packageDownloadLocation string
	switch {
	case strings.HasPrefix(out, "git@"):
		packageDownloadLocation = "git+ssh://" + strings.TrimPrefix(out, "git@")
	case strings.HasPrefix(out, "https://"):
		packageDownloadLocation = "git+https://" + strings.TrimPrefix(out, "https://")
	default:
		return "", nil
	}

	packageDownloadLocation = strings.TrimRightFunc(packageDownloadLocation, func(c rune) bool {
		// In windows newline is \r\n
		return c == '\r' || c == '\n'
	})

	return packageDownloadLocation, nil
}

func goversionm(ctx context.Context, file string) ([]string, error) {
	out := bytes.NewBuffer(nil)
	cmd := exec.CommandContext(ctx, "go", "version", "-m", file)
	cmd.Stdout = out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	lines := strings.Split(out.String(), "\n")
	var parsed []string
	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		prefix := fields[0]
		if prefix != "build" {
			continue
		}
		value := strings.Join(fields[1:], " ")
		parsed = append(parsed, value)
	}
	return parsed, nil
}
