/*
Copyright 2018 Google LLC All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package build

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	gb "go/build"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/containerd/stargz-snapshotter/estargz"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/google/ko/internal/sbom"
	"github.com/google/ko/pkg/build/binary"
	config "github.com/google/ko/pkg/build/config"
	buildoci "github.com/google/ko/pkg/build/oci"
	buildplatform "github.com/google/ko/pkg/build/platform"
	specsv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sigstore/cosign/pkg/oci"
	ocimutate "github.com/sigstore/cosign/pkg/oci/mutate"
	"github.com/sigstore/cosign/pkg/oci/signed"
	"github.com/sigstore/cosign/pkg/oci/static"
	ctypes "github.com/sigstore/cosign/pkg/types"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"golang.org/x/tools/go/packages"
)

const (
	defaultAppFilename = "ko-app"
)

// GetBase takes an importpath and returns a base image reference and base image (or index).
type GetBase func(context.Context, string) (name.Reference, Result, error)

type (
	builder func(context.Context, string, string, v1.Platform, config.Config) (string, error)
	sbomber func(context.Context, string, string, v1.Image) ([]byte, types.MediaType, error)
)

type gobuild struct {
	ctx                  context.Context
	getBase              GetBase
	creationTime         v1.Time
	kodataCreationTime   v1.Time
	build                builder
	sbom                 sbomber
	disableOptimizations bool
	trimpath             bool
	buildConfigs         map[string]config.Config
	platformMatcher      *buildplatform.Matcher
	dir                  string
	labels               map[string]string
	semaphore            *semaphore.Weighted
	ociConversion        bool

	cache *layerCache
}

// Option is a functional option for NewGo.
type Option func(*gobuildOpener) error

type gobuildOpener struct {
	ctx                  context.Context
	getBase              GetBase
	creationTime         v1.Time
	kodataCreationTime   v1.Time
	build                builder
	sbom                 sbomber
	disableOptimizations bool
	trimpath             bool
	buildConfigs         map[string]config.Config
	platforms            []string
	labels               map[string]string
	dir                  string
	jobs                 int
	ociConversion        bool
}

func (gbo *gobuildOpener) Open() (Interface, error) {
	if gbo.getBase == nil {
		return nil, errors.New("a way of providing base images must be specified, see build.WithBaseImages")
	}
	matcher, err := buildplatform.ParseSpec(gbo.platforms)
	if err != nil {
		return nil, err
	}
	if gbo.jobs == 0 {
		gbo.jobs = runtime.GOMAXPROCS(0)
	}
	return &gobuild{
		ctx:                  gbo.ctx,
		getBase:              gbo.getBase,
		creationTime:         gbo.creationTime,
		kodataCreationTime:   gbo.kodataCreationTime,
		build:                gbo.build,
		sbom:                 gbo.sbom,
		disableOptimizations: gbo.disableOptimizations,
		trimpath:             gbo.trimpath,
		buildConfigs:         gbo.buildConfigs,
		labels:               gbo.labels,
		dir:                  gbo.dir,
		ociConversion:        gbo.ociConversion,
		platformMatcher:      matcher,
		cache: &layerCache{
			buildToDiff: map[string]buildIDToDiffID{},
			diffToDesc:  map[string]diffIDToDescriptor{},
		},
		semaphore: semaphore.NewWeighted(int64(gbo.jobs)),
	}, nil
}

// NewGo returns a build.Interface implementation that:
//  1. builds go binaries named by importpath,
//  2. containerizes the binary on a suitable base.
//
// The `dir` argument is the working directory for executing the `go` tool.
// If `dir` is empty, the function uses the current process working directory.
func NewGo(ctx context.Context, dir string, options ...Option) (Interface, error) {
	gbo := &gobuildOpener{
		ctx:   ctx,
		build: binary.Build,
		dir:   dir,
		sbom:  spdx("(none)"),
	}

	for _, option := range options {
		if err := option(gbo); err != nil {
			return nil, err
		}
	}
	return gbo.Open()
}

func (g *gobuild) qualifyLocalImport(importpath string) (string, error) {
	dir := filepath.Clean(g.dir)
	if dir == "." {
		dir = ""
	}
	cfg := &packages.Config{
		Mode: packages.NeedName,
		Dir:  dir,
	}
	pkgs, err := packages.Load(cfg, importpath)
	if err != nil {
		return "", err
	}
	if len(pkgs) != 1 {
		return "", fmt.Errorf("found %d local packages, expected 1", len(pkgs))
	}
	return pkgs[0].PkgPath, nil
}

// QualifyImport implements build.Interface
func (g *gobuild) QualifyImport(importpath string) (string, error) {
	if gb.IsLocalImport(importpath) {
		var err error
		importpath, err = g.qualifyLocalImport(importpath)
		if err != nil {
			return "", fmt.Errorf("qualifying local import %s: %w", importpath, err)
		}
	}
	if !strings.HasPrefix(importpath, StrictScheme) {
		importpath = StrictScheme + importpath
	}
	return importpath, nil
}

// IsSupportedReference implements build.Interface
//
// Only valid importpaths that provide commands (i.e., are "package main") are
// supported.
func (g *gobuild) IsSupportedReference(s string) error {
	ref := newRef(s)
	if !ref.IsStrict() {
		return errors.New("importpath does not start with ko://")
	}
	dir := filepath.Clean(g.dir)
	if dir == "." {
		dir = ""
	}
	pkgs, err := packages.Load(&packages.Config{Dir: dir, Mode: packages.NeedName}, ref.Path())
	if err != nil {
		return fmt.Errorf("error loading package from %s: %w", ref.Path(), err)
	}
	if len(pkgs) != 1 {
		return fmt.Errorf("found %d local packages, expected 1", len(pkgs))
	}
	if pkgs[0].Name != "main" {
		return errors.New("importpath is not `package main`")
	}
	return nil
}

func goversionm(ctx context.Context, file string, appPath string, _ v1.Image) ([]byte, types.MediaType, error) {
	sbom := bytes.NewBuffer(nil)
	cmd := exec.CommandContext(ctx, "go", "version", "-m", file)
	cmd.Stdout = sbom
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, "", err
	}

	// In order to get deterministics SBOMs replace our randomized
	// file name with the path the app will get inside of the container.
	return []byte(strings.Replace(sbom.String(), file, appPath, 1)), "application/vnd.go.version-m", nil
}

func spdx(version string) sbomber {
	return func(ctx context.Context, file string, appPath string, img v1.Image) ([]byte, types.MediaType, error) {
		b, _, err := goversionm(ctx, file, appPath, img)
		if err != nil {
			return nil, "", err
		}

		cfg, err := img.ConfigFile()
		if err != nil {
			return nil, "", err
		}
		imgDigest, err := img.Digest()
		if err != nil {
			return nil, "", err
		}
		b, err = sbom.GenerateSPDX(version, cfg.Created.Time, b, imgDigest)
		if err != nil {
			return nil, "", err
		}
		return b, ctypes.SPDXMediaType, nil
	}
}

func cycloneDX() sbomber {
	return func(ctx context.Context, file string, appPath string, img v1.Image) ([]byte, types.MediaType, error) {
		b, _, err := goversionm(ctx, file, appPath, img)
		if err != nil {
			return nil, "", err
		}

		b, err = sbom.GenerateCycloneDX(b)
		if err != nil {
			return nil, "", err
		}
		return b, ctypes.CycloneDXMediaType, nil
	}
}

func appFilename(importpath string) string {
	base := filepath.Base(importpath)

	// If we fail to determine a good name from the importpath then use a
	// safe default.
	if base == "." || base == string(filepath.Separator) {
		return defaultAppFilename
	}

	return base
}

func (g *gobuild) configForImportPathGo(ip string) config.Config {
	config := g.buildConfigs[ip]
	if g.trimpath {
		// The `-trimpath` flag removes file system paths from the resulting binary, to aid reproducibility.
		// Ref: https://pkg.go.dev/cmd/go#hdr-Compile_packages_and_dependencies
		config.Flags = append(config.Flags, "-trimpath")
	}

	if g.disableOptimizations {
		// Disable optimizations (-N) and inlining (-l).
		config.Flags = append(config.Flags, "-gcflags", "all=-N -l")
	}

	if config.ID != "" {
		log.Printf("Using build config %s for %s", config.ID, ip)
	}

	return config
}

func (g *gobuild) configForImportPathTinyGo(ip string) config.Config {
	config := g.buildConfigs[ip]
	if !g.disableOptimizations {
		config.Flags = append(config.Flags, "-opt", "2")
	}
	if config.ID != "" {
		log.Printf("Using build config %s for %s", config.ID, ip)
	}

	return config
}

const wasmImageAnnotationKey = "module.wasm.image/variant"

func (g *gobuild) buildOne(
	ctx context.Context,
	refStr string,
	base v1.Image,
	platform *v1.Platform,
) (oci.SignedImage, error) {
	if err := g.semaphore.Acquire(ctx, 1); err != nil {
		return nil, err
	}
	defer g.semaphore.Release(1)

	ref := newRef(refStr)

	cf, err := base.ConfigFile()
	if err != nil {
		return nil, err
	}
	if platform == nil {
		platform = &v1.Platform{
			OS:           cf.OS,
			Architecture: cf.Architecture,
			OSVersion:    cf.OSVersion,
		}
	}

	isWasiPlatform := buildplatform.IsWasi(platform)
	refPath := ref.Path()
	var config config.Config
	if isWasiPlatform {
		config = g.configForImportPathTinyGo(refPath)
	} else {
		config = g.configForImportPathGo(refPath)
	}
	// Do the build into a temporary file.
	file, err := g.build(ctx, refPath, g.dir, *platform, config)
	if err != nil {
		return nil, err
	}
	if os.Getenv("KOCACHE") == "" {
		defer os.RemoveAll(filepath.Dir(file))
	}

	var layers []mutate.Addendum
	layerOpts := []tarball.LayerOption{
		tarball.WithCompressedCaching,
	}

	if g.ociConversion {
		layerOpts = append(layerOpts, tarball.WithMediaType(types.OCILayer))
	}

	// don't add kodata layer into wasi image
	if !isWasiPlatform {
		// Create a layer from the kodata directory under this import path.
		dataLayerBuf, err := tarKoData(g.dir, ref, platform, g.creationTime)
		if err != nil {
			return nil, err
		}

		dataLayerBytes := dataLayerBuf.Bytes()
		dataLayer, err := tarball.LayerFromOpener(func() (io.ReadCloser, error) {
			return ioutil.NopCloser(bytes.NewBuffer(dataLayerBytes)), nil
		}, layerOpts...)
		if err != nil {
			return nil, err
		}
		layers = append(layers, mutate.Addendum{
			Layer: dataLayer,
			History: v1.History{
				Author:    "ko",
				CreatedBy: "ko build " + ref.String(),
				Created:   g.kodataCreationTime,
				Comment:   "kodata contents, at $KO_DATA_PATH",
			},
		})
	}

	appDir := "/ko-app"
	filename := appFilename(refPath)
	if isWasiPlatform {
		// module.wasm.image/variant=compat-smart expects the entrypoint binary to be of form .wasm
		filename += ".wasm"
	}

	appPath := path.Join(appDir, filename)

	miss := func() (v1.Layer, error) {
		// When using estargz, prioritize downloading the binary entrypoint.
		layerOpts = append(layerOpts, tarball.WithEstargzOptions(estargz.WithPrioritizedFiles([]string{
			appPath,
		})))
		return buildLayer(appPath, file, platform, layerOpts...)
	}

	binaryLayer, err := g.cache.get(ctx, file, miss)
	if err != nil {
		return nil, err
	}

	binaryLayerAddendum := mutate.Addendum{
		Layer: binaryLayer,
		History: v1.History{
			Author:    "ko",
			Created:   g.creationTime,
			CreatedBy: "ko build " + ref.String(),
			Comment:   "go build output, at " + appPath,
		},
	}

	layers = append(layers, binaryLayerAddendum)

	// Augment the base image with our application layer.
	withApp, err := mutate.Append(base, layers...)
	if err != nil {
		return nil, err
	}

	// Start from a copy of the base image's config file, and set
	// the entrypoint to our app.
	cfg, err := withApp.ConfigFile()
	if err != nil {
		return nil, err
	}

	cfg = cfg.DeepCopy()
	cfg.Config.Entrypoint = []string{appPath}
	cfg.Config.Cmd = nil
	if platform.OS == "windows" {
		cfg.Config.Entrypoint = []string{`C:\ko-app\` + appFilename(ref.Path())}
		updatePath(cfg, `C:\ko-app`)
		cfg.Config.Env = append(cfg.Config.Env, `KO_DATA_PATH=C:\var\run\ko`)
	} else if platform.String() != "wasm/wasi" {
		updatePath(cfg, appDir)
		cfg.Config.Env = append(cfg.Config.Env, "KO_DATA_PATH="+kodataRoot)
	}
	cfg.Author = "github.com/google/ko"

	if cfg.Config.Labels == nil {
		cfg.Config.Labels = map[string]string{}
	}
	for k, v := range g.labels {
		cfg.Config.Labels[k] = v
	}

	empty := v1.Time{}
	if g.creationTime != empty {
		cfg.Created = g.creationTime
	}

	image, err := mutate.ConfigFile(withApp, cfg)
	if err != nil {
		return nil, err
	}

	// if g.ociConversion {
	//   image = mutate.ConfigMediaType(image, types.OCIConfigJSON)
	//   image = mutate.MediaType(image, types.OCIManifestSchema1)
	// }

	if isWasiPlatform {
		image = mutate.Annotations(image, map[string]string{
			wasmImageAnnotationKey: "compat-smart",
		}).(v1.Image)
	}

	log.Printf("%#v", image)
	si, err := generateSbom(ctx, file, appPath, g.sbom, signed.Image(image))
	if err != nil {
		return nil, fmt.Errorf("generating sbom: %w", err)
	}
	log.Printf("%#v", si)
	return si, nil
}

func generateSbom(
	ctx context.Context,
	file string,
	appPath string,
	sbomFunc sbomber,
	si oci.SignedImage,
) (oci.SignedImage, error) {
	if sbomFunc == nil {
		return si, nil
	}

	sbom, mt, err := sbomFunc(ctx, file, appPath, si)
	if err != nil {
		return nil, err
	}
	f, err := static.NewFile(sbom, static.WithLayerMediaType(mt))
	if err != nil {
		return nil, err
	}
	si, err = ocimutate.AttachFileToImage(si, "sbom", f)
	if err != nil {
		return nil, err
	}
	return si, nil
}

func buildLayer(appPath, file string, platform *v1.Platform, layerOpts ...tarball.LayerOption) (v1.Layer, error) {
	// Construct a tarball with the binary and produce a layer.
	binaryLayerBuf, err := tarBinary(appPath, file, platform)
	if err != nil {
		return nil, err
	}
	binaryLayerBytes := binaryLayerBuf.Bytes()
	return tarball.LayerFromOpener(func() (io.ReadCloser, error) {
		return ioutil.NopCloser(bytes.NewBuffer(binaryLayerBytes)), nil
	}, layerOpts...)
}

// Append appPath to the PATH environment variable, if it exists. Otherwise,
// set the PATH environment variable to appPath.
func updatePath(cf *v1.ConfigFile, appPath string) {
	for i, env := range cf.Config.Env {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			// Expect environment variables to be in the form KEY=VALUE, so this is unexpected.
			continue
		}
		key, value := parts[0], parts[1]
		if key == "PATH" {
			value = fmt.Sprintf("%s:%s", value, appPath)
			cf.Config.Env[i] = "PATH=" + value
			return
		}
	}

	// If we get here, we never saw PATH.
	cf.Config.Env = append(cf.Config.Env, "PATH="+appPath)
}

// Build implements build.Interface
func (g *gobuild) Build(ctx context.Context, s string) (Result, error) {
	// Determine the appropriate base image for this import path.
	// We use the overall gobuild.ctx because the Build ctx gets cancelled
	// early, and we lazily use the ctx within ggcr's remote package.
	baseRef, base, err := g.getBase(g.ctx, s)
	if err != nil {
		return nil, err
	}

	// Determine what kind of base we have and if we should publish an image or an index.
	mt, err := base.MediaType()
	if err != nil {
		return nil, err
	}

	// Annotate the base image we pass to the build function with
	// annotations indicating the digest (and possibly tag) of the
	// base image.  This will be inherited by the image produced.
	if mt != types.DockerManifestList {
		baseDigest, err := base.Digest()
		if err != nil {
			return nil, err
		}

		anns := map[string]string{
			specsv1.AnnotationBaseImageDigest: baseDigest.String(),
		}
		if _, ok := baseRef.(name.Tag); ok {
			anns[specsv1.AnnotationBaseImageName] = baseRef.Name()
		}
		base = mutate.Annotations(base, anns).(Result)
	}

	var res Result
	switch mt {
	case types.OCIImageIndex, types.DockerManifestList:
		baseIndex, ok := base.(v1.ImageIndex)
		if !ok {
			return nil, fmt.Errorf("failed to interpret base as index: %v", base)
		}
		res, err = g.buildAll(ctx, s, baseIndex)
	case types.OCIManifestSchema1, types.DockerManifestSchema2:
		baseImage, ok := base.(v1.Image)
		if !ok {
			return nil, fmt.Errorf("failed to interpret base as image: %v", base)
		}
		// user specified a platform, so force this platform, even if it fails
		var platform *v1.Platform
		if len(g.platformMatcher.Platforms) == 1 {
			platform = &g.platformMatcher.Platforms[0]
		}
		res, err = g.buildOne(ctx, s, baseImage, platform)
	default:
		return nil, fmt.Errorf("base image media type: %s", mt)
	}
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (g *gobuild) buildAll(ctx context.Context, ref string, baseIndex v1.ImageIndex) (Result, error) {
	im, err := baseIndex.IndexManifest()
	if err != nil {
		return nil, err
	}

	matches := []v1.Descriptor{}
	for _, desc := range im.Manifests {
		// Nested index is pretty rare. We could support this in theory, but return an error for now.
		if desc.MediaType != types.OCIManifestSchema1 && desc.MediaType != types.DockerManifestSchema2 {
			return nil, fmt.Errorf("%q has unexpected mediaType %q in base for %q", desc.Digest, desc.MediaType, ref)
		}

		if g.platformMatcher.Matches(desc.Platform) {
			matches = append(matches, desc)
		}
	}
	// no base image will ever have wasiPlatform set in their index, which is why we need to specially look for it
	if g.platformMatcher.Matches(buildplatform.Wasi) {
		matches = append(matches, v1.Descriptor{
      Platform: buildplatform.Wasi,
    })
	}

	if len(matches) == 0 {
		return nil, errors.New("no matching platforms in base image index")
	}

	// Build an image for each matching platform from the base and append
	// it to a new index to produce the result. We use the indices to
	// preserve the base image ordering here.
	errg, ctx := errgroup.WithContext(ctx)
	adds := make([]ocimutate.IndexAddendum, len(matches))
	for i, desc := range matches {
		i, desc := i, desc
		errg.Go(func() error {
			var baseImage v1.Image
			if buildplatform.IsWasi(desc.Platform) {
				baseImage = empty.Image
			} else {
				baseImage, err = baseIndex.Image(desc.Digest)
				if err != nil {
					return err
				}
			}
			if g.ociConversion {
        log.Println("converting baseImage to oci format")
				baseImage, err = buildoci.ConvertImage(baseImage)
        if err != nil {
          return err
        }
			}
			img, err := g.buildOne(ctx, ref, baseImage, desc.Platform)
			if err != nil {
				return err
			}
			adds[i] = ocimutate.IndexAddendum{
				Add: img,
				Descriptor: v1.Descriptor{
					URLs:        desc.URLs,
					MediaType:   desc.MediaType,
					Annotations: desc.Annotations,
					Platform:    desc.Platform,
				},
			}
			return nil
		})
	}
	if err := errg.Wait(); err != nil {
		return nil, err
	}

	baseType, err := baseIndex.MediaType()
	if err != nil {
		return nil, err
	}

	idx := ocimutate.AppendManifests(
		mutate.Annotations(
			mutate.IndexMediaType(empty.Index, baseType),
			im.Annotations).(v1.ImageIndex),
		adds...)

	// TODO(mattmoor): If we want to attach anything (e.g. signatures, attestations, SBOM)
	// at the index level, we would do it here!
	return idx, nil
}
