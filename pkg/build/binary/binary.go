package binary

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/ko/pkg/build/config"
	buildplatform "github.com/google/ko/pkg/build/platform"
)


func Build(ctx context.Context, ip string, dir string, platform v1.Platform, config config.Config) (string, error) {
	env, err := buildEnv(platform, os.Environ(), config.Env)
	if err != nil {
		return "", fmt.Errorf("could not create env for %s: %w", ip, err)
	}

	tmpDir, err := ioutil.TempDir("", "ko")
	if err != nil {
		return "", err
	}

	if dir := os.Getenv("KOCACHE"); dir != "" {
		dirInfo, err := os.Stat(dir)
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dir, os.ModePerm); err != nil && !os.IsExist(err) {
				return "", fmt.Errorf("could not create KOCACHE dir %s: %w", dir, err)
			}
		} else if !dirInfo.IsDir() {
			return "", fmt.Errorf("KOCACHE should be a directory, %s is not a directory", dir)
		}

		// TODO(#264): if KOCACHE is unset, default to filepath.Join(os.TempDir(), "ko").
		tmpDir = filepath.Join(dir, "bin", ip, platform.String())
		if err := os.MkdirAll(tmpDir, os.ModePerm); err != nil {
			return "", err
		}
	}

	file := filepath.Join(tmpDir, "out")

	goBinary := "go"
	if buildplatform.IsWasi(&platform) {
		goBinary = "tinygo"
	}
	cmd, err := gobuildCommand(ctx, config, goBinary, ip, file, dir, env)
	if err != nil {
		return "", fmt.Errorf("creating %v build exec.Cmd: %w", goBinary, err)
	}
	var output bytes.Buffer
	cmd.Stderr = &output
	cmd.Stdout = &output

	log.Printf("Building %s for %s with %s", ip, platform, goBinary)
	if err := cmd.Run(); err != nil {
		if os.Getenv("KOCACHE") == "" {
			os.RemoveAll(tmpDir)
		}
		log.Printf(`Unexpected error running "%v build": %v\n%v`, goBinary, err, output.String())
		return "", err
	}
	return file, nil
}

func gobuildCommand(
	ctx context.Context,
	config config.Config,
	goBinary string,
	ip string,
	outDir string,
	dir string,
	env []string,
) (*exec.Cmd, error) {
	buildArgs, err := createBuildArgs(config)
	if err != nil {
		return nil, err
	}

	args := make([]string, 0, 4+len(buildArgs))
	args = append(args, "build")
	args = append(args, buildArgs...)
	// if we are using tinygo, add wasi target
	if goBinary == "tinygo" {
		args = append(args, "-target", "wasi")
	}

	args = append(args, "-o", outDir)
	args = append(args, ip)
	cmd := exec.CommandContext(ctx, goBinary, args...)
	cmd.Dir = dir
	cmd.Env = env

	return cmd, nil
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
	if platform.Variant != "" {
		switch platform.Architecture {
		case "arm":
			// See: https://pkg.go.dev/cmd/go#hdr-Environment_variables
			goarm, err := getGoarm(platform)
			if err != nil {
				return nil, fmt.Errorf("goarm failure: %w", err)
			}
			if goarm != "" {
				env = append(env, "GOARM="+goarm)
			}
		case "amd64":
			// See: https://tip.golang.org/doc/go1.18#amd64
			env = append(env, "GOAMD64="+platform.Variant)
		}
	}

	env = append(env, userEnv...)
	env = append(env, configEnv...)
	return env, nil
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

func createBuildArgs(buildCfg config.Config) ([]string, error) {
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
