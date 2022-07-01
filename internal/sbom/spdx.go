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
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/sigstore/cosign/pkg/oci"
)

const dateFormat = "2006-01-02T15:04:05Z"

func GenerateImageSPDX(koVersion string, mod []byte, img oci.SignedImage) ([]byte, error) {
	var err error
	mod, err = massageGoVersionM(mod)
	if err != nil {
		return nil, err
	}

	bi, err := ParseBuildInfo(string(mod))
	if err != nil {
		return nil, err
	}

	imgDigest, err := img.Digest()
	if err != nil {
		return nil, err
	}
	cfg, err := img.ConfigFile()
	if err != nil {
		return nil, err
	}

	doc, imageID := starterDocument(koVersion, cfg.Created.Time, imgDigest)

	// image -> main package -> transitive deps
	doc.Packages = make([]Package, 0, 2+len(bi.Deps))
	doc.Relationships = make([]Relationship, 0, 2+len(bi.Deps))

	doc.Relationships = append(doc.Relationships, Relationship{
		Element: "SPDXRef-DOCUMENT",
		Type:    "DESCRIBES",
		Related: imageID,
	})

	doc.Packages = append(doc.Packages, Package{
		ID:   imageID,
		Name: imgDigest.String(),
		// TODO: PackageSupplier: "Organization: " + bs.Main.Path
		FilesAnalyzed: false,
		// TODO: PackageHomePage:  "https://" + bi.Main.Path,
		LicenseConcluded: NOASSERTION,
		LicenseDeclared:  NOASSERTION,
		CopyrightText:    NOASSERTION,
		ExternalRefs: []ExternalRef{{
			Category: "PACKAGE_MANAGER",
			Type:     "purl",
			Locator:  ociRef("image", imgDigest),
		}},
	})

	mainPackageID := modulePackageName(&bi.Main)

	doc.Relationships = append(doc.Relationships, Relationship{
		Element: imageID,
		Type:    "CONTAINS",
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
			Locator:  goRef(&bi.Main),
		}},
	})

	for _, dep := range bi.Deps {
		depID := modulePackageName(dep)

		doc.Relationships = append(doc.Relationships, Relationship{
			Element: mainPackageID,
			Type:    "DEPENDS_ON",
			Related: depID,
		})

		pkg := Package{
			ID:      depID,
			Name:    dep.Path,
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
				Locator:  goRef(dep),
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

func extractDate(sii oci.SignedImageIndex) (*time.Time, error) {
	im, err := sii.IndexManifest()
	if err != nil {
		return nil, err
	}
	for _, desc := range im.Manifests {
		switch desc.MediaType {
		case types.OCIManifestSchema1, types.DockerManifestSchema2:
			si, err := sii.SignedImage(desc.Digest)
			if err != nil {
				return nil, err
			}
			cfg, err := si.ConfigFile()
			if err != nil {
				return nil, err
			}
			return &cfg.Created.Time, nil

		default:
			// We shouldn't need to handle nested indices, since we don't build
			// them, but if we do we will need to do some sort of recursion here.
			return nil, fmt.Errorf("unknown media type: %v", desc.MediaType)
		}
	}
	return nil, errors.New("unable to extract date, no imaged found")
}

func GenerateIndexSPDX(koVersion string, sii oci.SignedImageIndex) ([]byte, error) {
	d, err := sii.Digest()
	if err != nil {
		return nil, err
	}

	date, err := extractDate(sii)
	if err != nil {
		return nil, err
	}

	doc, indexID := starterDocument(koVersion, *date, d)
	doc.Packages = []Package{{
		ID:               indexID,
		Name:             d.String(),
		FilesAnalyzed:    false,
		LicenseConcluded: NOASSERTION,
		LicenseDeclared:  NOASSERTION,
		CopyrightText:    NOASSERTION,
		Checksums: []Checksum{{
			Algorithm: strings.ToUpper(d.Algorithm),
			Value:     d.Hex,
		}},
		ExternalRefs: []ExternalRef{{
			Category: "PACKAGE_MANAGER",
			Type:     "purl",
			Locator:  ociRef("index", d),
		}},
	}}

	im, err := sii.IndexManifest()
	if err != nil {
		return nil, err
	}
	for _, desc := range im.Manifests {
		switch desc.MediaType {
		case types.OCIManifestSchema1, types.DockerManifestSchema2:
			si, err := sii.SignedImage(desc.Digest)
			if err != nil {
				return nil, err
			}

			f, err := si.Attachment("sbom")
			if err != nil {
				return nil, err
			}
			sd, err := f.Digest()
			if err != nil {
				return nil, err
			}

			// TODO: This should capture the ENTIRE desc.Platform, but I am
			// starting with just architecture.
			docID := fmt.Sprintf("DocumentRef-%s-image-sbom", desc.Platform.Architecture)

			// This must match the name of the package in the image's SBOM.
			spdxID := ociPackageName(desc.Digest)

			doc.Relationships = append(doc.Relationships, Relationship{
				Element: fmt.Sprintf("%s:%s", docID, spdxID),
				Type:    "VARIANT_OF",
				Related: indexID,
			})

			doc.ExternalDocumentRefs = append(doc.ExternalDocumentRefs, ExternalDocumentRef{
				Checksum: Checksum{
					Algorithm: strings.ToUpper(sd.Algorithm),
					Value:     sd.Hex,
				},
				ExternalDocumentID: docID,
				// This isn't quite right, but what we are trying to convey here
				// is that the external document being referenced is the SBOM
				// associated with `desc.Digest` with the checksum expressed
				// above.  This will result in a pURL like:
				//    pkg:oci/sbom@sha256:{image digest}
				// We should look for this SBOM alongside the image with tag:
				//    :sha256-{image digest}.sbom
				// ... and the SBOM itself will have the digest presented in the
				// checksum field above.
				SPDXDocument: ociRef("sbom", desc.Digest),
			})

		default:
			// We shouldn't need to handle nested indices, since we don't build
			// them, but if we do we will need to do some sort of recursion here.
			return nil, fmt.Errorf("unknown media type: %v", desc.MediaType)
		}
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(doc); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func ociPackageName(d v1.Hash) string {
	return fmt.Sprintf("SPDXRef-Package-%s-%s", d.Algorithm, d.Hex)
}

func modulePackageName(mod *debug.Module) string {
	return fmt.Sprintf("SPDXRef-Package-%s-%s",
		strings.ReplaceAll(mod.Path, "/", "."),
		mod.Version)
}

func starterDocument(koVersion string, date time.Time, d v1.Hash) (Document, string) {
	digestID := ociPackageName(d)
	return Document{
		ID:      "SPDXRef-DOCUMENT",
		Version: Version,
		CreationInfo: CreationInfo{
			Created:  date.Format(dateFormat),
			Creators: []string{"Tool: ko " + koVersion},
		},
		DataLicense:       "CC0-1.0",
		Name:              "sbom-" + d.String(),
		Namespace:         "http://spdx.org/spdxdocs/ko" + d.String(),
		DocumentDescribes: []string{digestID},
	}, digestID
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

type ExternalDocumentRef struct {
	Checksum           Checksum `json:"checksum"`
	ExternalDocumentID string   `json:"externalDocumentId"`
	SPDXDocument       string   `json:"spdxDocument"`
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
