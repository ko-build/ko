// Copyright 2018 Google LLC All Rights Reserved.
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

package build

import (
	"archive/tar"
	"context"
	"fmt"
	gb "go/build"
	"io"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/random"
)

func TestGoBuildIsSupportedRef(t *testing.T) {
	base, err := random.Image(1024, 3)
	if err != nil {
		t.Fatalf("random.Image() = %v", err)
	}

	ng, err := NewGo(WithBaseImages(func(string) (Result, error) { return base, nil }))
	if err != nil {
		t.Fatalf("NewGo() = %v", err)
	}

	// Supported import paths.
	for _, importpath := range []string{
		"ko://github.com/google/ko/cmd/ko", // ko can build itself.
	} {
		t.Run(importpath, func(t *testing.T) {
			if !ng.IsSupportedReference(importpath) {
				t.Errorf("IsSupportedReference(%q) = false, want true", importpath)
			}
		})
	}

	// Unsupported import paths.
	for _, importpath := range []string{
		"ko://github.com/google/ko/pkg/build",       // not a command.
		"ko://github.com/google/ko/pkg/nonexistent", // does not exist.
	} {
		t.Run(importpath, func(t *testing.T) {
			if ng.IsSupportedReference(importpath) {
				t.Errorf("IsSupportedReference(%v) = true, want false", importpath)
			}
		})
	}
}

func TestGoBuildIsSupportedRefWithModules(t *testing.T) {
	base, err := random.Image(1024, 3)
	if err != nil {
		t.Fatalf("random.Image() = %v", err)
	}

	mods := &modules{
		main: &modInfo{
			Path: "github.com/google/ko/cmd/ko/test",
			Dir:  ".",
		},
		deps: map[string]*modInfo{
			"github.com/some/module/cmd": &modInfo{
				Path: "github.com/some/module/cmd",
				Dir:  ".",
			},
		},
	}

	opts := []Option{
		WithBaseImages(func(string) (Result, error) { return base, nil }),
		withModuleInfo(mods),
		withBuildContext(stubBuildContext{
			// make all referenced deps commands
			"github.com/google/ko/cmd/ko/test": &gb.Package{Name: "main"},
			"github.com/some/module/cmd":       &gb.Package{Name: "main"},

			"github.com/google/ko/pkg/build": &gb.Package{Name: "build"},
		}),
	}

	ng, err := NewGo(opts...)
	if err != nil {
		t.Fatalf("NewGo() = %v", err)
	}

	// Supported import paths.
	for _, importpath := range []string{
		"ko://github.com/google/ko/cmd/ko/test", // ko can build the test package.
		"ko://github.com/some/module/cmd",       // ko can build commands in dependent modules
	} {
		t.Run(importpath, func(t *testing.T) {
			if !ng.IsSupportedReference(importpath) {
				t.Errorf("IsSupportedReference(%q) = false, want true", importpath)
			}
		})
	}

	// Unsupported import paths.
	for _, importpath := range []string{
		"ko://github.com/google/ko/pkg/build",       // not a command.
		"ko://github.com/google/ko/pkg/nonexistent", // does not exist.
		"ko://github.com/google/ko/cmd/ko",          // not in this module.
	} {
		t.Run(importpath, func(t *testing.T) {
			if ng.IsSupportedReference(importpath) {
				t.Errorf("IsSupportedReference(%v) = true, want false", importpath)
			}
		})
	}
}

// A helper method we use to substitute for the default "build" method.
func writeTempFile(_ context.Context, s string, _ v1.Platform, _ bool) (string, error) {
	tmpDir, err := ioutil.TempDir("", "ko")
	if err != nil {
		return "", err
	}

	file, err := ioutil.TempFile(tmpDir, "out")
	if err != nil {
		return "", err
	}
	defer file.Close()
	if _, err := file.WriteString(filepath.ToSlash(s)); err != nil {
		return "", err
	}
	return file.Name(), nil
}

