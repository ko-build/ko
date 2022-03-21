package publish

import (
	"context"
	"fmt"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/ko/pkg/build"
	"github.com/google/ko/pkg/publish/k3s"
	"log"
	"os"
	"strings"
)

const (
	//K3sDomain   k3s local sentinel registry where the images gets loaded
	// TODO(kamesh) handle namespaces as part of URI ??
	K3sDomain = "k3s.local"
)

type k3sPublisher struct {
	namer Namer
	tags  []string
}

//NewK3sPublisher returns a new publish.Interface that loads image into k3s clusters
func NewK3sPublisher(namer Namer, tags []string) Interface {
	return &k3sPublisher{
		namer: namer,
		tags:  tags,
	}
}

//Publish implements publish.Interface
func (k *k3sPublisher) Publish(ctx context.Context, br build.Result, s string) (name.Reference, error) {
	s = strings.TrimPrefix(s, build.StrictScheme)
	s = strings.ToLower(s)

	// There's no way to write an index to a kind, so attempt to downcast it to an image.
	var img v1.Image
	switch i := br.(type) {
	case v1.Image:
		img = i
	case v1.ImageIndex:
		im, err := i.IndexManifest()
		if err != nil {
			return nil, err
		}
		goos, goarch := os.Getenv("GOOS"), os.Getenv("GOARCH")
		if goos == "" {
			goos = "linux"
		}
		if goarch == "" {
			goarch = "amd64"
		}
		for _, manifest := range im.Manifests {
			if manifest.Platform == nil {
				continue
			}
			if manifest.Platform.OS != goos {
				continue
			}
			if manifest.Platform.Architecture != goarch {
				continue
			}
			img, err = i.Image(manifest.Digest)
			if err != nil {
				return nil, err
			}
			break
		}
		if img == nil {
			return nil, fmt.Errorf("failed to find %s/%s image in index for image: %v", goos, goarch, s)
		}
	default:
		return nil, fmt.Errorf("failed to interpret %s result as image: %v", s, br)
	}

	h, err := img.Digest()
	if err != nil {
		return nil, err
	}

	digestTag, err := name.NewTag(fmt.Sprintf("%s:%s", k.namer(K3sDomain, s), h.Hex))
	if err != nil {
		return nil, err
	}

	log.Printf("Loading %v", digestTag)
	if err := k3s.Write(ctx, digestTag, img); err != nil {
		return nil, err
	}
	log.Printf("Loaded %v", digestTag)

	for _, tagName := range k.tags {
		log.Printf("Adding tag %s", tagName)
		tag, err := name.NewTag(fmt.Sprintf("%s:%s", k.namer(K3sDomain, s), tagName))
		if err != nil {
			return nil, err
		}

		if err := k3s.Tag(ctx, digestTag, tag); err != nil {
			return nil, err
		}
		log.Printf("Added tag %v", tagName)
	}

	return &digestTag, nil
}

//Close implements publish.Interface
func (k *k3sPublisher) Close() error {
	return nil
}
