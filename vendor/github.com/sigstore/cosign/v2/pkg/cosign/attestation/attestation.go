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

package attestation

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	slsa "github.com/in-toto/in-toto-golang/in_toto/slsa_provenance/v0.2"

	"github.com/in-toto/in-toto-golang/in_toto"
)

const (
	// CosignCustomProvenanceV01 specifies the type of the Predicate.
	CosignCustomProvenanceV01 = "cosign.sigstore.dev/attestation/v1"

	// CosignVulnProvenanceV01 specifies the type of VulnerabilityScan Predicate
	CosignVulnProvenanceV01 = "cosign.sigstore.dev/attestation/vuln/v1"
)

// CosignPredicate specifies the format of the Custom Predicate.
type CosignPredicate struct {
	Data      interface{}
	Timestamp string
}

// VulnPredicate specifies the format of the Vulnerability Scan Predicate
type CosignVulnPredicate struct {
	Invocation Invocation `json:"invocation"`
	Scanner    Scanner    `json:"scanner"`
	Metadata   Metadata   `json:"metadata"`
}

// I think this will be moving to upstream in-toto in the fullness of time
// but creating it here for now so that we have a way to deserialize it
// as a InToto Statement
// https://github.com/in-toto/attestation/issues/58
type CosignVulnStatement struct {
	in_toto.StatementHeader
	Predicate CosignVulnPredicate `json:"predicate"`
}

type Invocation struct {
	Parameters interface{} `json:"parameters"`
	URI        string      `json:"uri"`
	EventID    string      `json:"event_id"`
	BuilderID  string      `json:"builder.id"`
}

type DB struct {
	URI     string `json:"uri"`
	Version string `json:"version"`
}

type Scanner struct {
	URI     string      `json:"uri"`
	Version string      `json:"version"`
	DB      DB          `json:"db"`
	Result  interface{} `json:"result"`
}

type Metadata struct {
	ScanStartedOn  time.Time `json:"scanStartedOn"`
	ScanFinishedOn time.Time `json:"scanFinishedOn"`
}

// GenerateOpts specifies the options of the Statement generator.
type GenerateOpts struct {
	// Predicate is the source of bytes (e.g. a file) to use as the statement's predicate.
	Predicate io.Reader
	// Type is the pre-defined enums (provenance|link|spdx).
	// default: custom
	Type string
	// Digest of the Image reference.
	Digest string
	// Repo context of the reference.
	Repo string

	// Function to return the time to set
	Time func() time.Time
}

// GenerateStatement returns an in-toto statement based on the provided
// predicate type (custom|slsaprovenance|spdx|spdxjson|cyclonedx|link).
func GenerateStatement(opts GenerateOpts) (interface{}, error) {
	predicate, err := io.ReadAll(opts.Predicate)
	if err != nil {
		return nil, err
	}

	switch opts.Type {
	case "slsaprovenance":
		return generateSLSAProvenanceStatement(predicate, opts.Digest, opts.Repo)
	case "spdx":
		return generateSPDXStatement(predicate, opts.Digest, opts.Repo, false)
	case "spdxjson":
		return generateSPDXStatement(predicate, opts.Digest, opts.Repo, true)
	case "cyclonedx":
		return generateCycloneDXStatement(predicate, opts.Digest, opts.Repo)
	case "link":
		return generateLinkStatement(predicate, opts.Digest, opts.Repo)
	case "vuln":
		return generateVulnStatement(predicate, opts.Digest, opts.Repo)
	default:
		stamp := timestamp(opts)
		predicateType := customType(opts)
		return generateCustomStatement(predicate, predicateType, opts.Digest, opts.Repo, stamp)
	}
}

func generateVulnStatement(predicate []byte, digest string, repo string) (interface{}, error) {
	var vuln CosignVulnPredicate

	err := json.Unmarshal(predicate, &vuln)
	if err != nil {
		return nil, err
	}

	return in_toto.Statement{
		StatementHeader: generateStatementHeader(digest, repo, CosignVulnProvenanceV01),
		Predicate:       vuln,
	}, nil
}

func timestamp(opts GenerateOpts) string {
	if opts.Time == nil {
		opts.Time = time.Now
	}
	now := opts.Time()
	return now.UTC().Format(time.RFC3339)
}

func customType(opts GenerateOpts) string {
	if opts.Type != "custom" {
		return opts.Type
	}
	return CosignCustomProvenanceV01
}

