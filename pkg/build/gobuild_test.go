// Copyright 2018 ko Build Authors All Rights Reserved.
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
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/types"
	specsv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sigstore/cosign/v2/pkg/oci"
)

func repoRootDir() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("could not get current filename")
	}
	basepath := filepath.Dir(filename)
	repoDir := filepath.Join(basepath, "..", "..")
	return filepath.Rel(basepath, repoDir)
}

func TestGoBuildQualifyImport(t *testing.T) {
	base, err := random.Image(1024, 1)
	if err != nil {
		t.Fatalf("random.Image() = %v", err)
	}

	repoDir, err := repoRootDir()
	if err != nil {
		t.Fatalf("could not get Git repository root directory")
	}

	tests := []struct {
		description         string
		rawImportpath       string
		dir                 string
		qualifiedImportpath string
		expectError         bool
	}{
		{
			description:         "strict qualified import path",
			rawImportpath:       "ko://github.com/google/ko",
			dir:                 "",
			qualifiedImportpath: "ko://github.com/google/ko",
			expectError:         false,
		},
		{
			description:         "strict qualified import path in subdirectory of go.mod",
			rawImportpath:       "ko://github.com/google/ko/test",
			dir:                 "",
			qualifiedImportpath: "ko://github.com/google/ko/test",
			expectError:         false,
		},
		{
			description:         "non-strict qualified import path",
			rawImportpath:       "github.com/google/ko",
			dir:                 "",
			qualifiedImportpath: "ko://github.com/google/ko",
			expectError:         false,
		},
		{
			description:         "non-strict local import path in repository root directory",
			rawImportpath:       "./test",
			dir:                 repoDir,
			qualifiedImportpath: "ko://github.com/google/ko/test",
			expectError:         false,
		},
		{
			description:         "non-strict local import path in subdirectory",
			rawImportpath:       ".",
			dir:                 filepath.Join(repoDir, "test"),
			qualifiedImportpath: "ko://github.com/google/ko/test",
			expectError:         false,
		},
		{
			description:         "non-existent non-strict local import path",
			rawImportpath:       "./does-not-exist",
			dir:                 "/",
			qualifiedImportpath: "should return error",
			expectError:         true,
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			ng, err := NewGo(context.Background(), test.dir, WithBaseImages(func(context.Context, string) (name.Reference, Result, error) { return nil, base, nil }))
			if err != nil {
				t.Fatalf("NewGo() = %v", err)
			}
			gotImportpath, err := ng.QualifyImport(test.rawImportpath)
			if err != nil && test.expectError {
				return
			}
			if err != nil && !test.expectError {
				t.Errorf("QualifyImport(dir=%q)(%q) was error (%v), want nil error", test.dir, test.rawImportpath, err)
			}
			if err == nil && test.expectError {
				t.Errorf("QualifyImport(dir=%q)(%q) was nil error, want non-nil error", test.dir, test.rawImportpath)
			}
			if gotImportpath != test.qualifiedImportpath {
				t.Errorf("QualifyImport(dir=%q)(%q) = (%q, nil), want (%q, nil)", test.dir, test.rawImportpath, gotImportpath, test.qualifiedImportpath)
			}
		})
	}
}

var baseRef = name.MustParseReference("all.your/base")

func TestGoBuildIsSupportedRef(t *testing.T) {
	base, err := random.Image(1024, 3)
	if err != nil {
		t.Fatalf("random.Image() = %v", err)
	}

	ng, err := NewGo(context.Background(), "", WithBaseImages(func(context.Context, string) (name.Reference, Result, error) { return nil, base, nil }))
	if err != nil {
		t.Fatalf("NewGo() = %v", err)
	}

	// Supported import paths.
	for _, importpath := range []string{
		"ko://github.com/google/ko", // ko can build itself.
	} {
		t.Run(importpath, func(t *testing.T) {
			if err := ng.IsSupportedReference(importpath); err != nil {
				t.Errorf("IsSupportedReference(%q) = (%v), want nil", importpath, err)
			}
		})
	}

	// Unsupported import paths.
	for _, importpath := range []string{
		"ko://github.com/google/ko/pkg/build",       // not a command.
		"ko://github.com/google/ko/pkg/nonexistent", // does not exist.
	} {
		t.Run(importpath, func(t *testing.T) {
			if err := ng.IsSupportedReference(importpath); err == nil {
				t.Errorf("IsSupportedReference(%v) = nil, want error", importpath)
			}
		})
	}
}

