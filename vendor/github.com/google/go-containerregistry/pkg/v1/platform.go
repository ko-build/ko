// Copyright 2018 Google LLC All Rights Reserved.
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

package v1

import (
	"fmt"
	"sort"
	"strings"
)

// Platform represents the target os/arch for an image.
type Platform struct {
	Architecture string   `json:"architecture"`
	OS           string   `json:"os"`
	OSVersion    string   `json:"os.version,omitempty"`
	OSFeatures   []string `json:"os.features,omitempty"`
	Variant      string   `json:"variant,omitempty"`
	Features     []string `json:"features,omitempty"`
}

func (p Platform) String() string {
	if p.OS == "" {
		return ""
	}
	var b strings.Builder
	b.WriteString(p.OS)
	if p.Architecture != "" {
		b.WriteString("/")
		b.WriteString(p.Architecture)
	}
	if p.Variant != "" {
		b.WriteString("/")
		b.WriteString(p.Variant)
	}
	if p.OSVersion != "" {
		b.WriteString(":")
		b.WriteString(p.OSVersion)
	}
	if len(p.OSFeatures) != 0 {
		b.WriteString(" (osfeatures=")
		b.WriteString(strings.Join(p.OSFeatures, ","))
		b.WriteString(")")
	}
	if len(p.Features) != 0 {
		b.WriteString(" (features=")
		b.WriteString(strings.Join(p.Features, ","))
		b.WriteString(")")
	}
	return b.String()
}

// PlatformFromString parses a string representing a Platform, if possible.
func PlatformFromString(s string) (*Platform, error) {
	var p Platform
	parts := strings.Split(s, ":")
	if len(parts) == 2 {
		p.OSVersion = parts[1]
	}
	parts = strings.Split(parts[0], " ")
	if len(parts) > 3 {
		return nil, fmt.Errorf("too many spaces in platform spec: %s", s)
	}
	if len(parts) > 1 {
		for _, part := range parts[1:] {
			if strings.HasPrefix(part, "(osfeatures=") && strings.HasSuffix(part, ")") {
				pt := strings.TrimPrefix(part, "(osfeatures=")
				pt = strings.TrimSuffix(pt, ")")
				p.OSFeatures = strings.Split(pt, ",")
			} else if strings.HasPrefix(part, "(features=") && strings.HasSuffix(part, ")") {
				pt := strings.TrimPrefix(part, "(features=")
				pt = strings.TrimSuffix(pt, ")")
				p.Features = strings.Split(pt, ",")
			} else {
				return nil, fmt.Errorf("unknown parenthetical: %s", part)
			}
		}
	}

	parts = strings.Split(parts[0], "/")
	if len(parts) > 0 {
		p.OS = parts[0]
	}
	if len(parts) > 1 {
		p.Architecture = parts[1]
	}
	if len(parts) > 2 {
		p.Variant = parts[2]
	}
	if len(parts) > 3 {
		return nil, fmt.Errorf("too many slashes in platform spec: %s", s)
	}
	return &p, nil
}

// Equals returns true if the given platform is semantically equivalent to this one.
// The order of Features and OSFeatures is not important.
func (p Platform) Equals(o Platform) bool {
	return p.OS == o.OS &&
		p.Architecture == o.Architecture &&
		p.Variant == o.Variant &&
		p.OSVersion == o.OSVersion &&
		stringSliceEqualIgnoreOrder(p.OSFeatures, o.OSFeatures) &&
		stringSliceEqualIgnoreOrder(p.Features, o.Features)
}

// FuzzyEquals returns true if the given platform is semantically a "subset" of
// this one.
//
// - If this platform specifies only an OS, only the given platform's OS must
// be equal.
// - If this platform specifies an architecture, the given platform's
// architecture must be equal.
// - If this platform specifies a variant, the given platform's variant must be
// equal.
// - If this platform specifies osversion, the given platform's osversion must
// be equal, or this platform's osversion must be a partial match according to
// semantic versioning.
//
// The order of Features and OSFeatures is not important, if they are specified.
func (p Platform) FuzzyEquals(o Platform) bool {
	if p.OS != o.OS {
		return false
	}
	if p.Architecture == "" {
		return true
	}
	if p.Architecture != o.Architecture {
		return false
	}
	if p.Variant != "" && p.Variant != o.Variant {
		return false
	}
	if p.OSVersion == "" {
		return true
	}
	pparts, oparts := strings.Split(p.OSVersion, "."), strings.Split(o.OSVersion, ".")
	if len(pparts) > len(oparts) {
		return false
	}
	for i, ppart := range pparts {
		if ppart != oparts[i] {
			return false
		}
	}
	return stringSliceEqualIgnoreOrder(p.OSFeatures, o.OSFeatures) && stringSliceEqualIgnoreOrder(p.Features, o.Features)
}

// stringSliceEqual compares 2 string slices and returns if their contents are identical.
func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, elm := range a {
		if elm != b[i] {
			return false
		}
	}
	return true
}

// stringSliceEqualIgnoreOrder compares 2 string slices and returns if their contents are identical, ignoring order
func stringSliceEqualIgnoreOrder(a, b []string) bool {
	if a != nil && b != nil {
		sort.Strings(a)
		sort.Strings(b)
	}
	return stringSliceEqual(a, b)
}
