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
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
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
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

const (
	appDir             = "/ko-app"
	defaultAppFilename = "ko-app"
)

// GetBase takes an importpath and returns a base image.
type GetBase func(string) (Result, error)

type builder func(context.Context, string, v1.Platform, bool) (string, error)

type buildContext interface {
	Import(path string, srcDir string, mode gb.ImportMode) (*gb.Package, error)
}

type gobuild struct {
	getBase              GetBase
	creationTime         v1.Time
	build                builder
	disableOptimizations bool
	mod                  *modules
	buildContext         buildContext
}

// Option is a functional option for NewGo.
type Option func(*gobuildOpener) error

type gobuildOpener struct {
	getBase              GetBase
	creationTime         v1.Time
	build                builder
	disableOptimizations bool
	mod                  *modules
	buildContext         buildContext
}

func (gbo *gobuildOpener) Open() (Interface, error) {
	if gbo.getBase == nil {
		return nil, errors.New("a way of providing base images must be specified, see build.WithBaseImages")
	}
	return &gobuild{
		getBase:              gbo.getBase,
		creationTime:         gbo.creationTime,
		build:                gbo.build,
		disableOptimizations: gbo.disableOptimizations,
		mod:                  gbo.mod,
		buildContext:         gbo.buildContext,
	}, nil
}

// https://golang.org/pkg/cmd/go/internal/modinfo/#ModulePublic
type modules struct {
	main *modInfo
	deps map[string]*modInfo
}

type modInfo struct {
	Path string
	Dir  string
	Main bool
}

// moduleInfo returns the module path and module root directory for a project
// using go modules, otherwise returns nil.
//
// Related: https://github.com/golang/go/issues/26504
func moduleInfo() (*modules, error) {
	modules := modules{
		deps: make(map[string]*modInfo),
	}

	// TODO we read all the output as a single byte array - it may
	// be possible & more efficient to stream it
	output, err := exec.Command("go", "list", "-mod=readonly", "-json", "-m", "all").Output()
	if err != nil {
		return nil, nil
	}

	dec := json.NewDecoder(bytes.NewReader(output))

	for {
		var info modInfo

		err := dec.Decode(&info)
		if err == io.EOF {
			// all done
			break
		}

		modules.deps[info.Path] = &info

		if info.Main {
			modules.main = &info
		}

		if err != nil {
			return nil, fmt.Errorf("error reading module data %w", err)
		}
	}

	if modules.main == nil {
		return nil, fmt.Errorf("couldn't find main module")
	}

	return &modules, nil
}

// NewGo returns a build.Interface implementation that:
//  1. builds go binaries named by importpath,
//  2. containerizes the binary on a suitable base,
func NewGo(options ...Option) (Interface, error) {
	module, err := moduleInfo()
	if err != nil {
		return nil, err
	}

	gbo := &gobuildOpener{
		build:        build,
		mod:          module,
		buildContext: &gb.Default,
	}

	for _, option := range options {
		if err := option(gbo); err != nil {
			return nil, err
		}
	}
	return gbo.Open()
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
	p, err := g.importPackage(ref)
	if err != nil {
		return err
	}
	if !p.IsCommand() {
		return errors.New("importpath is not `package main`")
	}
	return nil
}

// importPackage wraps go/build.Import to handle go modules.
//
// Note that we will fall back to GOPATH if the project isn't using go modules.
func (g *gobuild) importPackage(ref reference) (*gb.Package, error) {
	if g.mod == nil {
		return g.buildContext.Import(ref.Path(), gb.Default.GOPATH, gb.ImportComment)
	}

	// If we're inside a go modules project, try to use the module's directory
	// as our source root to import:
	// * any strict reference we get
	// * paths that match module path prefix (they should be in this project)
	// * relative paths (they should also be in this project)
	// * path is a module

	_, isDep := g.mod.deps[ref.Path()]
	if ref.IsStrict() || strings.HasPrefix(ref.Path(), g.mod.main.Path) || gb.IsLocalImport(ref.Path()) || isDep {
		return g.buildContext.Import(ref.Path(), g.mod.main.Dir, gb.ImportComment)
	}

	return nil, fmt.Errorf("unmatched importPackage %q with gomodules", ref.String())
}

