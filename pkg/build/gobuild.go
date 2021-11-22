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
	"archive/tar"
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
	"strconv"
	"strings"
	"text/template"

	"github.com/containerd/stargz-snapshotter/estargz"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/google/ko/internal/sbom"
	specsv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sigstore/cosign/pkg/oci"
	ocimutate "github.com/sigstore/cosign/pkg/oci/mutate"
	"github.com/sigstore/cosign/pkg/oci/signed"
	"github.com/sigstore/cosign/pkg/oci/static"
	ctypes "github.com/sigstore/cosign/pkg/types"
	"golang.org/x/tools/go/packages"
)

const (
	defaultAppFilename = "ko-app"
)

// GetBase takes an importpath and returns a base image reference and base image (or index).
type GetBase func(context.Context, string) (name.Reference, Result, error)

type builder func(context.Context, string, string, v1.Platform, Config) (string, error)
type sbomber func(context.Context, string, string, v1.Image) ([]byte, types.MediaType, error)

type platformMatcher struct {
	spec      string
	platforms []v1.Platform
}

type gobuild struct {
	getBase              GetBase
	creationTime         v1.Time
	kodataCreationTime   v1.Time
	build                builder
	sbom                 sbomber
	disableOptimizations bool
	trimpath             bool
	buildConfigs         map[string]Config
	platformMatcher      *platformMatcher
	dir                  string
	labels               map[string]string
}

// Option is a functional option for NewGo.
type Option func(*gobuildOpener) error

type gobuildOpener struct {
	getBase              GetBase
	creationTime         v1.Time
	kodataCreationTime   v1.Time
	build                builder
	sbom                 sbomber
	disableOptimizations bool
	trimpath             bool
	buildConfigs         map[string]Config
	platform             string
	labels               map[string]string
	dir                  string
}

func (gbo *gobuildOpener) Open() (Interface, error) {
	if gbo.getBase == nil {
		return nil, errors.New("a way of providing base images must be specified, see build.WithBaseImages")
	}
	matcher, err := parseSpec(gbo.platform)
	if err != nil {
		return nil, err
	}
	return &gobuild{
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
		platformMatcher:      matcher,
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
		build: build,
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
	cfg := &packages.Config{
		Mode: packages.NeedName,
		Dir:  g.dir,
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
	pkgs, err := packages.Load(&packages.Config{Dir: g.dir, Mode: packages.NeedName}, ref.Path())
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

func getGoarm(platform v1.Platform) (string, error) {
	if !strings.HasPrefix(platform.Variant, "v") {
		return "", fmt.Errorf("strange arm variant: %v", platform.Variant)
	}

	vs := strings.TrimPrefix(platform.Variant, "v")
	variant, err := strconv.Atoi(vs)
	if err != nil {
		return "", fmt.Errorf("cannot parse arm variant %q: %w", platform.Variant, err)
	}
	if variant >= 5 {
		// TODO(golang/go#29373): Allow for 8 in later go versions if this is fixed.
		if variant > 7 {
			vs = "7"
		}
		return vs, nil
	}
	return "", nil
}

// TODO(jonjohnsonjr): Upstream something like this.
func platformToString(p v1.Platform) string {
	if p.Variant != "" {
		return fmt.Sprintf("%s/%s/%s", p.OS, p.Architecture, p.Variant)
	}
	return fmt.Sprintf("%s/%s", p.OS, p.Architecture)
}

func build(ctx context.Context, ip string, dir string, platform v1.Platform, config Config) (string, error) {
	tmpDir, err := ioutil.TempDir("", "ko")
	if err != nil {
		return "", err
	}
	file := filepath.Join(tmpDir, "out")

	buildArgs, err := createBuildArgs(config)
	if err != nil {
		return "", err
	}

	args := make([]string, 0, 4+len(buildArgs))
	args = append(args, "build")
	args = append(args, buildArgs...)
	args = append(args, "-o", file)
	args = append(args, ip)
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = dir

	env, err := buildEnv(platform, os.Environ(), config.Env)
	if err != nil {
		return "", fmt.Errorf("could not create env for %s: %w", ip, err)
	}
	cmd.Env = env

	var output bytes.Buffer
	cmd.Stderr = &output
	cmd.Stdout = &output

	log.Printf("Building %s for %s", ip, platformToString(platform))
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tmpDir)
		log.Printf("Unexpected error running \"go build\": %v\n%v", err, output.String())
		return "", err
	}
	return file, nil
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

		b, err = sbom.GenerateSPDX(version, cfg.Created.Time, b)
		if err != nil {
			return nil, "", err
		}
		return b, ctypes.SPDXMediaType, nil
	}
}

