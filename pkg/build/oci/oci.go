package build

import (
	"fmt"
	"io"
	"io/ioutil"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

// Drops docker specific properties
// See: https://github.com/opencontainers/image-spec/blob/main/config.md
func toOCIV1Config(config v1.Config) v1.Config {
	return v1.Config{
		User:         config.User,
		ExposedPorts: config.ExposedPorts,
		Env:          config.Env,
		Entrypoint:   config.Entrypoint,
		Cmd:          config.Cmd,
		Volumes:      config.Volumes,
		WorkingDir:   config.WorkingDir,
		Labels:       config.Labels,
		StopSignal:   config.StopSignal,
	}
}

func toOCIV1ConfigFile(cf *v1.ConfigFile) *v1.ConfigFile {
	return &v1.ConfigFile{
		Created:      cf.Created,
		Author:       cf.Author,
		Architecture: cf.Architecture,
		OS:           cf.OS,
		OSVersion:    cf.OSVersion,
		History:      cf.History,
		RootFS:       cf.RootFS,
		Config:       toOCIV1Config(cf.Config),
	}
}

func convertLayers(layers []v1.Layer) ([]v1.Layer, error) {
	newLayers := []v1.Layer{}
	for _, layer := range layers {
		mediaType, err := layer.MediaType()
		if err != nil {
			return nil, err
		}
		switch mediaType {
		case types.DockerLayer:
			reader, err := layer.Compressed()
			if err != nil {
				return nil, fmt.Errorf("getting layer: %w", err)
			}
			layer, err = tarball.LayerFromOpener(func() (io.ReadCloser, error) {
				return ioutil.NopCloser(reader), nil
			}, tarball.WithMediaType(types.OCILayer))
			if err != nil {
				return nil, fmt.Errorf("building layer: %w", err)
			}
		case types.DockerUncompressedLayer:
			reader, err := layer.Uncompressed()
			if err != nil {
				return nil, fmt.Errorf("getting layer: %w", err)
			}
			layer, err = tarball.LayerFromOpener(func() (io.ReadCloser, error) {
				return ioutil.NopCloser(reader), nil
			}, tarball.WithMediaType(types.OCIUncompressedLayer))
			if err != nil {
				return nil, fmt.Errorf("building layer: %w", err)
			}
		}
		newLayers = append(newLayers, layer)
	}
	return newLayers, nil
}

// ConvertImage performs convertion of base to oci spec.
// Check image-spec to see which properties are ported and which are dropped.
// https://github.com/opencontainers/image-spec/blob/main/config.md
func ConvertImage(base v1.Image) (v1.Image, error) {
	// Get original manifest
	m, err := base.Manifest()
	if err != nil {
		return nil, err
	}
	// Convert config
	cfg, err := base.ConfigFile()
	if err != nil {
		return nil, err
	}
	cfg = toOCIV1ConfigFile(cfg)

	layers, err := base.Layers()
	if err != nil {
		return nil, err
	}

	layers, err = convertLayers(layers)
	if err != nil {
		return nil, err
	}

	base, err = mutate.AppendLayers(empty.Image, layers...)
	if err != nil {
		return nil, err
	}

	base = mutate.MediaType(base, types.OCIManifestSchema1)
	base = mutate.ConfigMediaType(base, types.OCIConfigJSON)
	base = mutate.Annotations(base, m.Annotations).(v1.Image)
	base, err = mutate.ConfigFile(base, cfg)
	if err != nil {
		return nil, err
	}
	return base, nil
}