func build(ctx context.Context, ip string, platform v1.Platform, disableOptimizations bool) (string, error) {
	tmpDir, err := ioutil.TempDir("", "ko")
	if err != nil {
		return "", err
	}
	file := filepath.Join(tmpDir, "out")

	args := make([]string, 0, 7)
	args = append(args, "build")
	if disableOptimizations {
		// Disable optimizations (-N) and inlining (-l).
		args = append(args, "-gcflags", "all=-N -l")
	}
	args = append(args, "-o", file)
	args = addGo113TrimPathFlag(args)
	args = append(args, ip)
	cmd := exec.CommandContext(ctx, "go", args...)

	// Last one wins
	defaultEnv := []string{
		"CGO_ENABLED=0",
		"GOOS=" + platform.OS,
		"GOARCH=" + platform.Architecture,
	}
	cmd.Env = append(defaultEnv, os.Environ()...)

	var output bytes.Buffer
	cmd.Stderr = &output
	cmd.Stdout = &output

	log.Printf("Building %s for %s/%s", ip, platform.OS, platform.Architecture)
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tmpDir)
		log.Printf("Unexpected error running \"go build\": %v\n%v", err, output.String())
		return "", err
	}
	return file, nil
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

func tarAddDirectories(tw *tar.Writer, dir string) error {
	if dir == "." || dir == string(filepath.Separator) {
		return nil
	}

	// Write parent directories first
	if err := tarAddDirectories(tw, filepath.Dir(dir)); err != nil {
		return err
	}

	// write the directory header to the tarball archive
	if err := tw.WriteHeader(&tar.Header{
		Name:     dir,
		Typeflag: tar.TypeDir,
		// Use a fixed Mode, so that this isn't sensitive to the directory and umask
		// under which it was created. Additionally, windows can only set 0222,
		// 0444, or 0666, none of which are executable.
		Mode: 0555,
	}); err != nil {
		return err
	}

	return nil
}

func tarBinary(name, binary string) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)
	// Compress this before calling tarball.LayerFromOpener, since it eagerly
	// calculates digests and diffids. This prevents us from double compressing
	// the layer when we have to actually upload the blob.
	//
	// https://github.com/google/go-containerregistry/issues/413
	gw, _ := gzip.NewWriterLevel(buf, gzip.BestSpeed)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// write the parent directories to the tarball archive
	if err := tarAddDirectories(tw, path.Dir(name)); err != nil {
		return nil, err
	}

	file, err := os.Open(binary)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	header := &tar.Header{
		Name:     name,
		Size:     stat.Size(),
		Typeflag: tar.TypeReg,
		// Use a fixed Mode, so that this isn't sensitive to the directory and umask
		// under which it was created. Additionally, windows can only set 0222,
		// 0444, or 0666, none of which are executable.
		Mode: 0555,
	}
	// write the header to the tarball archive
	if err := tw.WriteHeader(header); err != nil {
		return nil, err
	}
	// copy the file data to the tarball
	if _, err := io.Copy(tw, file); err != nil {
		return nil, err
	}

	return buf, nil
}

func (g *gobuild) kodataPath(ref reference) (string, error) {
	p, err := g.importPackage(ref)
	if err != nil {
		return "", err
	}
	return filepath.Join(p.Dir, "kodata"), nil
}

// Where kodata lives in the image.
const kodataRoot = "/var/run/ko"

