// Copyright 2022 Chainguard, Inc.
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

package cyclonedx

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	purl "github.com/package-url/packageurl-go"

	"chainguard.dev/apko/pkg/sbom/options"
)

type CycloneDX struct{}

func New() CycloneDX {
	return CycloneDX{}
}

func (cdx *CycloneDX) Key() string {
	return "cyclonedx"
}

func (cdx *CycloneDX) Ext() string {
	return "cdx"
}

// Generate writes a CycloneDX sbom in path
func (cdx *CycloneDX) Generate(opts *options.Options, path string) error {
	pkgComponents := []Component{}
	pkgDependencies := []Dependency{}

	mm := map[string]string{"arch": opts.ImageInfo.Arch.ToAPK()}

	for _, pkg := range opts.Packages {
		// add the component
		c := Component{
			BOMRef: purl.NewPackageURL(
				"apk", opts.OS.ID, pkg.Name, pkg.Version,
				purl.QualifiersFromMap(mm), "").String(),
			Name:        pkg.Name,
			Version:     pkg.Version,
			Description: pkg.Description,
			Licenses: []License{
				{
					Expression: pkg.License,
				},
			},
			PUrl: purl.NewPackageURL(
				"apk", opts.OS.ID, pkg.Name, pkg.Version,
				purl.QualifiersFromMap(mm), "").String(),
			// TODO(kaniini): Talk with CycloneDX people about adding "package" type.
			Type: "operating-system",
		}

		pkgComponents = append(pkgComponents, c)

		// walk the dependency list
		depRefs := []string{}
		for _, dep := range pkg.Dependencies {
			// TODO(kaniini): Properly handle virtual dependencies...
			if strings.ContainsRune(dep, ':') {
				continue
			}

			i := strings.IndexAny(dep, " ~<>=/!")
			if i > -1 {
				dep = dep[:i]
			}
			if dep == "" {
				continue
			}

			depRefs = append(depRefs, purl.NewPackageURL("apk", opts.OS.ID, dep, "",
				purl.QualifiersFromMap(mm), "").String())
		}

		d := Dependency{
			Ref: purl.NewPackageURL(
				"apk", opts.OS.ID, pkg.Name, pkg.Version,
				purl.QualifiersFromMap(mm), "").String(),
			DependsOn: depRefs,
		}
		pkgDependencies = append(pkgDependencies, d)
	}

	// Main package purl qualifiers
	var imageComponent Component
	layerComponent := Component{
		BOMRef: purl.NewPackageURL(
			purl.TypeOCI, "", opts.ImagePurlName(), opts.ImageInfo.LayerDigest,
			nil, "",
		).String() + "?" + opts.ImagePurlQualifiers().String(),
		Name:        opts.OS.Name,
		Description: "apko OS layer",
		PUrl: purl.NewPackageURL(
			purl.TypeOCI, "", opts.ImagePurlName(), opts.ImageInfo.LayerDigest,
			nil, "",
		).String() + "?" + opts.LayerPurlQualifiers().String(),
		Version:    opts.OS.Version,
		Type:       "operating-system",
		Components: pkgComponents,
	}

	if opts.ImageInfo.ImageDigest != "" {
		imageComponent = Component{
			BOMRef: purl.NewPackageURL(
				purl.TypeOCI, "", opts.ImagePurlName(), opts.ImageInfo.ImageDigest,
				nil, "",
			).String() + "?" + opts.ImagePurlQualifiers().String(),
			Type: "container",
			Name: "",
			// Version:            "",
			Description: "apko container image",
			PUrl: purl.NewPackageURL(
				purl.TypeOCI, "", opts.ImagePurlName(), opts.ImageInfo.ImageDigest,
				nil, "",
			).String() + "?" + opts.ImagePurlQualifiers().String(),
			Components: []Component{layerComponent},
		}
	}

	if opts.ImageInfo.VCSUrl != "" {
		vcsURLRef := ExternalReference{
			URL:  opts.ImageInfo.VCSUrl,
			Type: "vcs",
		}

		if opts.ImageInfo.ImageDigest != "" {
			imageComponent.ExternalReferences = append(imageComponent.ExternalReferences, vcsURLRef)
		} else {
			layerComponent.ExternalReferences = append(layerComponent.ExternalReferences, vcsURLRef)
		}
	}

	bom := Document{
		BOMFormat:    "CycloneDX",
		SpecVersion:  "1.4",
		Version:      1,
		Dependencies: pkgDependencies,
	}

	if opts.ImageInfo.ImageDigest != "" {
		bom.Components = []Component{imageComponent}
	} else {
		bom.Components = []Component{layerComponent}
	}

	if err := renderDoc(&bom, path); err != nil {
		return fmt.Errorf("rendering sbom to disk: %w", err)
	}

	return nil
}