func TestGoBuildNoKoData(t *testing.T) {
	baseLayers := int64(3)
	base, err := random.Image(1024, baseLayers)
	if err != nil {
		t.Fatalf("random.Image() = %v", err)
	}
	importpath := "github.com/google/ko"

	creationTime := v1.Time{Time: time.Unix(5000, 0)}
	ng, err := NewGo(
		WithCreationTime(creationTime),
		WithBaseImages(func(string) (Result, error) { return base, nil }),
		withBuilder(writeTempFile),
	)
	if err != nil {
		t.Fatalf("NewGo() = %v", err)
	}

	result, err := ng.Build(context.Background(), StrictScheme+filepath.Join(importpath, "cmd", "ko"))
	if err != nil {
		t.Fatalf("Build() = %v", err)
	}

	img, ok := result.(v1.Image)
	if !ok {
		t.Fatalf("Build() not an image: %v", result)
	}

	ls, err := img.Layers()
	if err != nil {
		t.Fatalf("Layers() = %v", err)
	}

	// Check that we have the expected number of layers.
	t.Run("check layer count", func(t *testing.T) {
		// We get a layer for the go binary and a layer for the kodata/
		if got, want := int64(len(ls)), baseLayers+2; got != want {
			t.Fatalf("len(Layers()) = %v, want %v", got, want)
		}
	})

	// Check that rebuilding the image again results in the same image digest.
	t.Run("check determinism", func(t *testing.T) {
		result2, err := ng.Build(context.Background(), StrictScheme+filepath.Join(importpath, "cmd", "ko"))
		if err != nil {
			t.Fatalf("Build() = %v", err)
		}

		d1, err := result.Digest()
		if err != nil {
			t.Fatalf("Digest() = %v", err)
		}
		d2, err := result2.Digest()
		if err != nil {
			t.Fatalf("Digest() = %v", err)
		}

		if d1 != d2 {
			t.Errorf("Digest mismatch: %s != %s", d1, d2)
		}
	})

	// Check that the entrypoint of the image is configured to invoke our Go application
	t.Run("check entrypoint", func(t *testing.T) {
		cfg, err := img.ConfigFile()
		if err != nil {
			t.Errorf("ConfigFile() = %v", err)
		}
		entrypoint := cfg.Config.Entrypoint
		if got, want := len(entrypoint), 1; got != want {
			t.Errorf("len(entrypoint) = %v, want %v", got, want)
		}

		if got, want := entrypoint[0], "/ko-app/ko"; got != want {
			t.Errorf("entrypoint = %v, want %v", got, want)
		}
	})

	t.Run("check creation time", func(t *testing.T) {
		cfg, err := img.ConfigFile()
		if err != nil {
			t.Errorf("ConfigFile() = %v", err)
		}

		actual := cfg.Created
		if actual.Time != creationTime.Time {
			t.Errorf("created = %v, want %v", actual, creationTime)
		}
	})
}

