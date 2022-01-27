package sbom

import (
	"bytes"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/CycloneDX/cyclonedx-gomod/pkg/generate/bin"
	"github.com/rs/zerolog"
)

func GenerateCycloneDX(binaryPath string) ([]byte, error) {
	gen, err := bin.NewGenerator(binaryPath, bin.WithLogger(zerolog.Nop()))
	if err != nil {
		return nil, err
	}
	bom, err := gen.Generate()
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	enc := cdx.NewBOMEncoder(&buf, cdx.BOMFileFormatJSON)
	enc.SetPretty(true)
	if err := enc.Encode(bom); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