func TestGoBuildIsSupportedRefWithModules(t *testing.T) {
	base, err := random.Image(1024, 3)
	if err != nil {
		t.Fatalf("random.Image() = %v", err)
	}

	opts := []Option{
		WithBaseImages(func(context.Context, string) (name.Reference, Result, error) { return baseRef, base, nil }),
	}

	ng, err := NewGo(context.Background(), "", opts...)
	if err != nil {
		t.Fatalf("NewGo() = %v", err)
	}

	// Supported import paths.
	for _, importpath := range []string{
		"ko://github.com/google/ko/test",         // ko can build the test package.
		"ko://github.com/go-training/helloworld", // ko can build commands in dependent modules
	} {
		t.Run(importpath, func(t *testing.T) {
			if err := ng.IsSupportedReference(importpath); err != nil {
				t.Errorf("IsSupportedReference(%q) = (%v), want nil", err, importpath)
			}
		})
	}

	// Unsupported import paths.
	for _, importpath := range []string{
		"ko://github.com/google/ko/pkg/build",       // not a command.
		"ko://github.com/google/ko/pkg/nonexistent", // does not exist.
		"ko://github.com/google/go-github",          // not in this module.
	} {
		t.Run(importpath, func(t *testing.T) {
			if err := ng.IsSupportedReference(importpath); err == nil {
				t.Errorf("IsSupportedReference(%v) = nil, want error", importpath)
			}
		})
	}
}

func TestBuildEnv(t *testing.T) {
	tests := []struct {
		description  string
		platform     v1.Platform
		userEnv      []string
		configEnv    []string
		expectedEnvs map[string]string
	}{{
		description: "defaults",
		platform: v1.Platform{
			OS:           "linux",
			Architecture: "amd64",
		},
		expectedEnvs: map[string]string{
			"GOOS":        "linux",
			"GOARCH":      "amd64",
			"CGO_ENABLED": "0",
		},
	}, {
		description: "override a default value",
		configEnv:   []string{"CGO_ENABLED=1"},
		expectedEnvs: map[string]string{
			"CGO_ENABLED": "1",
		},
	}, {
		description: "override an envvar and add an envvar",
		userEnv:     []string{"CGO_ENABLED=0"},
		configEnv:   []string{"CGO_ENABLED=1", "GOPRIVATE=git.internal.example.com,source.developers.google.com"},
		expectedEnvs: map[string]string{
			"CGO_ENABLED": "1",
			"GOPRIVATE":   "git.internal.example.com,source.developers.google.com",
		},
	}, {
		description: "arm variant",
		platform: v1.Platform{
			Architecture: "arm",
			Variant:      "v7",
		},
		expectedEnvs: map[string]string{
			"GOARCH": "arm",
			"GOARM":  "7",
		},
	}, {
		// GOARM is ignored for arm64.
		description: "arm64 variant",
		platform: v1.Platform{
			Architecture: "arm64",
			Variant:      "v8",
		},
		expectedEnvs: map[string]string{
			"GOARCH": "arm64",
			"GOARM":  "",
		},
	}, {
		description: "amd64 variant",
		platform: v1.Platform{
			Architecture: "amd64",
			Variant:      "v3",
		},
		expectedEnvs: map[string]string{
			"GOARCH":  "amd64",
			"GOAMD64": "v3",
		},
	}}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			env, err := buildEnv(test.platform, test.userEnv, test.configEnv)
			if err != nil {
				t.Fatalf("unexpected error running buildEnv(): %v", err)
			}
			envs := map[string]string{}
			for _, e := range env {
				split := strings.SplitN(e, "=", 2)
				envs[split[0]] = split[1]
			}
			for key, val := range test.expectedEnvs {
				if envs[key] != val {
					t.Errorf("buildEnv(): expected %s=%s, got %s=%s", key, val, key, envs[key])
				}
			}
		})
	}
}