func validateImage(t *testing.T, img v1.Image, baseLayers int64, creationTime v1.Time) {
	t.Helper()

	ls, err := img.Layers()
	if err != nil {
		t.Fatalf("Layers() = %v", err)
	}

	// Check that we have the expected number of layers.
	t.Run("check layer count", func(t *testing.T) {
		// We get a layer for the go binary and a layer for the kodata/
		if got, want := int64(len(ls)), baseLayers+2; got != want {
			t.Fatalf("len(Layers()) = %v, want %v", got, want)
		}
	})

	t.Run("check app layer contents", func(t *testing.T) {
		dataLayer := ls[baseLayers]

		if _, err := dataLayer.Digest(); err != nil {
			t.Errorf("Digest() = %v", err)
		}
		// We don't check the data layer here because it includes a symlink of refs and
		// will produce a distinct hash each time we commit something.

		r, err := dataLayer.Uncompressed()
		if err != nil {
			t.Errorf("Uncompressed() = %v", err)
		}
		defer r.Close()
		tr := tar.NewReader(r)
		if _, err := tr.Next(); err == io.EOF {
			t.Errorf("Layer contained no files")
		}
	})

	// Check that the kodata layer contains the expected data (even though it was a symlink
	// outside kodata).
	t.Run("check kodata", func(t *testing.T) {
		dataLayer := ls[baseLayers]
		r, err := dataLayer.Uncompressed()
		if err != nil {
			t.Errorf("Uncompressed() = %v", err)
		}
		defer r.Close()
		found := false
		tr := tar.NewReader(r)
		for {
			header, err := tr.Next()
			if err == io.EOF {
				break
			} else if err != nil {
				t.Errorf("Next() = %v", err)
				continue
			}
			if header.Name != path.Join(kodataRoot, "kenobi") {
				continue
			}
			found = true
			body, err := ioutil.ReadAll(tr)
			if err != nil {
				t.Errorf("ReadAll() = %v", err)
			} else if want, got := "Hello there\n", string(body); got != want {
				t.Errorf("ReadAll() = %v, wanted %v", got, want)
			}
		}
		if !found {
			t.Error("Didn't find expected file in tarball")
		}
	})

	// Check that the entrypoint of the image is configured to invoke our Go application
	t.Run("check entrypoint", func(t *testing.T) {
		cfg, err := img.ConfigFile()
		if err != nil {
			t.Errorf("ConfigFile() = %v", err)
		}
		entrypoint := cfg.Config.Entrypoint
		if got, want := len(entrypoint), 1; got != want {
			t.Errorf("len(entrypoint) = %v, want %v", got, want)
		}

		if got, want := entrypoint[0], "/ko-app/test"; got != want {
			t.Errorf("entrypoint = %v, want %v", got, want)
		}
	})

	// Check that the environment contains the KO_DATA_PATH environment variable.
	t.Run("check KO_DATA_PATH env var", func(t *testing.T) {
		cfg, err := img.ConfigFile()
		if err != nil {
			t.Errorf("ConfigFile() = %v", err)
		}
		found := false
		for _, entry := range cfg.Config.Env {
			if entry == "KO_DATA_PATH="+kodataRoot {
				found = true
			}
		}
		if !found {
			t.Error("Didn't find KO_DATA_PATH.")
		}
	})

	// Check that PATH contains the directory of the produced binary.
	t.Run("check PATH env var", func(t *testing.T) {
		cfg, err := img.ConfigFile()
		if err != nil {
			t.Errorf("ConfigFile() = %v", err)
		}
		found := false
		for _, envVar := range cfg.Config.Env {
			if strings.HasPrefix(envVar, "PATH=") {
				pathValue := strings.TrimPrefix(envVar, "PATH=")
				pathEntries := strings.Split(pathValue, ":")
				for _, pathEntry := range pathEntries {
					if pathEntry == appDir {
						found = true
					}
				}
			}
		}
		if !found {
			t.Error("Didn't find entrypoint in PATH.")
		}
	})

	t.Run("check creation time", func(t *testing.T) {
		cfg, err := img.ConfigFile()
		if err != nil {
			t.Errorf("ConfigFile() = %v", err)
		}

		actual := cfg.Created
		if actual.Time != creationTime.Time {
			t.Errorf("created = %v, want %v", actual, creationTime)
		}
	})
}

type stubBuildContext map[string]*gb.Package

func (s stubBuildContext) Import(path string, srcDir string, mode gb.ImportMode) (*gb.Package, error) {
	p, ok := s[path]
	if ok {
		return p, nil
	}
	return nil, fmt.Errorf("not found: %s", path)
}

