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
	specsv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"golang.org/x/tools/go/packages"
)

const (
	defaultAppFilename = "ko-app"
)

// GetBase takes an importpath and returns a base image reference and base image (or index).
type GetBase func(context.Context, string) (name.Reference, Result, error)

type builder func(context.Context, string, string, v1.Platform, Config) (string, error)

type buildContext interface {
	Import(path string, srcDir string, mode gb.ImportMode) (*gb.Package, error)
}

type platformMatcher struct {
	spec      string
	platforms []v1.Platform
}

type gobuild struct {
	getBase              GetBase
	creationTime         v1.Time
	kodataCreationTime   v1.Time
	build                builder
	disableOptimizations bool
	buildConfigs         map[string]Config
	mod                  *modules
	buildContext         buildContext
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
	disableOptimizations bool
	buildConfigs         map[string]Config
	mod                  *modules
	buildContext         buildContext
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
		disableOptimizations: gbo.disableOptimizations,
		buildConfigs:         gbo.buildConfigs,
		mod:                  gbo.mod,
		buildContext:         gbo.buildContext,
		labels:               gbo.labels,
		dir:                  gbo.dir,
		platformMatcher:      matcher,
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
func moduleInfo(ctx context.Context, dir string) (*modules, error) {
	modules := modules{
		deps: make(map[string]*modInfo),
	}

	// TODO we read all the output as a single byte array - it may
	// be possible & more efficient to stream it
	cmd := exec.CommandContext(ctx, "go", "list", "-mod=readonly", "-json", "-m", "all")
	cmd.Dir = dir
	output, err := cmd.Output()
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

// getGoroot shells out to `go env GOROOT` to determine
// the GOROOT for the installed version of go so that we
// can set it in our buildContext. By default, the GOROOT
// of our buildContext is set to the GOROOT at install
// time for `ko`, which means that we break when certain
// package managers update go or when using a pre-built
// `ko` binary that expects a different GOROOT.
//
// See https://github.com/google/ko/issues/106
func getGoroot(ctx context.Context, dir string) (string, error) {
	cmd := exec.CommandContext(ctx, "go", "env", "GOROOT")
	// It's probably not necessary to set the command working directory here,
	// but it helps keep everything consistent.
	cmd.Dir = dir
	output, err := cmd.Output()
	return strings.TrimSpace(string(output)), err
}

// NewGo returns a build.Interface implementation that:
//  1. builds go binaries named by importpath,
//  2. containerizes the binary on a suitable base.
//
// The `dir` argument is the working directory for executing the `go` tool.
// If `dir` is empty, the function uses the current process working directory.
func NewGo(ctx context.Context, dir string, options ...Option) (Interface, error) {
	// TODO: We could do moduleInfo() and getGoroot() concurrently.
	module, err := moduleInfo(ctx, dir)
	if err != nil {
		return nil, err
	}

	goroot, err := getGoroot(ctx, dir)
	if err != nil {
		// On error, print the output and set goroot to "" to avoid using it later.
		log.Printf("Unexpected error running \"go env GOROOT\": %v\n%v", err, goroot)
		goroot = ""
	} else if goroot == "" {
		log.Printf(`Unexpected: $(go env GOROOT) == ""`)
	}

	// If $(go env GOROOT) successfully returns a non-empty string that differs from
	// the default build context GOROOT, use $(go env GOROOT) instead.
	bc := gb.Default
	bc.Dir = dir
	if goroot != "" && bc.GOROOT != goroot {
		bc.GOROOT = goroot
	}

	gbo := &gobuildOpener{
		build:        build,
		mod:          module,
		buildContext: &bc,
		// dir is set on both buildContext and on gbo. Not ideal, but the
		// build.Context interface doesn't expose
		dir: dir,
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
			return "", fmt.Errorf("qualifying local import %s: %v", importpath, err)
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

func getGoarm(platform v1.Platform) (string, error) {
	if !strings.HasPrefix(platform.Variant, "v") {
		return "", fmt.Errorf("strange arm variant: %v", platform.Variant)
	}

	vs := strings.TrimPrefix(platform.Variant, "v")
	variant, err := strconv.Atoi(vs)
	if err != nil {
		return "", fmt.Errorf("cannot parse arm variant %q: %v", platform.Variant, err)
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
		return "", fmt.Errorf("could not create env for %s: %v", ip, err)
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
			return nil, fmt.Errorf("goarm failure: %v", err)
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
			return nil, fmt.Errorf("writing dir %q: %v", dir, err)
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
			return nil, fmt.Errorf("writing dir %q: %v", dir, err)
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
	config, ok := g.buildConfigs[ip]
	if !ok {
		// Apply default build flags in case none were supplied
		config.Flags = append(config.Flags, "-trimpath")
	}

	if g.disableOptimizations {
		// Disable optimizations (-N) and inlining (-l).
		config.Flags = append(config.Flags, "-gcflags", "all=-N -l")
	}

	return config
}

func (g *gobuild) buildOne(ctx context.Context, refStr string, base v1.Image, platform *v1.Platform) (v1.Image, error) {
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
			CreatedBy: "ko publish " + ref.String(),
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
		return mutate.CreatedAt(image, g.creationTime)
	}
	return image, nil
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

	// Annotate the image or index with base image information.
	// (Docker manifest lists don't support annotations)
	if mt != types.DockerManifestList {
		anns := map[string]string{
			specsv1.AnnotationBaseImageDigest: baseDigest.String(),
		}
		if _, ok := baseRef.(name.Tag); ok {
			anns[specsv1.AnnotationBaseImageName] = baseRef.Name()
		}
		res = mutate.Annotations(res, anns).(Result)
	}

	return res, nil
}

// TODO(#192): Do these in parallel?
func (g *gobuild) buildAll(ctx context.Context, ref string, baseIndex v1.ImageIndex) (v1.ImageIndex, error) {
	im, err := baseIndex.IndexManifest()
	if err != nil {
		return nil, err
	}

	// Build an image for each child from the base and append it to a new index to produce the result.
	adds := []mutate.IndexAddendum{}
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

	baseType, err := baseIndex.MediaType()
	if err != nil {
		return nil, err
	}
	idx := mutate.IndexMediaType(mutate.AppendManifests(empty.Index, adds...), baseType)

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
