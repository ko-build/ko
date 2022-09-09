package build

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/ko/internal/ociconv"
)

const wasmImageAnnotationKey = "module.wasm.image/variant"

func isWasi(p *v1.Platform) bool {
	return p.OS == wasiPlatform.OS && p.Architecture == wasiPlatform.Architecture
}

var wasiPlatform = &v1.Platform{
	OS:           "wasm",
	Architecture: "wasi",
}

type tinygoExecutableBuilder struct {
	dir string
}

func (b *tinygoExecutableBuilder) BuildExecutable(ctx context.Context, ip string, config Config) (string, error) {
	tmpDir, err := createExecutableTempDir(ip, wasiPlatform)
	if err != nil {
		return "", fmt.Errorf("could not create tempdir for %s: %w", ip, err)
	}
	file := filepath.Join(tmpDir, "out")
	args := []string{
		"build",
		"-o",
		file,
		"-target=wasi",
		ip,
	}
	cmd := exec.CommandContext(ctx, "tinygo", args...)

	var output bytes.Buffer
	cmd.Stderr = &output
	cmd.Stdout = &output

	log.Printf("Building %s for %s", ip, wasiPlatform)
	if err := cmd.Run(); err != nil {
		if os.Getenv("KOCACHE") == "" {
			os.RemoveAll(tmpDir)
		}
		log.Printf("Unexpected error running \"tinygo build\": %v\n%v", err, output.String())
		return "", err
	}
	return file, nil
}

func appendWasiExtension(appPath string) string {
	return appPath + ".wasm"
}

func annotateImageForWasi(image v1.Image) (v1.Image, error) {
	image, err := ociconv.OCIImage(image)
	if err != nil {
		return nil, fmt.Errorf("converting image to oci: %w", err)
	}
	image = mutate.Annotations(image, map[string]string{
		wasmImageAnnotationKey: "compat-smart",
	}).(v1.Image)
	return image, nil
}
