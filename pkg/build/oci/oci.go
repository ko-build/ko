package build

import (
	"fmt"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/match"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

// implementation from github.com/google/go-containerregistry/pull/1293
// use upstream once merged

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
      // TODO: use tarball.LayerFromOpener, using nopCloser and reader doesn't work.
      // Reader returns eof when trying to read from layer.uncompressedopener
      // Ref: https://github.com/google/go-containerregistry/blob/4d7b65b04609719eb0f23afa8669ba4b47178571/pkg/v1/tarball/layer.go#L253
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
    return nil, fmt.Errorf("getting image layers: %w", err)
	}

	layers, err = convertLayers(layers)
	if err != nil {
    return nil, fmt.Errorf("converting layers: %w", err)
	}

	base, err = mutate.AppendLayers(empty.Image, layers...)
	if err != nil {
    return nil, fmt.Errorf("appending layers to empty.Image: %w", err)
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

func ConvertImageIndex(base v1.ImageIndex) (v1.ImageIndex, error) {
	base = mutate.IndexMediaType(base, types.OCIImageIndex)
	mn, err := base.IndexManifest()
	if err != nil {
		return nil, err
	}

	removals := []v1.Hash{}
	addendums := []mutate.IndexAddendum{}

	for _, manifest := range mn.Manifests {
		if !manifest.MediaType.IsImage() {
			// it is not an image, leave it as is
			continue
		}
		img, err := base.Image(manifest.Digest)
		if err != nil {
			return nil, err
		}
		img, err = ConvertImage(img)
		if err != nil {
			return nil, err
		}
		mt, err := img.MediaType()
		if err != nil {
			return nil, err
		}
		removals = append(removals, manifest.Digest)
		addendums = append(addendums, mutate.IndexAddendum{Add: img, Descriptor: v1.Descriptor{
			URLs:        manifest.URLs,
			MediaType:   mt,
			Annotations: manifest.Annotations,
			Platform:    manifest.Platform,
		}})
	}
	base = mutate.RemoveManifests(base, match.Digests(removals...))
	base = mutate.AppendManifests(base, addendums...)
	return base, nil
}