func generateStatementHeader(digest, repo, predicateType string) in_toto.StatementHeader {
	return in_toto.StatementHeader{
		Type:          in_toto.StatementInTotoV01,
		PredicateType: predicateType,
		Subject: []in_toto.Subject{
			{
				Name: repo,
				Digest: map[string]string{
					"sha256": digest,
				},
			},
		},
	}
}

func generateCustomStatement(rawPayload []byte, customType, digest, repo, timestamp string) (interface{}, error) {
	payload, err := generateCustomPredicate(rawPayload, customType, timestamp)
	if err != nil {
		return nil, err
	}

	return in_toto.Statement{
		StatementHeader: generateStatementHeader(digest, repo, customType),
		Predicate:       payload,
	}, nil
}

func generateCustomPredicate(rawPayload []byte, customType, timestamp string) (interface{}, error) {
	if customType == CosignCustomProvenanceV01 {
		return &CosignPredicate{
			Data:      string(rawPayload),
			Timestamp: timestamp,
		}, nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(rawPayload, &result); err != nil {
		return nil, fmt.Errorf("invalid JSON payload for predicate type %s: %w", customType, err)
	}

	return result, nil
}

func generateSLSAProvenanceStatement(rawPayload []byte, digest string, repo string) (interface{}, error) {
	var predicate slsa.ProvenancePredicate
	err := checkRequiredJSONFields(rawPayload, reflect.TypeOf(predicate))
	if err != nil {
		return nil, fmt.Errorf("provenance predicate: %w", err)
	}
	err = json.Unmarshal(rawPayload, &predicate)
	if err != nil {
		return "", fmt.Errorf("unmarshal Provenance predicate: %w", err)
	}
	return in_toto.ProvenanceStatement{
		StatementHeader: generateStatementHeader(digest, repo, slsa.PredicateSLSAProvenance),
		Predicate:       predicate,
	}, nil
}

func generateLinkStatement(rawPayload []byte, digest string, repo string) (interface{}, error) {
	var link in_toto.Link
	err := checkRequiredJSONFields(rawPayload, reflect.TypeOf(link))
	if err != nil {
		return nil, fmt.Errorf("link statement: %w", err)
	}
	err = json.Unmarshal(rawPayload, &link)
	if err != nil {
		return "", fmt.Errorf("unmarshal Link statement: %w", err)
	}
	return in_toto.LinkStatement{
		StatementHeader: generateStatementHeader(digest, repo, in_toto.PredicateLinkV1),
		Predicate:       link,
	}, nil
}

func generateSPDXStatement(rawPayload []byte, digest string, repo string, parseJSON bool) (interface{}, error) {
	var data interface{}
	if parseJSON {
		if err := json.Unmarshal(rawPayload, &data); err != nil {
			return nil, err
		}
	} else {
		data = string(rawPayload)
	}
	return in_toto.SPDXStatement{
		StatementHeader: generateStatementHeader(digest, repo, in_toto.PredicateSPDX),
		Predicate: CosignPredicate{
			Data: data,
		},
	}, nil
}

func generateCycloneDXStatement(rawPayload []byte, digest string, repo string) (interface{}, error) {
	var data interface{}
	if err := json.Unmarshal(rawPayload, &data); err != nil {
		return nil, err
	}
	return in_toto.SPDXStatement{
		StatementHeader: generateStatementHeader(digest, repo, in_toto.PredicateCycloneDX),
		Predicate: CosignPredicate{
			Data: data,
		},
	}, nil
}

func checkRequiredJSONFields(rawPayload []byte, typ reflect.Type) error {
	var tmp map[string]interface{}
	if err := json.Unmarshal(rawPayload, &tmp); err != nil {
		return err
	}
	// Create list of json tags, e.g. `json:"_type"`
	attributeCount := typ.NumField()
	allFields := make([]string, 0)
	for i := 0; i < attributeCount; i++ {
		jsonTagFields := strings.SplitN(typ.Field(i).Tag.Get("json"), ",", 2)
		if len(jsonTagFields) < 2 {
			allFields = append(allFields, jsonTagFields[0])
		}
	}

	// Assert that there's a key in the passed map for each tag
	for _, field := range allFields {
		if _, ok := tmp[field]; !ok {
			return fmt.Errorf("required field %s missing", field)
		}
	}
	return nil
}
