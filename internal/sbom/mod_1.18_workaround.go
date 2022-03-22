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

package sbom

import (
	"bufio"
	"strings"
)

var tab = "\t"

// fixLocalReplaceDirectives is a workaround for go v1.18:
// debug.ParseBuildInfo expects a 3rd field (checksum), but go v1.18 does not populate the 3rd field
func fixLocalReplaceDirectives(data string) string {
	var lines []string
	sc := bufio.NewScanner(strings.NewReader(data))
	for sc.Scan() {
		elem := strings.Split(sc.Text(), tab)
		// Add extra tab to:
		// =>	../api	(devel)
		if len(elem) == 3 && elem[0] == "=>" && elem[2] == "(devel)" {
			lines = append(lines, sc.Text()+tab)
			continue
		}
		lines = append(lines, sc.Text())
	}
	return strings.Join(lines, "\n")
}
