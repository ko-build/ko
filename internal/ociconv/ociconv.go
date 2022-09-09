// from https://github.com/google/go-containerregistry/blob/53739b507dcc56b7ff29ee17982d6c5179b77aaa/pkg/v1/mutate/oci.go

package ociconv

import (
	"fmt"

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

// OCIImage mutates the provided v1.Image to be OCI compliant v1.Image
// Check image-spec to see which properties are ported and which are dropped.
// https://github.com/opencontainers/image-spec/blob/main/config.md
func OCIImage(base v1.Image) (v1.Image, error) {
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
			layer, err = tarball.LayerFromReader(reader, tarball.WithMediaType(types.OCILayer))
			if err != nil {
				return nil, fmt.Errorf("building layer: %w", err)
			}
		case types.DockerUncompressedLayer:
			reader, err := layer.Uncompressed()
			if err != nil {
				return nil, fmt.Errorf("getting layer: %w", err)
			}
			layer, err = tarball.LayerFromReader(reader, tarball.WithMediaType(types.OCIUncompressedLayer))
			if err != nil {
				return nil, fmt.Errorf("building layer: %w", err)
			}
		}
		newLayers = append(newLayers, layer)
	}

	base, err = mutate.AppendLayers(empty.Image, newLayers...)
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