func TestBuildConfig(t *testing.T) {
	tests := []struct {
		description  string
		options      []Option
		importpath   string
		expectConfig Config
	}{
		{
			description: "minimal options",
			options: []Option{
				WithBaseImages(nilGetBase),
			},
		},
		{
			description: "trimpath flag",
			options: []Option{
				WithBaseImages(nilGetBase),
				WithTrimpath(true),
			},
			expectConfig: Config{
				Flags: FlagArray{"-trimpath"},
			},
		},
		{
			description: "no trimpath flag",
			options: []Option{
				WithBaseImages(nilGetBase),
				WithTrimpath(false),
			},
		},
		{
			description: "build config and trimpath",
			options: []Option{
				WithBaseImages(nilGetBase),
				WithConfig(map[string]Config{
					"example.com/foo": {
						Flags: FlagArray{"-v"},
					},
				}),
				WithTrimpath(true),
			},
			importpath: "example.com/foo",
			expectConfig: Config{
				Flags: FlagArray{"-v", "-trimpath"},
			},
		},
		{
			description: "no trimpath overridden by build config flag",
			options: []Option{
				WithBaseImages(nilGetBase),
				WithConfig(map[string]Config{
					"example.com/bar": {
						Flags: FlagArray{"-trimpath"},
					},
				}),
				WithTrimpath(false),
			},
			importpath: "example.com/bar",
			expectConfig: Config{
				Flags: FlagArray{"-trimpath"},
			},
		},
		{
			description: "disable optimizations",
			options: []Option{
				WithBaseImages(nilGetBase),
				WithDisabledOptimizations(),
			},
			expectConfig: Config{
				Flags: FlagArray{"-gcflags", "all=-N -l"},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			i, err := NewGo(context.Background(), "", test.options...)
			if err != nil {
				t.Fatalf("NewGo(): unexpected error: %+v", err)
			}
			gb, ok := i.(*gobuild)
			if !ok {
				t.Fatal("NewGo() did not return *gobuild{} as expected")
			}
			config := gb.configForImportPath(test.importpath)
			if diff := cmp.Diff(test.expectConfig, config, cmpopts.EquateEmpty(),
				cmpopts.SortSlices(func(x, y string) bool { return x < y })); diff != "" {
				t.Errorf("%T differ (-got, +want): %s", test.expectConfig, diff)
			}
		})
	}
}

func nilGetBase(context.Context, string) (name.Reference, Result, error) {
	return nil, nil, nil
}

const wantSBOM = "This is our fake SBOM"

// A helper method we use to substitute for the default "build" method.
func fauxSBOM(context.Context, string, string, string, oci.SignedEntity, string) ([]byte, types.MediaType, error) {
	return []byte(wantSBOM), "application/vnd.garbage", nil
}