// walkRecursive performs a filepath.Walk of the given root directory adding it
// to the provided tar.Writer with root -> chroot.  All symlinks are dereferenced,
// which is what leads to recursion when we encounter a directory symlink.
func walkRecursive(tw *tar.Writer, root, chroot string) error {
	return filepath.Walk(root, func(hostPath string, info os.FileInfo, err error) error {
		if hostPath == root {
			// Add an entry for the root directory of our walk.
			return tw.WriteHeader(&tar.Header{
				Name:     chroot,
				Typeflag: tar.TypeDir,
				// Use a fixed Mode, so that this isn't sensitive to the directory and umask
				// under which it was created. Additionally, windows can only set 0222,
				// 0444, or 0666, none of which are executable.
				Mode: 0555,
			})
		}
		if err != nil {
			return err
		}
		// Skip other directories.
		if info.Mode().IsDir() {
			return nil
		}
		newPath := path.Join(chroot, filepath.ToSlash(hostPath[len(root):]))

		hostPath, err = filepath.EvalSymlinks(hostPath)
		if err != nil {
			return err
		}

		// Chase symlinks.
		info, err = os.Stat(hostPath)
		if err != nil {
			return err
		}
		// Skip other directories.
		if info.Mode().IsDir() {
			return walkRecursive(tw, hostPath, newPath)
		}

		// Open the file to copy it into the tarball.
		file, err := os.Open(hostPath)
		if err != nil {
			return err
		}
		defer file.Close()

		// Copy the file into the image tarball.
		if err := tw.WriteHeader(&tar.Header{
			Name:     newPath,
			Size:     info.Size(),
			Typeflag: tar.TypeReg,
			// Use a fixed Mode, so that this isn't sensitive to the directory and umask
			// under which it was created. Additionally, windows can only set 0222,
			// 0444, or 0666, none of which are executable.
			Mode: 0555,
		}); err != nil {
			return err
		}
		_, err = io.Copy(tw, file)
		return err
	})
}

func (g *gobuild) tarKoData(ref reference) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)
	// Compress this before calling tarball.LayerFromOpener, since it eagerly
	// calculates digests and diffids. This prevents us from double compressing
	// the layer when we have to actually upload the blob.
	//
	// https://github.com/google/go-containerregistry/issues/413
	gw, _ := gzip.NewWriterLevel(buf, gzip.BestSpeed)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	root, err := g.kodataPath(ref)
	if err != nil {
		return nil, err
	}

	return buf, walkRecursive(tw, root, kodataRoot)
}

func (g *gobuild) buildOne(ctx context.Context, s string, base v1.Image) (v1.Image, error) {
	ref := newRef(s)

	cf, err := base.ConfigFile()
	if err != nil {
		return nil, err
	}
	platform := v1.Platform{
		OS:           cf.OS,
		Architecture: cf.Architecture,
	}

	// Do the build into a temporary file.
	file, err := g.build(ctx, ref.Path(), platform, g.disableOptimizations)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(filepath.Dir(file))

	var layers []mutate.Addendum
	// Create a layer from the kodata directory under this import path.
	dataLayerBuf, err := g.tarKoData(ref)
	if err != nil {
		return nil, err
	}
	dataLayerBytes := dataLayerBuf.Bytes()
	dataLayer, err := tarball.LayerFromOpener(func() (io.ReadCloser, error) {
		return ioutil.NopCloser(bytes.NewBuffer(dataLayerBytes)), nil
	})
	if err != nil {
		return nil, err
	}
	layers = append(layers, mutate.Addendum{
		Layer: dataLayer,
		History: v1.History{
			Author:    "ko",
			CreatedBy: "ko publish " + ref.String(),
			Comment:   "kodata contents, at $KO_DATA_PATH",
		},
	})

	appPath := path.Join(appDir, appFilename(ref.Path()))

	// Construct a tarball with the binary and produce a layer.
	binaryLayerBuf, err := tarBinary(appPath, file)
	if err != nil {
		return nil, err
	}
	binaryLayerBytes := binaryLayerBuf.Bytes()
	binaryLayer, err := tarball.LayerFromOpener(func() (io.ReadCloser, error) {
		return ioutil.NopCloser(bytes.NewBuffer(binaryLayerBytes)), nil
	})
	if err != nil {
		return nil, err
	}
	layers = append(layers, mutate.Addendum{
		Layer: binaryLayer,
		History: v1.History{
			Author:    "ko",
			CreatedBy: "ko publish " + ref.String(),
			Comment:   "go build output, at " + appPath,
		},
	})

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
	updatePath(cfg)
	cfg.Config.Env = append(cfg.Config.Env, "KO_DATA_PATH="+kodataRoot)
	cfg.Author = "github.com/google/ko"

	image, err := mutate.ConfigFile(withApp, cfg)
	if err != nil {
		return nil, err
	}

	empty := v1.Time{}
	if g.creationTime != empty {
		return mutate.CreatedAt(image, g.creationTime)
	}
	return image, nil
}

