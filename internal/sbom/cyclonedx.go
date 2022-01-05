package sbom

import (
	"bytes"
	"github.com/CycloneDX/cyclonedx-go"
	"github.com/google/uuid"
	"time"
)

func GenerateCycloneDX(koVersion string, date time.Time, mod []byte) ([]byte, error) {
	bi := &BuildInfo{}
	if err := bi.UnmarshalText(mod); err != nil {
		return nil, err
	}

	bom := cyclonedx.NewBOM()
	bom.SerialNumber = uuid.New().URN()
	bom.Metadata = &cyclonedx.Metadata{}
	bom.Metadata.Timestamp = time.Now().Format(time.RFC3339)
	bom.Metadata.Tools = &[]cyclonedx.Tool{{
		Vendor:  "ko " + koVersion,
		Name:    "ko",
		Version: koVersion,
	}}

	deps := make([]cyclonedx.Dependency, len(bi.Deps))
	for _, d := range bi.Deps {
		dependency := cyclonedx.Dependency{
			Ref: d.Path,
		}
		deps = append(deps, dependency)
	}

	bom.Dependencies = &deps

	buf := new(bytes.Buffer)
	be := cyclonedx.NewBOMEncoder(buf, cyclonedx.BOMFileFormatXML)

	err := be.Encode(bom)

	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
