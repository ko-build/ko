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

package options

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	sigs "github.com/sigstore/cosign/v2/pkg/signature"
)

// AnnotationOptions is the top level wrapper for the annotations.
type AnnotationOptions struct {
	Annotations []string
}

var _ Interface = (*AnnotationOptions)(nil)

func (o *AnnotationOptions) AnnotationsMap() (sigs.AnnotationsMap, error) {
	ann := sigs.AnnotationsMap{}
	for _, a := range o.Annotations {
		kv := strings.Split(a, "=")
		if len(kv) != 2 {
			return ann, fmt.Errorf("unable to parse annotation: %s", a)
		}
		if ann.Annotations == nil {
			ann.Annotations = map[string]interface{}{}
		}
		ann.Annotations[kv[0]] = kv[1]
	}
	return ann, nil
}

// AddFlags implements Interface
func (o *AnnotationOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringSliceVarP(&o.Annotations, "annotations", "a", nil,
		"extra key=value pairs to sign")
}