// buildEnv creates the environment variables used by the `go build` command.
// From `os/exec.Cmd`: If Env contains duplicate environment keys, only the last
// value in the slice for each duplicate key is used.
func buildEnv(platform v1.Platform, userEnv, configEnv []string) ([]string, error) {
	// Default env
	env := []string{
		"CGO_ENABLED=0",
		"GOOS=" + platform.OS,
		"GOARCH=" + platform.Architecture,
	}

	if strings.HasPrefix(platform.Architecture, "arm") && platform.Variant != "" {
		goarm, err := getGoarm(platform)
		if err != nil {
			return nil, fmt.Errorf("goarm failure: %w", err)
		}
		if goarm != "" {
			env = append(env, "GOARM="+goarm)
		}
	}

	env = append(env, userEnv...)
	env = append(env, configEnv...)
	return env, nil
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

// userOwnerAndGroupSID is a magic value needed to make the binary executable
// in a Windows container.
//
// owner: BUILTIN/Users group: BUILTIN/Users ($sddlValue="O:BUG:BU")
const userOwnerAndGroupSID = "AQAAgBQAAAAkAAAAAAAAAAAAAAABAgAAAAAABSAAAAAhAgAAAQIAAAAAAAUgAAAAIQIAAA=="

func tarBinary(name, binary string, creationTime v1.Time, platform *v1.Platform) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	// Write the parent directories to the tarball archive.
	// For Windows, the layer must contain a Hives/ directory, and the root
	// of the actual filesystem goes in a Files/ directory.
	// For Linux, the binary goes into /ko-app/
	dirs := []string{"ko-app"}
	if platform.OS == "windows" {
		dirs = []string{
			"Hives",
			"Files",
			"Files/ko-app",
		}
		name = "Files" + name
	}
	for _, dir := range dirs {
		if err := tw.WriteHeader(&tar.Header{
			Name:     dir,
			Typeflag: tar.TypeDir,
			// Use a fixed Mode, so that this isn't sensitive to the directory and umask
			// under which it was created. Additionally, windows can only set 0222,
			// 0444, or 0666, none of which are executable.
			Mode:    0555,
			ModTime: creationTime.Time,
		}); err != nil {
			return nil, fmt.Errorf("writing dir %q: %w", dir, err)
		}
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
		Mode:    0555,
		ModTime: creationTime.Time,
	}
	if platform.OS == "windows" {
		// This magic value is for some reason needed for Windows to be
		// able to execute the binary.
		header.PAXRecords = map[string]string{
			"MSWINDOWS.rawsd": userOwnerAndGroupSID,
		}
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
	pkgs, err := packages.Load(&packages.Config{Dir: g.dir, Mode: packages.NeedFiles}, ref.Path())
	if err != nil {
		return "", fmt.Errorf("error loading package from %s: %w", ref.Path(), err)
	}
	if len(pkgs) != 1 {
		return "", fmt.Errorf("found %d local packages, expected 1", len(pkgs))
	}
	if len(pkgs[0].GoFiles) == 0 {
		return "", fmt.Errorf("package %s contains no Go files", pkgs[0])
	}
	return filepath.Join(filepath.Dir(pkgs[0].GoFiles[0]), "kodata"), nil
}

// Where kodata lives in the image.
const kodataRoot = "/var/run/ko"

// walkRecursive performs a filepath.Walk of the given root directory adding it
// to the provided tar.Writer with root -> chroot.  All symlinks are dereferenced,
// which is what leads to recursion when we encounter a directory symlink.
func walkRecursive(tw *tar.Writer, root, chroot string, creationTime v1.Time, platform *v1.Platform) error {
	return filepath.Walk(root, func(hostPath string, info os.FileInfo, err error) error {
		if hostPath == root {
			return nil
		}
		if err != nil {
			return fmt.Errorf("filepath.Walk(%q): %w", root, err)
		}
		// Skip other directories.
		if info.Mode().IsDir() {
			return nil
		}
		newPath := path.Join(chroot, filepath.ToSlash(hostPath[len(root):]))

		// Don't chase symlinks on Windows, where cross-compiled symlink support is not possible.
		if platform.OS == "windows" {
			if info.Mode()&os.ModeSymlink != 0 {
				log.Println("skipping symlink in kodata for windows:", info.Name())
				return nil
			}
		}

		evalPath, err := filepath.EvalSymlinks(hostPath)
		if err != nil {
			return fmt.Errorf("filepath.EvalSymlinks(%q): %w", hostPath, err)
		}

		// Chase symlinks.
		info, err = os.Stat(evalPath)
		if err != nil {
			return fmt.Errorf("os.Stat(%q): %w", evalPath, err)
		}
		// Skip other directories.
		if info.Mode().IsDir() {
			return walkRecursive(tw, evalPath, newPath, creationTime, platform)
		}

		// Open the file to copy it into the tarball.
		file, err := os.Open(evalPath)
		if err != nil {
			return fmt.Errorf("os.Open(%q): %w", evalPath, err)
		}
		defer file.Close()

		// Copy the file into the image tarball.
		header := &tar.Header{
			Name:     newPath,
			Size:     info.Size(),
			Typeflag: tar.TypeReg,
			// Use a fixed Mode, so that this isn't sensitive to the directory and umask
			// under which it was created. Additionally, windows can only set 0222,
			// 0444, or 0666, none of which are executable.
			Mode:    0555,
			ModTime: creationTime.Time,
		}
		if platform.OS == "windows" {
			// This magic value is for some reason needed for Windows to be
			// able to execute the binary.
			header.PAXRecords = map[string]string{
				"MSWINDOWS.rawsd": userOwnerAndGroupSID,
			}
		}
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("tar.Writer.WriteHeader(%q): %w", newPath, err)
		}
		if _, err := io.Copy(tw, file); err != nil {
			return fmt.Errorf("io.Copy(%q, %q): %w", newPath, evalPath, err)
		}
		return nil
	})
}

