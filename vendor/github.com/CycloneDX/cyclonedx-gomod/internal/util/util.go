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

package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// IsSubPath checks (lexically) if subject is a subpath of path.
func IsSubPath(subject, path string) (bool, error) {
	pathAbs, err := filepath.Abs(path)
	if err != nil {
		return false, fmt.Errorf("failed to make %s absolute: %w", path, err)
	}

	subjectAbs, err := filepath.Abs(subject)
	if err != nil {
		return false, fmt.Errorf("failed to make %s absolute: %w", subject, err)
	}

	if !strings.HasPrefix(subjectAbs, pathAbs) {
		return false, nil
	}

	return true, nil
}

// StringsIndexOf determines the index of a string in a string slice.
func StringsIndexOf(haystack []string, needle string) int {
	for i := range haystack {
		if haystack[i] == needle {
			return i
		}
	}
	return -1
}
