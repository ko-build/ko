// Copyright 2022 Google LLC All Rights Reserved.
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

package publish_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/ko/pkg/build"
	"github.com/google/ko/pkg/publish"
)

func TestNoOp(t *testing.T) {
	repoName := "quarkus.io/charm"
	importPath := "crane"
	noop, err := publish.NewNoOp(repoName)
	if err != nil {
		t.Fatalf("NewNoOp() = %v", err)
	}
	img, err := random.Image(1024, 1)
	if err != nil {
		t.Fatalf("random.Image() = %v", err)
	}
	ref, err := noop.Publish(context.TODO(), img, build.StrictScheme+importPath)
	if err != nil {
		t.Fatalf("Publish() = %v", err)
	}
	if !strings.HasPrefix(ref.String(), repoName) {
		t.Errorf("Publish() = %v, wanted preifx %s", ref, repoName)
	}
}

func TestNoOpWithCustomNamer(t *testing.T) {
	repoName := "quarkus.io/charm"
	importPath := "crane"
	noop, err := publish.NewNoOp(repoName, publish.NoOpWithNamer(md5Hash))
	if err != nil {
		t.Fatalf("NewNoOp() = %v", err)
	}
	img, err := random.Image(1024, 1)
	if err != nil {
		t.Fatalf("random.Image() = %v", err)
	}
	ref, err := noop.Publish(context.TODO(), img, build.StrictScheme+importPath)
	if err != nil {
		t.Fatalf("Publish() = %v", err)
	}
	if !strings.HasPrefix(ref.String(), repoName) {
		t.Errorf("Publish() = %v, wanted preifx %s", ref, repoName)
	}
	if !strings.HasSuffix(ref.Context().String(), md5Hash("", strings.ToLower(importPath))) {
		t.Errorf("Publish() = %v, wanted suffix %v", ref.Context(), md5Hash("", importPath))
	}
}

func TestNoOpWithTags(t *testing.T) {
	cases := []struct {
		name            string
		tags            []string
		expectedTag     string
		expectOpenError bool
	}{
		{
			name: "no tags",
		},
		{
			name: "latest tag",
			tags: []string{"latest"},
		},
		{
			name:        "multiple tags",
			tags:        []string{"v0.1", "v0.1.1"},
			expectedTag: "v0.1",
		},
		{
			name:        "single tag",
			tags:        []string{"v0.1"},
			expectedTag: "v0.1",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repoName := "quarkus.io/charm"
			importPath := "crane"
			noop, err := publish.NewNoOp(repoName, publish.NoOpWithTags(tc.tags))
			if tc.expectOpenError {
				if err == nil {
					t.Error("NewNoOp() - expected error, got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("NewNoOp() = %v", err)
			}
			img, err := random.Image(1024, 1)
			if err != nil {
				t.Fatalf("random.Image() = %v", err)
			}
			ref, err := noop.Publish(context.TODO(), img, build.StrictScheme+importPath)
			if err != nil {
				t.Fatalf("Publish() = %v", err)
			}
			if !strings.HasPrefix(ref.String(), repoName) {
				t.Errorf("Publish() = %v, wanted preifx %s", ref, repoName)
			}
			if tc.expectedTag != "" && !strings.Contains(ref.String(), fmt.Sprintf(":%s@", tc.expectedTag)) {
				t.Errorf("Publish() = %v, expected tag %s", ref.String(), tc.expectedTag)
			}
		})
	}
}

func TestNoOpWithTagOnly(t *testing.T) {
	cases := []struct {
		name            string
		tags            []string
		expectedTag     string
		expectOpenError bool
	}{
		{
			name:            "no tags",
			expectOpenError: true,
		},
		{
			name:            "latest tag",
			tags:            []string{"latest"},
			expectOpenError: true,
		},
		{
			name:            "multiple tags",
			tags:            []string{"v0.1", "v0.1.1"},
			expectOpenError: true,
		},
		{
			name:        "single tag",
			tags:        []string{"v0.1"},
			expectedTag: "v0.1",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repoName := "quarkus.io/charm"
			importPath := "crane"
			noop, err := publish.NewNoOp(repoName,
				publish.NoOpWithTags(tc.tags),
				publish.NoOpWithTagOnly(true))
			if tc.expectOpenError {
				if err == nil {
					t.Error("NewNoOp() - expected error, got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("NewNoOp() = %v", err)
			}
			img, err := random.Image(1024, 1)
			if err != nil {
				t.Fatalf("random.Image() = %v", err)
			}
			ref, err := noop.Publish(context.TODO(), img, build.StrictScheme+importPath)
			if err != nil {
				t.Fatalf("Publish() = %v", err)
			}
			if !strings.HasPrefix(ref.String(), repoName) {
				t.Errorf("Publish() = %v, wanted preifx %s", ref, repoName)
			}
			if tc.expectedTag != "" && !strings.HasSuffix(ref.String(), fmt.Sprintf(":%s", tc.expectedTag)) {
				t.Errorf("Publish() = %v, expected only tag %s", ref.String(), tc.expectedTag)
			}
		})
	}
}