func (g *gobuild) tarKoData(ref reference, platform *v1.Platform) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	root, err := g.kodataPath(ref)
	if err != nil {
		return nil, err
	}

	creationTime := g.kodataCreationTime

	// Write the parent directories to the tarball archive.
	// For Windows, the layer must contain a Hives/ directory, and the root
	// of the actual filesystem goes in a Files/ directory.
	// For Linux, kodata starts at /var/run/ko.
	chroot := kodataRoot
	dirs := []string{
		"/var",
		"/var/run",
		"/var/run/ko",
	}
	if platform.OS == "windows" {
		chroot = "Files" + kodataRoot
		dirs = []string{
			"Hives",
			"Files",
			"Files/var",
			"Files/var/run",
			"Files/var/run/ko",
		}
	}
	for _, dir := range dirs {
		if err := tw.WriteHeader(&tar.Header{
			Name:     dir,
			Typeflag: tar.TypeDir,
			// Use a fixed Mode, so that this isn't sensitive to the directory and umask
			// under which it was created. Additionally, windows can only set 0222,
			// 0444, or 0666, none of which are executable.
			Mode:    0555,
			ModTime: creationTime.Time,
		}); err != nil {
			return nil, fmt.Errorf("writing dir %q: %w", dir, err)
		}
	}

	return buf, walkRecursive(tw, root, chroot, creationTime, platform)
}