// TODO(kaniini): Move most of this over to gitlab.alpinelinux.org/alpine/go.
type Document struct {
	BOMFormat    string       `json:"bomFormat"`
	SpecVersion  string       `json:"specVersion"`
	Version      int          `json:"version"`
	Components   []Component  `json:"components,omitempty"`
	Dependencies []Dependency `json:"dependencies,omitempty"`
}

type Component struct {
	BOMRef             string              `json:"bom-ref"`
	Type               string              `json:"type"`
	Name               string              `json:"name"`
	Version            string              `json:"version"`
	Description        string              `json:"description"`
	PUrl               string              `json:"purl"`
	Hashes             []Hash              `json:"hashes,omitempty"`
	ExternalReferences []ExternalReference `json:"externalReferences,omitempty"`
	Licenses           []License           `json:"licenses,omitempty"`
	Components         []Component         `json:"components,omitempty"`
}

type License struct {
	Expression string `json:"expression"`
}

type ExternalReference struct {
	URL  string `json:"url"`
	Type string `json:"type"`
}

type Dependency struct {
	Ref       string   `json:"ref"`
	DependsOn []string `json:"dependsOn"`
}

type HashAlgorithm string

type Hash struct {
	Algorithm HashAlgorithm `json:"alg"`
	Value     string        `json:"content"`
}

func (cdx *CycloneDX) GenerateIndex(opts *options.Options, path string) error {
	purlString := purl.NewPackageURL(
		purl.TypeOCI, "", opts.IndexPurlName(), opts.ImageInfo.IndexDigest.String(),
		nil, "",
	).String() + "?" + opts.IndexPurlQualifiers().String()

	indexComponent := Component{
		BOMRef:      purlString,
		Type:        "container",
		Name:        opts.ImageInfo.IndexDigest.String(),
		Version:     opts.ImageInfo.IndexDigest.DeepCopy().Hex,
		Description: "Multi-arch image index",
		PUrl:        purlString,
		Hashes: []Hash{
			{
				Algorithm: "SHA-256",
				Value:     opts.ImageInfo.IndexDigest.DeepCopy().Hex,
			},
		},
		Components: []Component{},
	}

	if opts.ImageInfo.VCSUrl != "" {
		indexComponent.ExternalReferences = append(indexComponent.ExternalReferences, ExternalReference{
			URL:  opts.ImageInfo.VCSUrl,
			Type: "vcs",
		})
	}

	// Add the images as subcomponents
	for _, info := range opts.ImageInfo.Images {
		indexComponent.Components = append(
			indexComponent.Components, cdx.archImageComponent(opts, info),
		)
	}

	bom := Document{
		BOMFormat:   "CycloneDX",
		SpecVersion: "1.4",
		Version:     1,
		Components: []Component{
			indexComponent,
		},
		Dependencies: []Dependency{},
	}

	if err := renderDoc(&bom, path); err != nil {
		return fmt.Errorf("rendering SBOM: %w", err)
	}

	return nil
}

// imageComponent takes an image and returns a component representing it
func (cdx *CycloneDX) archImageComponent(opts *options.Options, info options.ArchImageInfo) Component {
	purlString := purl.NewPackageURL(
		purl.TypeOCI, "", opts.ImagePurlName(), info.Digest.DeepCopy().String(),
		nil, "",
	).String() + "?" + opts.ArchImagePurlQualifiers(&info).String()

	return Component{
		BOMRef: purlString,
		Type:   "container",
		Name:   info.Digest.DeepCopy().String(),
		Description: fmt.Sprintf(
			"apko image for %s/%s", info.Arch.ToOCIPlatform().OS, info.Arch,
		),
		PUrl:    purlString,
		Version: info.Digest.DeepCopy().String(),
		Hashes: []Hash{
			{
				Algorithm: "SHA-256",
				Value:     info.Digest.DeepCopy().Hex,
			},
		},
		ExternalReferences: []ExternalReference{},
		Licenses:           []License{},
		Components:         []Component{},
	}
}

// renderDoc marshals a document to json and writes it to disk
func renderDoc(doc *Document, path string) error {
	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("opening SBOM path %s for writing: %w", path, err)
	}
	defer out.Close()

	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)

	if err := enc.Encode(doc); err != nil {
		return fmt.Errorf("encoding spdx sbom: %w", err)
	}
	return nil
}