func TestGoBuild(t *testing.T) {
	baseLayers := int64(3)
	base, err := random.Image(1024, baseLayers)
	if err != nil {
		t.Fatalf("random.Image() = %v", err)
	}
	importpath := "github.com/google/ko"

	creationTime := v1.Time{time.Unix(5000, 0)}
	ng, err := NewGo(
		WithCreationTime(creationTime),
		WithBaseImages(func(string) (Result, error) { return base, nil }),
		withBuilder(writeTempFile),
	)
	if err != nil {
		t.Fatalf("NewGo() = %v", err)
	}

	result, err := ng.Build(context.Background(), StrictScheme+filepath.Join(importpath, "cmd", "ko", "test"))
	if err != nil {
		t.Fatalf("Build() = %v", err)
	}

	img, ok := result.(v1.Image)
	if !ok {
		t.Fatalf("Build() not an image: %v", result)
	}

	validateImage(t, img, baseLayers, creationTime)

	// Check that rebuilding the image again results in the same image digest.
	t.Run("check determinism", func(t *testing.T) {
		result2, err := ng.Build(context.Background(), StrictScheme+filepath.Join(importpath, "cmd", "ko", "test"))
		if err != nil {
			t.Fatalf("Build() = %v", err)
		}

		d1, err := result.Digest()
		if err != nil {
			t.Fatalf("Digest() = %v", err)
		}
		d2, err := result2.Digest()
		if err != nil {
			t.Fatalf("Digest() = %v", err)
		}

		if d1 != d2 {
			t.Errorf("Digest mismatch: %s != %s", d1, d2)
		}
	})

}

func TestGoBuildIndex(t *testing.T) {
	baseLayers := int64(3)
	images := int64(2)
	base, err := random.Index(1024, baseLayers, images)
	if err != nil {
		t.Fatalf("random.Image() = %v", err)
	}
	importpath := "github.com/google/ko"

	creationTime := v1.Time{time.Unix(5000, 0)}
	ng, err := NewGo(
		WithCreationTime(creationTime),
		WithBaseImages(func(string) (Result, error) { return base, nil }),
		withBuilder(writeTempFile),
	)
	if err != nil {
		t.Fatalf("NewGo() = %v", err)
	}

	result, err := ng.Build(context.Background(), StrictScheme+filepath.Join(importpath, "cmd", "ko", "test"))
	if err != nil {
		t.Fatalf("Build() = %v", err)
	}

	idx, ok := result.(v1.ImageIndex)
	if !ok {
		t.Fatalf("Build() not an image: %v", result)
	}

	im, err := idx.IndexManifest()
	if err != nil {
		t.Fatalf("IndexManifest() = %v", err)
	}

	for _, desc := range im.Manifests {
		img, err := idx.Image(desc.Digest)
		if err != nil {
			t.Fatalf("idx.Image(%s) = %v", desc.Digest, err)
		}
		validateImage(t, img, baseLayers, creationTime)
	}

	if want, got := images, int64(len(im.Manifests)); want != got {
		t.Fatalf("len(Manifests()) = %v, want %v", got, want)
	}

	// Check that rebuilding the image again results in the same image digest.
	t.Run("check determinism", func(t *testing.T) {
		result2, err := ng.Build(context.Background(), StrictScheme+filepath.Join(importpath, "cmd", "ko", "test"))
		if err != nil {
			t.Fatalf("Build() = %v", err)
		}

		d1, err := result.Digest()
		if err != nil {
			t.Fatalf("Digest() = %v", err)
		}
		d2, err := result2.Digest()
		if err != nil {
			t.Fatalf("Digest() = %v", err)
		}

		if d1 != d2 {
			t.Errorf("Digest mismatch: %s != %s", d1, d2)
		}
	})
}

func TestNestedIndex(t *testing.T) {
	baseLayers := int64(3)
	images := int64(2)
	base, err := random.Index(1024, baseLayers, images)
	if err != nil {
		t.Fatalf("random.Image() = %v", err)
	}
	importpath := "github.com/google/ko"

	nestedBase := mutate.AppendManifests(empty.Index, mutate.IndexAddendum{Add: base})

	creationTime := v1.Time{time.Unix(5000, 0)}
	ng, err := NewGo(
		WithCreationTime(creationTime),
		WithBaseImages(func(string) (Result, error) { return nestedBase, nil }),
		withBuilder(writeTempFile),
	)
	if err != nil {
		t.Fatalf("NewGo() = %v", err)
	}

	_, err = ng.Build(context.Background(), StrictScheme+filepath.Join(importpath, "cmd", "ko", "test"))
	if err == nil {
		t.Fatal("Build() expected err")
	}

	if !strings.Contains(err.Error(), "unexpected mediaType") {
		t.Errorf("Build() expected unexpected mediaType error, got: %s", err)
	}
}