func createTemplateData() map[string]interface{} {
	envVars := map[string]string{}
	for _, entry := range os.Environ() {
		kv := strings.SplitN(entry, "=", 2)
		envVars[kv[0]] = kv[1]
	}

	return map[string]interface{}{
		"Env": envVars,
	}
}

func applyTemplating(list []string, data map[string]interface{}) error {
	for i, entry := range list {
		tmpl, err := template.New("argsTmpl").Option("missingkey=error").Parse(entry)
		if err != nil {
			return err
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return err
		}

		list[i] = buf.String()
	}

	return nil
}

func createBuildArgs(buildCfg Config) ([]string, error) {
	var args []string

	data := createTemplateData()

	if len(buildCfg.Flags) > 0 {
		if err := applyTemplating(buildCfg.Flags, data); err != nil {
			return nil, err
		}

		args = append(args, buildCfg.Flags...)
	}

	if len(buildCfg.Ldflags) > 0 {
		if err := applyTemplating(buildCfg.Ldflags, data); err != nil {
			return nil, err
		}

		args = append(args, fmt.Sprintf("-ldflags=%s", strings.Join(buildCfg.Ldflags, " ")))
	}

	return args, nil
}

func (g *gobuild) configForImportPath(ip string) Config {
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

func (g *gobuild) buildOne(ctx context.Context, refStr string, base v1.Image, platform *v1.Platform) (oci.SignedImage, error) {
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

	// Do the build into a temporary file.
	file, err := g.build(ctx, ref.Path(), g.dir, *platform, g.configForImportPath(ref.Path()))
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(filepath.Dir(file))

	var layers []mutate.Addendum

	// Create a layer from the kodata directory under this import path.
	dataLayerBuf, err := g.tarKoData(ref, platform)
	if err != nil {
		return nil, err
	}
	dataLayerBytes := dataLayerBuf.Bytes()
	dataLayer, err := tarball.LayerFromOpener(func() (io.ReadCloser, error) {
		return ioutil.NopCloser(bytes.NewBuffer(dataLayerBytes)), nil
	}, tarball.WithCompressedCaching)
	if err != nil {
		return nil, err
	}
	layers = append(layers, mutate.Addendum{
		Layer: dataLayer,
		History: v1.History{
			Author:    "ko",
			CreatedBy: "ko build " + ref.String(),
			Comment:   "kodata contents, at $KO_DATA_PATH",
		},
	})

	appDir := "/ko-app"
	appPath := path.Join(appDir, appFilename(ref.Path()))

	// Construct a tarball with the binary and produce a layer.
	binaryLayerBuf, err := tarBinary(appPath, file, v1.Time{}, platform)
	if err != nil {
		return nil, err
	}
	binaryLayerBytes := binaryLayerBuf.Bytes()
	binaryLayer, err := tarball.LayerFromOpener(func() (io.ReadCloser, error) {
		return ioutil.NopCloser(bytes.NewBuffer(binaryLayerBytes)), nil
	}, tarball.WithCompressedCaching, tarball.WithEstargzOptions(estargz.WithPrioritizedFiles([]string{
		// When using estargz, prioritize downloading the binary entrypoint.
		appPath,
	})))
	if err != nil {
		return nil, err
	}
	layers = append(layers, mutate.Addendum{
		Layer: binaryLayer,
		History: v1.History{
			Author:    "ko",
			CreatedBy: "ko build " + ref.String(),
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
	if platform.OS == "windows" {
		cfg.Config.Entrypoint = []string{`C:\ko-app\` + appFilename(ref.Path())}
		updatePath(cfg, `C:\ko-app`)
		cfg.Config.Env = append(cfg.Config.Env, `KO_DATA_PATH=C:\var\run\ko`)
	} else {
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

	image, err := mutate.ConfigFile(withApp, cfg)
	if err != nil {
		return nil, err
	}

	empty := v1.Time{}
	if g.creationTime != empty {
		image, err = mutate.CreatedAt(image, g.creationTime)
		if err != nil {
			return nil, err
		}
	}

	si := signed.Image(image)

	if g.sbom != nil {
		sbom, mt, err := g.sbom(ctx, file, appPath, image)
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
	}
	return si, nil
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
	baseRef, base, err := g.getBase(ctx, s)
	if err != nil {
		return nil, err
	}

	// Determine what kind of base we have and if we should publish an image or an index.
	mt, err := base.MediaType()
	if err != nil {
		return nil, err
	}

	// Take the digest of the base index or image, to annotate images we'll build later.
	baseDigest, err := base.Digest()
	if err != nil {
		return nil, err
	}

	// Annotate the base image we pass to the build function with
	// annotations indicating the digest (and possibly tag) of the
	// base image.  This will be inherited by the image produced.
	if mt != types.DockerManifestList {
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
		res, err = g.buildOne(ctx, s, baseImage, nil)
	default:
		return nil, fmt.Errorf("base image media type: %s", mt)
	}
	if err != nil {
		return nil, err
	}
	return res, nil
}

// TODO(#192): Do these in parallel?
func (g *gobuild) buildAll(ctx context.Context, ref string, baseIndex v1.ImageIndex) (oci.SignedImageIndex, error) {
	im, err := baseIndex.IndexManifest()
	if err != nil {
		return nil, err
	}

	// Build an image for each child from the base and append it to a new index to produce the result.
	adds := []ocimutate.IndexAddendum{}
	for _, desc := range im.Manifests {
		// Nested index is pretty rare. We could support this in theory, but return an error for now.
		if desc.MediaType != types.OCIManifestSchema1 && desc.MediaType != types.DockerManifestSchema2 {
			return nil, fmt.Errorf("%q has unexpected mediaType %q in base for %q", desc.Digest, desc.MediaType, ref)
		}

		if !g.platformMatcher.matches(desc.Platform) {
			continue
		}

		baseImage, err := baseIndex.Image(desc.Digest)
		if err != nil {
			return nil, err
		}
		img, err := g.buildOne(ctx, ref, baseImage, desc.Platform)
		if err != nil {
			return nil, err
		}
		adds = append(adds, ocimutate.IndexAddendum{
			Add: img,
			Descriptor: v1.Descriptor{
				URLs:        desc.URLs,
				MediaType:   desc.MediaType,
				Annotations: desc.Annotations,
				Platform:    desc.Platform,
			},
		})
	}

	baseType, err := baseIndex.MediaType()
	if err != nil {
		return nil, err
	}
	idx := ocimutate.AppendManifests(mutate.IndexMediaType(empty.Index, baseType), adds...)

	// TODO(mattmoor): If we want to attach anything (e.g. signatures, attestations, SBOM)
	// at the index level, we would do it here!
	return idx, nil
}

func parseSpec(spec string) (*platformMatcher, error) {
	// Don't bother parsing "all".
	// "" should never happen because we default to linux/amd64.
	platforms := []v1.Platform{}
	if spec == "all" || spec == "" {
		return &platformMatcher{spec: spec}, nil
	}

	for _, platform := range strings.Split(spec, ",") {
		var p v1.Platform
		parts := strings.Split(strings.TrimSpace(platform), "/")
		if len(parts) > 0 {
			p.OS = parts[0]
		}
		if len(parts) > 1 {
			p.Architecture = parts[1]
		}
		if len(parts) > 2 {
			p.Variant = parts[2]
		}
		if len(parts) > 3 {
			return nil, fmt.Errorf("too many slashes in platform spec: %s", platform)
		}
		platforms = append(platforms, p)
	}
	return &platformMatcher{spec: spec, platforms: platforms}, nil
}

func (pm *platformMatcher) matches(base *v1.Platform) bool {
	if pm.spec == "all" {
		return true
	}

	// Don't build anything without a platform field unless "all". Unclear what we should do here.
	if base == nil {
		return false
	}

	for _, p := range pm.platforms {
		if p.OS != "" && base.OS != p.OS {
			continue
		}
		if p.Architecture != "" && base.Architecture != p.Architecture {
			continue
		}
		if p.Variant != "" && base.Variant != p.Variant {
			continue
		}

		return true
	}

	return false
}
