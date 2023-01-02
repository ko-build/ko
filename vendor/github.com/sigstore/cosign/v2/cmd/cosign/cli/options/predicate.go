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
	"net/url"

	slsa "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/v0.2"

	"github.com/in-toto/in-toto-golang/in_toto"
	"github.com/spf13/cobra"

	"github.com/sigstore/cosign/v2/pkg/cosign/attestation"
)

const (
	PredicateCustom    = "custom"
	PredicateSLSA      = "slsaprovenance"
	PredicateSPDX      = "spdx"
	PredicateSPDXJSON  = "spdxjson"
	PredicateCycloneDX = "cyclonedx"
	PredicateLink      = "link"
	PredicateVuln      = "vuln"
)

// PredicateTypeMap is the mapping between the predicate `type` option to predicate URI.
var PredicateTypeMap = map[string]string{
	PredicateCustom:    attestation.CosignCustomProvenanceV01,
	PredicateSLSA:      slsa.PredicateSLSAProvenance,
	PredicateSPDX:      in_toto.PredicateSPDX,
	PredicateSPDXJSON:  in_toto.PredicateSPDX,
	PredicateCycloneDX: in_toto.PredicateCycloneDX,
	PredicateLink:      in_toto.PredicateLinkV1,
	PredicateVuln:      attestation.CosignVulnProvenanceV01,
}

// PredicateOptions is the wrapper for predicate related options.
type PredicateOptions struct {
	Type string
}

var _ Interface = (*PredicateOptions)(nil)

// AddFlags implements Interface
func (o *PredicateOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Type, "type", "custom",
		"specify a predicate type (slsaprovenance|link|spdx|spdxjson|cyclonedx|vuln|custom) or an URI")
}

// ParsePredicateType parses the predicate `type` flag passed into a predicate URI, or validates `type` is a valid URI.
func ParsePredicateType(t string) (string, error) {
	uri, ok := PredicateTypeMap[t]
	if !ok {
		if _, err := url.ParseRequestURI(t); err != nil {
			return "", fmt.Errorf("invalid predicate type: %s", t)
		}
		uri = t
	}
	return uri, nil
}

// PredicateLocalOptions is the wrapper for predicate related options.
type PredicateLocalOptions struct {
	PredicateOptions
	Path string
}

var _ Interface = (*PredicateLocalOptions)(nil)

// AddFlags implements Interface
func (o *PredicateLocalOptions) AddFlags(cmd *cobra.Command) {
	o.PredicateOptions.AddFlags(cmd)

	cmd.Flags().StringVar(&o.Path, "predicate", "",
		"path to the predicate file.")
}

// PredicateRemoteOptions is the wrapper for remote predicate related options.
type PredicateRemoteOptions struct {
	PredicateOptions
}

var _ Interface = (*PredicateRemoteOptions)(nil)

// AddFlags implements Interface
func (o *PredicateRemoteOptions) AddFlags(cmd *cobra.Command) {
	o.PredicateOptions.AddFlags(cmd)
}