// A helper method we use to substitute for the default "build" method.
func writeTempFile(_ context.Context, s string, _ string, _ v1.Platform, _ Config) (string, error) {
	tmpDir, err := os.MkdirTemp("", "ko")
	if err != nil {
		return "", err
	}

	file, err := os.CreateTemp(tmpDir, "out")
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
		context.Background(),
		"",
		WithCreationTime(creationTime),
		WithBaseImages(func(context.Context, string) (name.Reference, Result, error) { return baseRef, base, nil }),
		withBuilder(writeTempFile),
		withSBOMber(fauxSBOM),
		WithPlatforms("all"),
	)
	if err != nil {
		t.Fatalf("NewGo() = %v", err)
	}

	result, err := ng.Build(context.Background(), StrictScheme+importpath)
	if err != nil {
		t.Fatalf("Build() = %v", err)
	}

	img, ok := result.(v1.Image)
	if !ok {
		t.Fatalf("Build() not an Image: %T", result)
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
		result2, err := ng.Build(context.Background(), StrictScheme+importpath)
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

func validateImage(t *testing.T, img oci.SignedImage, baseLayers int64, creationTime v1.Time, checkAnnotations bool, expectSBOM bool) {
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
		if _, err := tr.Next(); errors.Is(err, io.EOF) {
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
			if errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				t.Errorf("Next() = %v", err)
				continue
			}
			if header.Name != path.Join(kodataRoot, "kenobi") {
				continue
			}
			found = true
			body, err := io.ReadAll(tr)
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
					if pathEntry == "/ko-app" {
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

	t.Run("check annotations", func(t *testing.T) {
		if !checkAnnotations {
			t.Skip("skipping annotations check")
		}
		mf, err := img.Manifest()
		if err != nil {
			t.Fatalf("Manifest() = %v", err)
		}
		t.Logf("Got annotations: %v", mf.Annotations)
		if _, found := mf.Annotations[specsv1.AnnotationBaseImageDigest]; !found {
			t.Errorf("image annotations did not contain base image digest")
		}
		want := baseRef.Name()
		if got := mf.Annotations[specsv1.AnnotationBaseImageName]; got != want {
			t.Errorf("base image ref; got %q, want %q", got, want)
		}
	})

	if expectSBOM {
		t.Run("checking for SBOM", func(t *testing.T) {
			f, err := img.Attachment("sbom")
			if err != nil {
				t.Fatalf("Attachment() = %v", err)
			}
			b, err := f.Payload()
			if err != nil {
				t.Fatalf("Payload() = %v", err)
			}
			t.Logf("Got SBOM: %v", string(b))
			if string(b) != wantSBOM {
				t.Errorf("got SBOM %s, wanted %s", string(b), wantSBOM)
			}
		})
	} else {
		t.Run("checking for no SBOM", func(t *testing.T) {
			f, err := img.Attachment("sbom")
			if err == nil {
				b, err := f.Payload()
				if err != nil {
					t.Fatalf("Payload() = %v", err)
				}
				t.Fatalf("Attachment() = %v, wanted error", string(b))
			}
		})
	}
}

func TestGoBuild(t *testing.T) {
	baseLayers := int64(3)
	base, err := random.Image(1024, baseLayers)
	if err != nil {
		t.Fatalf("random.Image() = %v", err)
	}
	importpath := "github.com/google/ko"

	creationTime := v1.Time{Time: time.Unix(5000, 0)}
	ng, err := NewGo(
		context.Background(),
		"",
		WithCreationTime(creationTime),
		WithBaseImages(func(context.Context, string) (name.Reference, Result, error) { return baseRef, base, nil }),
		withBuilder(writeTempFile),
		withSBOMber(fauxSBOM),
		WithLabel("foo", "bar"),
		WithLabel("hello", "world"),
		WithPlatforms("all"),
	)
	if err != nil {
		t.Fatalf("NewGo() = %v", err)
	}

	result, err := ng.Build(context.Background(), StrictScheme+filepath.Join(importpath, "test"))
	if err != nil {
		t.Fatalf("Build() = %v", err)
	}

	img, ok := result.(oci.SignedImage)
	if !ok {
		t.Fatalf("Build() not a SignedImage: %T", result)
	}

	validateImage(t, img, baseLayers, creationTime, true, true)

	// Check that rebuilding the image again results in the same image digest.
	t.Run("check determinism", func(t *testing.T) {
		result2, err := ng.Build(context.Background(), StrictScheme+filepath.Join(importpath, "test"))
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

	t.Run("check labels", func(t *testing.T) {
		cfg, err := img.ConfigFile()
		if err != nil {
			t.Fatalf("ConfigFile() = %v", err)
		}

		want := map[string]string{
			"foo":   "bar",
			"hello": "world",
		}
		got := cfg.Config.Labels
		if d := cmp.Diff(got, want); d != "" {
			t.Fatalf("Labels diff (-got,+want): %s", d)
		}
	})
}

func TestGoBuildWithKOCACHE(t *testing.T) {
	now := time.Now() // current local time
	sec := now.Unix()
	tmpDir := t.TempDir()
	koCacheDir := filepath.Join(tmpDir, strconv.FormatInt(sec, 10))

	t.Setenv("KOCACHE", koCacheDir)
	baseLayers := int64(3)
	base, err := random.Image(1024, baseLayers)
	if err != nil {
		t.Fatalf("random.Image() = %v", err)
	}
	importpath := "github.com/google/ko"

	creationTime := v1.Time{Time: time.Unix(5000, 0)}
	ng, err := NewGo(
		context.Background(),
		"",
		WithCreationTime(creationTime),
		WithBaseImages(func(context.Context, string) (name.Reference, Result, error) { return baseRef, base, nil }),
		WithPlatforms("all"),
	)
	if err != nil {
		t.Fatalf("NewGo() = %v", err)
	}

	_, err = ng.Build(context.Background(), StrictScheme+filepath.Join(importpath, "test"))
	if err != nil {
		t.Fatalf("Build() = %v", err)
	}

	t.Run("check KOCACHE exists", func(t *testing.T) {
		_, err := os.Stat(koCacheDir)
		if os.IsNotExist(err) {
			t.Fatalf("KOCACHE directory %s should be exists= %v", koCacheDir, err)
		}
	})
}

func TestGoBuildWithoutSBOM(t *testing.T) {
	baseLayers := int64(3)
	base, err := random.Image(1024, baseLayers)
	if err != nil {
		t.Fatalf("random.Image() = %v", err)
	}
	importpath := "github.com/google/ko"

	creationTime := v1.Time{Time: time.Unix(5000, 0)}
	ng, err := NewGo(
		context.Background(),
		"",
		WithCreationTime(creationTime),
		WithBaseImages(func(context.Context, string) (name.Reference, Result, error) { return baseRef, base, nil }),
		withBuilder(writeTempFile),
		withSBOMber(fauxSBOM),
		WithLabel("foo", "bar"),
		WithLabel("hello", "world"),
		WithDisabledSBOM(),
		WithPlatforms("all"),
	)
	if err != nil {
		t.Fatalf("NewGo() = %v", err)
	}

	result, err := ng.Build(context.Background(), StrictScheme+filepath.Join(importpath, "test"))
	if err != nil {
		t.Fatalf("Build() = %v", err)
	}

	img, ok := result.(oci.SignedImage)
	if !ok {
		t.Fatalf("Build() not a SignedImage: %T", result)
	}

	validateImage(t, img, baseLayers, creationTime, true, false)
}

func TestGoBuildIndex(t *testing.T) {
	baseLayers := int64(3)
	images := int64(2)
	base, err := random.Index(1024, baseLayers, images)
	if err != nil {
		t.Fatalf("random.Image() = %v", err)
	}
	importpath := "github.com/google/ko"

	creationTime := v1.Time{Time: time.Unix(5000, 0)}
	ng, err := NewGo(
		context.Background(),
		"",
		WithCreationTime(creationTime),
		WithBaseImages(func(context.Context, string) (name.Reference, Result, error) { return baseRef, base, nil }),
		WithPlatforms("all"),
		withBuilder(writeTempFile),
		withSBOMber(fauxSBOM),
	)
	if err != nil {
		t.Fatalf("NewGo() = %v", err)
	}

	result, err := ng.Build(context.Background(), StrictScheme+filepath.Join(importpath, "test"))
	if err != nil {
		t.Fatalf("Build() = %v", err)
	}

	idx, ok := result.(oci.SignedImageIndex)
	if !ok {
		t.Fatalf("Build() not a SignedImageIndex: %T", result)
	}

	im, err := idx.IndexManifest()
	if err != nil {
		t.Fatalf("IndexManifest() = %v", err)
	}

	for _, desc := range im.Manifests {
		img, err := idx.SignedImage(desc.Digest)
		if err != nil {
			t.Fatalf("idx.Image(%s) = %v", desc.Digest, err)
		}
		validateImage(t, img, baseLayers, creationTime, false, true)
	}

	if want, got := images, int64(len(im.Manifests)); want != got {
		t.Fatalf("len(Manifests()) = %v, want %v", got, want)
	}

	// Check that rebuilding the image again results in the same image digest.
	t.Run("check determinism", func(t *testing.T) {
		result2, err := ng.Build(context.Background(), StrictScheme+filepath.Join(importpath, "test"))
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

	creationTime := v1.Time{Time: time.Unix(5000, 0)}
	ng, err := NewGo(
		context.Background(),
		"",
		WithCreationTime(creationTime),
		WithBaseImages(func(context.Context, string) (name.Reference, Result, error) { return baseRef, nestedBase, nil }),
		withBuilder(writeTempFile),
		withSBOMber(fauxSBOM),
	)
	if err != nil {
		t.Fatalf("NewGo() = %v", err)
	}

	_, err = ng.Build(context.Background(), StrictScheme+filepath.Join(importpath, "test"))
	if err == nil {
		t.Fatal("Build() expected err")
	}

	if !strings.Contains(err.Error(), "unexpected mediaType") {
		t.Errorf("Build() expected unexpected mediaType error, got: %s", err)
	}
}

func TestGoarm(t *testing.T) {
	// From golang@sha256:1ba0da74b20aad52b091877b0e0ece503c563f39e37aa6b0e46777c4d820a2ae
	// and made up invalid cases.
	for _, tc := range []struct {
		platform v1.Platform
		variant  string
		err      bool
	}{{
		platform: v1.Platform{
			Architecture: "arm",
			OS:           "linux",
			Variant:      "vnot-a-number",
		},
		err: true,
	}, {
		platform: v1.Platform{
			Architecture: "arm",
			OS:           "linux",
			Variant:      "wrong-prefix",
		},
		err: true,
	}, {
		platform: v1.Platform{
			Architecture: "arm64",
			OS:           "linux",
			Variant:      "v3",
		},
		variant: "",
	}, {
		platform: v1.Platform{
			Architecture: "arm",
			OS:           "linux",
			Variant:      "v5",
		},
		variant: "5",
	}, {
		platform: v1.Platform{
			Architecture: "arm",
			OS:           "linux",
			Variant:      "v7",
		},
		variant: "7",
	}, {
		platform: v1.Platform{
			Architecture: "arm64",
			OS:           "linux",
			Variant:      "v8",
		},
		variant: "7",
	},
	} {
		variant, err := getGoarm(tc.platform)
		if tc.err {
			if err == nil {
				t.Errorf("getGoarm(%v) expected err", tc.platform)
			}
			continue
		}
		if err != nil {
			t.Fatalf("getGoarm failed for %v: %v", tc.platform, err)
		}
		if got, want := variant, tc.variant; got != want {
			t.Errorf("wrong variant for %v: want %q got %q", tc.platform, want, got)
		}
	}
}

func TestMatchesPlatformSpec(t *testing.T) {
	for _, tc := range []struct {
		platform *v1.Platform
		spec     []string
		result   bool
		err      bool
	}{{
		platform: nil,
		spec:     []string{"all"},
		result:   true,
	}, {
		platform: nil,
		spec:     []string{"linux/amd64"},
		result:   false,
	}, {
		platform: &v1.Platform{
			Architecture: "amd64",
			OS:           "linux",
		},
		spec:   []string{"all"},
		result: true,
	}, {
		platform: &v1.Platform{
			Architecture: "amd64",
			OS:           "windows",
		},
		spec:   []string{"linux"},
		result: false,
	}, {
		platform: &v1.Platform{
			Architecture: "arm64",
			OS:           "linux",
			Variant:      "v3",
		},
		spec:   []string{"linux/amd64", "linux/arm64"},
		result: true,
	}, {
		platform: &v1.Platform{
			Architecture: "arm64",
			OS:           "linux",
			Variant:      "v3",
		},
		spec:   []string{"linux/amd64", "linux/arm64/v4"},
		result: false,
	}, {
		platform: &v1.Platform{
			Architecture: "arm64",
			OS:           "linux",
			Variant:      "v3",
		},
		spec: []string{"linux/amd64", "linux/arm64/v3/z5"},
		err:  true,
	}, {
		spec: []string{},
		platform: &v1.Platform{
			Architecture: "amd64",
			OS:           "linux",
		},
		result: false,
	}, {
		// Exact match w/ osversion
		spec: []string{"windows/amd64:10.0.17763.1234"},
		platform: &v1.Platform{
			OS:           "windows",
			Architecture: "amd64",
			OSVersion:    "10.0.17763.1234",
		},
		result: true,
	}, {
		// OSVersion partial match using relaxed semantics.
		spec: []string{"windows/amd64:10.0.17763"},
		platform: &v1.Platform{
			OS:           "windows",
			Architecture: "amd64",
			OSVersion:    "10.0.17763.1234",
		},
		result: true,
	}, {
		// Not windows and osversion isn't exact match.
		spec: []string{"linux/amd64:10.0.17763"},
		platform: &v1.Platform{
			OS:           "linux",
			Architecture: "amd64",
			OSVersion:    "10.0.17763.1234",
		},
		result: false,
	}, {
		// Not matching X.Y.Z
		spec: []string{"windows/amd64:10"},
		platform: &v1.Platform{
			OS:           "windows",
			Architecture: "amd64",
			OSVersion:    "10.0.17763.1234",
		},
		result: false,
	}, {
		// Requirement is more specific.
		spec: []string{"windows/amd64:10.0.17763.1234"},
		platform: &v1.Platform{
			OS:           "windows",
			Architecture: "amd64",
			OSVersion:    "10.0.17763", // this won't happen in the wild, but it shouldn't match.
		},
		result: false,
	}, {
		// Requirement is not specific enough.
		spec: []string{"windows/amd64:10.0.17763.1234"},
		platform: &v1.Platform{
			OS:           "windows",
			Architecture: "amd64",
			OSVersion:    "10.0.17763.1234.5678", // this won't happen in the wild, but it shouldn't match.
		},
		result: false,
	}, {
		// Even --platform=all does not match unknown/unknown.
		platform: &v1.Platform{Architecture: "unknown", OS: "unknown"},
		spec:     []string{"all"},
		result:   false,
	}} {
		pm, err := parseSpec(tc.spec)
		if tc.err {
			if err == nil {
				t.Errorf("parseSpec(%v, %q) expected err", tc.platform, tc.spec)
			}
			continue
		}
		if err != nil {
			t.Fatalf("parseSpec failed for %v %q: %v", tc.platform, tc.spec, err)
		}
		matches := pm.matches(tc.platform)
		if got, want := matches, tc.result; got != want {
			t.Errorf("wrong result for %v %q: want %t got %t", tc.platform, tc.spec, want, got)
		}
	}
}

func TestGoBuildConsistentMediaTypes(t *testing.T) {
	for _, c := range []struct {
		desc                      string
		mediaType, layerMediaType types.MediaType
	}{{
		desc:           "docker types",
		mediaType:      types.DockerManifestSchema2,
		layerMediaType: types.DockerLayer,
	}, {
		desc:           "oci types",
		mediaType:      types.OCIManifestSchema1,
		layerMediaType: types.OCILayer,
	}} {
		t.Run(c.desc, func(t *testing.T) {
			base := mutate.MediaType(empty.Image, c.mediaType)
			l, err := random.Layer(10, c.layerMediaType)
			if err != nil {
				t.Fatal(err)
			}
			base, err = mutate.AppendLayers(base, l)
			if err != nil {
				t.Fatal(err)
			}

			ng, err := NewGo(
				context.Background(),
				"",
				WithBaseImages(func(context.Context, string) (name.Reference, Result, error) { return baseRef, base, nil }),
				withBuilder(writeTempFile),
				withSBOMber(fauxSBOM),
				WithPlatforms("all"),
			)
			if err != nil {
				t.Fatalf("NewGo() = %v", err)
			}

			importpath := "github.com/google/ko"

			result, err := ng.Build(context.Background(), StrictScheme+importpath)
			if err != nil {
				t.Fatalf("Build() = %v", err)
			}

			img, ok := result.(v1.Image)
			if !ok {
				t.Fatalf("Build() not an Image: %T", result)
			}

			ls, err := img.Layers()
			if err != nil {
				t.Fatalf("Layers() = %v", err)
			}

			for i, l := range ls {
				gotMT, err := l.MediaType()
				if err != nil {
					t.Fatal(err)
				}
				if gotMT != c.layerMediaType {
					t.Errorf("layer %d: got mediaType %q, want %q", i, gotMT, c.layerMediaType)
				}
			}

			gotMT, err := img.MediaType()
			if err != nil {
				t.Fatal(err)
			}
			if gotMT != c.mediaType {
				t.Errorf("got image mediaType %q, want %q", gotMT, c.layerMediaType)
			}
		})
	}
}
