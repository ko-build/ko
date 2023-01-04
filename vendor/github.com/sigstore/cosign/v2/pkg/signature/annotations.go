//
// Copyright 2021 The Sigstore Authors.
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

package signature

import (
	_ "crypto/sha256" // for `crypto.SHA256`
	"fmt"
	"strings"
)

type AnnotationsMap struct {
	Annotations map[string]interface{}
}

func (a *AnnotationsMap) Set(s string) error {
	if a.Annotations == nil {
		a.Annotations = map[string]interface{}{}
	}
	kvp := strings.SplitN(s, "=", 2)
	if len(kvp) != 2 {
		return fmt.Errorf("invalid flag: %s, expected key=value", s)
	}

	a.Annotations[kvp[0]] = kvp[1]
	return nil
}

func (a *AnnotationsMap) String() string {
	s := []string{}
	for k, v := range a.Annotations {
		s = append(s, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(s, ",")
}
