// Copyright 2021 Google LLC All Rights Reserved.
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
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

const dateFormat = "2006-01-02T15:04:05Z"

func GenerateSPDX(koVersion string, date time.Time, mod []byte, imgDigest v1.Hash) ([]byte, error) {
	var err error
	mod, err = massageGoVersionM(mod)
	if err != nil {
		return nil, err
	}

	bi, err := ParseBuildInfo(string(mod))
	if err != nil {
		return nil, err
	}

	mainPackageID := "SPDXRef-Package-" + strings.ReplaceAll(bi.Main.Path, "/", ".")

	doc := Document{
		Version:           Version,
		DataLicense:       "CC0-1.0",
		ID:                "SPDXRef-DOCUMENT",
		Name:              bi.Main.Path,
		Namespace:         "http://spdx.org/spdxdocs/" + bi.Main.Path,
		DocumentDescribes: []string{mainPackageID},
		CreationInfo: CreationInfo{
			Created:  date.Format(dateFormat),
			Creators: []string{"Tool: ko " + koVersion},
		},
		Packages:      make([]Package, 0, 1+len(bi.Deps)),
		Relationships: make([]Relationship, 0, 1+len(bi.Deps)),
	}

	doc.Relationships = append(doc.Relationships, Relationship{
		Element: "SPDXRef-DOCUMENT",
		Type:    "DESCRIBES",
		Related: mainPackageID,
	})

	doc.Packages = append(doc.Packages, Package{
		Name: bi.Main.Path,
		ID:   mainPackageID,
		// TODO: PackageSupplier: "Organization: " + bs.Main.Path
		DownloadLocation: "https://" + bi.Main.Path,
		FilesAnalyzed:    false,
		// TODO: PackageHomePage:  "https://" + bi.Main.Path,
		LicenseConcluded: NOASSERTION,
		LicenseDeclared:  NOASSERTION,
		CopyrightText:    NOASSERTION,
		ExternalRefs: []ExternalRef{{
			Category: "PACKAGE_MANAGER",
			Type:     "purl",
			Locator:  ociRef(bi.Path, imgDigest),
		}},
	})

	for _, dep := range bi.Deps {
		depID := fmt.Sprintf("SPDXRef-Package-%s-%s",
			strings.ReplaceAll(dep.Path, "/", "."),
			dep.Version)

		doc.Relationships = append(doc.Relationships, Relationship{
			Element: mainPackageID,
			Type:    "DEPENDS_ON",
			Related: depID,
		})

		pkg := Package{
			Name:    dep.Path,
			ID:      depID,
			Version: dep.Version,
			// TODO: PackageSupplier: "Organization: " + dep.Path
			DownloadLocation: fmt.Sprintf("https://proxy.golang.org/%s/@v/%s.zip", dep.Path, dep.Version),
			FilesAnalyzed:    false,
			LicenseConcluded: NOASSERTION,
			LicenseDeclared:  NOASSERTION,
			CopyrightText:    NOASSERTION,
			ExternalRefs: []ExternalRef{{
				Category: "PACKAGE_MANAGER",
				Type:     "purl",
				Locator:  goRef(dep.Path, dep.Version),
			}},
		}

		if dep.Sum != "" {
			pkg.Checksums = []Checksum{{
				Algorithm: "SHA256",
				Value:     h1ToSHA256(dep.Sum),
			}}
		}

		doc.Packages = append(doc.Packages, pkg)
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(doc); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Below this is forked from here:
// https://github.com/kubernetes-sigs/bom/blob/main/pkg/spdx/json/v2.2.2/types.go

/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

const (
	NOASSERTION = "NOASSERTION"
	Version     = "SPDX-2.2"
)

type Document struct {
	ID                   string                `json:"SPDXID"`
	Name                 string                `json:"name"`
	Version              string                `json:"spdxVersion"`
	CreationInfo         CreationInfo          `json:"creationInfo"`
	DataLicense          string                `json:"dataLicense"`
	Namespace            string                `json:"documentNamespace"`
	DocumentDescribes    []string              `json:"documentDescribes,omitempty"`
	Files                []File                `json:"files,omitempty"`
	Packages             []Package             `json:"packages,omitempty"`
	Relationships        []Relationship        `json:"relationships,omitempty"`
	ExternalDocumentRefs []ExternalDocumentRef `json:"externalDocumentRefs,omitempty"`
}

type CreationInfo struct {
	Created            string   `json:"created"` // Date
	Creators           []string `json:"creators,omitempty"`
	LicenseListVersion string   `json:"licenseListVersion,omitempty"`
}

type Package struct {
	ID                   string                   `json:"SPDXID"`
	Name                 string                   `json:"name"`
	Version              string                   `json:"versionInfo"`
	FilesAnalyzed        bool                     `json:"filesAnalyzed"`
	LicenseDeclared      string                   `json:"licenseDeclared"`
	LicenseConcluded     string                   `json:"licenseConcluded"`
	Description          string                   `json:"description,omitempty"`
	DownloadLocation     string                   `json:"downloadLocation"`
	Originator           string                   `json:"originator,omitempty"`
	SourceInfo           string                   `json:"sourceInfo,omitempty"`
	CopyrightText        string                   `json:"copyrightText"`
	HasFiles             []string                 `json:"hasFiles,omitempty"`
	LicenseInfoFromFiles []string                 `json:"licenseInfoFromFiles,omitempty"`
	Checksums            []Checksum               `json:"checksums,omitempty"`
	ExternalRefs         []ExternalRef            `json:"externalRefs,omitempty"`
	VerificationCode     *PackageVerificationCode `json:"packageVerificationCode,omitempty"`
}

type PackageVerificationCode struct {
	Value         string   `json:"packageVerificationCodeValue"`
	ExcludedFiles []string `json:"packageVerificationCodeExcludedFiles,omitempty"`
}

type File struct {
	ID                string     `json:"SPDXID"`
	Name              string     `json:"fileName"`
	CopyrightText     string     `json:"copyrightText"`
	NoticeText        string     `json:"noticeText,omitempty"`
	LicenseConcluded  string     `json:"licenseConcluded"`
	Description       string     `json:"description,omitempty"`
	FileTypes         []string   `json:"fileTypes,omitempty"`
	LicenseInfoInFile []string   `json:"licenseInfoInFiles"` // List of licenses
	Checksums         []Checksum `json:"checksums"`
}

type Checksum struct {
	Algorithm string `json:"algorithm"`
	Value     string `json:"checksumValue"`
}

type ExternalRef struct {
	Category string `json:"referenceCategory"`
	Locator  string `json:"referenceLocator"`
	Type     string `json:"referenceType"`
}

type Relationship struct {
	Element string `json:"spdxElementId"`
	Type    string `json:"relationshipType"`
	Related string `json:"relatedSpdxElement"`
}

type ExternalDocumentRef struct {
	Checksum           Checksum `json:"checksum"`
	ExternalDocumentID string   `json:"externalDocumentId"`
	SPDXDocument       string   `json:"spdxDocument"`
}