// Append appDir to the PATH environment variable, if it exists. Otherwise,
// set the PATH environment variable to appDir.
func updatePath(cf *v1.ConfigFile) {
	for i, env := range cf.Config.Env {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			// Expect environment variables to be in the form KEY=VALUE, so this is unexpected.
			continue
		}
		key, value := parts[0], parts[1]
		if key == "PATH" {
			value = fmt.Sprintf("%s:%s", value, appDir)
			cf.Config.Env[i] = "PATH=" + value
			return
		}
	}

	// If we get here, we never saw PATH.
	cf.Config.Env = append(cf.Config.Env, "PATH="+appDir)
}

// Build implements build.Interface
func (g *gobuild) Build(ctx context.Context, s string) (Result, error) {
	// Determine the appropriate base image for this import path.
	base, err := g.getBase(s)
	if err != nil {
		return nil, err
	}

	// Determine what kind of base we have and if we should publish an image or an index.
	mt, err := base.MediaType()
	if err != nil {
		return nil, err
	}

	switch mt {
	case types.OCIImageIndex, types.DockerManifestList:
		base, ok := base.(v1.ImageIndex)
		if !ok {
			return nil, fmt.Errorf("failed to interpret base as index: %v", base)
		}
		return g.buildAll(ctx, s, base)
	case types.OCIManifestSchema1, types.DockerManifestSchema2:
		base, ok := base.(v1.Image)
		if !ok {
			return nil, fmt.Errorf("failed to interpret base as image: %v", base)
		}
		return g.buildOne(ctx, s, base)
	default:
		return nil, fmt.Errorf("base image media type: %s", mt)
	}
}

// TODO(#192): Do these in parallel?
func (g *gobuild) buildAll(ctx context.Context, s string, base v1.ImageIndex) (v1.ImageIndex, error) {
	im, err := base.IndexManifest()
	if err != nil {
		return nil, err
	}

	// Build an image for each child from the base and append it to a new index to produce the result.
	adds := []mutate.IndexAddendum{}
	for _, desc := range im.Manifests {
		// Nested index is pretty rare. We could support this in theory, but return an error for now.
		if desc.MediaType != types.OCIManifestSchema1 && desc.MediaType != types.DockerManifestSchema2 {
			return nil, fmt.Errorf("%q has unexpected mediaType %q in base for %q", desc.Digest, desc.MediaType, s)
		}

		base, err := base.Image(desc.Digest)
		if err != nil {
			return nil, err
		}
		img, err := g.buildOne(ctx, s, base)
		if err != nil {
			return nil, err
		}
		adds = append(adds, mutate.IndexAddendum{
			Add: img,
			Descriptor: v1.Descriptor{
				URLs:        desc.URLs,
				MediaType:   desc.MediaType,
				Annotations: desc.Annotations,
				Platform:    desc.Platform,
			},
		})
	}

	baseType, err := base.MediaType()
	if err != nil {
		return nil, err
	}

	return mutate.IndexMediaType(mutate.AppendManifests(empty.Index, adds...), baseType), nil
}
