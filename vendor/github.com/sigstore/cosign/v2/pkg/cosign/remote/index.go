//
// Copyright 2021 The Sigstore Authors.
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

package remote

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/sigstore/cosign/v2/pkg/oci/static"
)

type File interface {
	Contents() ([]byte, error)
	Platform() *v1.Platform
	String() string
	Path() string
}

func FilesFromFlagList(sl []string) []File {
	files := make([]File, len(sl))
	for i, s := range sl {
		files[i] = FileFromFlag(s)
	}
	return files
}

func FileFromFlag(s string) File {
	split := strings.Split(s, ":")
	f := file{
		path: split[0],
	}
	if len(split) > 1 {
		split = strings.Split(split[1], "/")
		f.platform = &v1.Platform{
			OS: split[0],
		}
		if len(split) > 1 {
			f.platform.Architecture = split[1]
		}
	}
	return &f
}

type file struct {
	path     string
	platform *v1.Platform
}

func (f *file) Path() string {
	return f.path
}

func (f *file) Contents() ([]byte, error) {
	return os.ReadFile(f.path)
}
func (f *file) Platform() *v1.Platform {
	return f.platform
}

func (f *file) String() string {
	r := f.path
	if f.platform == nil {
		return r
	}
	r += ":" + f.platform.OS
	if f.platform.Architecture == "" {
		return r
	}
	r += "/" + f.platform.Architecture
	return r
}

type MediaTypeGetter func(b []byte) types.MediaType

func DefaultMediaTypeGetter(b []byte) types.MediaType {
	return types.MediaType(strings.Split(http.DetectContentType(b), ";")[0])
}

func UploadFiles(ref name.Reference, files []File, annotations map[string]string, getMt MediaTypeGetter, remoteOpts ...remote.Option) (name.Digest, error) {
	var lastHash v1.Hash
	var idx v1.ImageIndex = empty.Index

	for _, f := range files {
		b, err := f.Contents()
		if err != nil {
			return name.Digest{}, err
		}
		mt := getMt(b)
		fmt.Fprintf(os.Stderr, "Uploading file from [%s] to [%s] with media type [%s]\n", f.Path(), ref.Name(), mt)

		img, err := static.NewFile(b, static.WithLayerMediaType(mt), static.WithAnnotations(annotations))
		if err != nil {
			return name.Digest{}, err
		}

		lastHash, err = img.Digest()
		if err != nil {
			return name.Digest{}, err
		}

		if err := remote.Write(ref, img, remoteOpts...); err != nil {
			return name.Digest{}, err
		}
		l, err := img.Layers()
		if err != nil {
			return name.Digest{}, err
		}
		layerHash, err := l[0].Digest()
		if err != nil {
			return name.Digest{}, err
		}

		blobURL := ref.Context().Registry.RegistryStr() + "/v2/" + ref.Context().RepositoryStr() + "/blobs/" + layerHash.String()
		fmt.Fprintf(os.Stderr, "File [%s] is available directly at [%s]\n", f.Path(), blobURL)

		if f.Platform() != nil {
			idx = mutate.AppendManifests(idx, mutate.IndexAddendum{
				Add: img,
				Descriptor: v1.Descriptor{
					Platform: f.Platform(),
				},
			})
		}
	}

	if len(files) > 1 {
		if annotations != nil {
			idx = mutate.Annotations(idx, annotations).(v1.ImageIndex)
		}
		err := remote.WriteIndex(ref, idx, remoteOpts...)
		if err != nil {
			return name.Digest{}, err
		}
		lastHash, err = idx.Digest()
		if err != nil {
			return name.Digest{}, err
		}
	}
	return ref.Context().Digest(lastHash.String()), nil
}
