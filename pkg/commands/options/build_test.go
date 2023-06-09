// Copyright 2021 ko Build Authors All Rights Reserved.
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

package options

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/google/ko/pkg/build"
	"github.com/spf13/cobra"
)

func TestDefaultBaseImage(t *testing.T) {
	bo := &BuildOptions{
		WorkingDirectory: "testdata/config",
	}
	err := bo.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}

	wantDefaultBaseImage := "alpine" // matches value in ./testdata/config/.ko.yaml
	if bo.BaseImage != wantDefaultBaseImage {
		t.Fatalf("wanted BaseImage %s, got %s", wantDefaultBaseImage, bo.BaseImage)
	}
}

func TestDefaultPlatformsAll(t *testing.T) {
	allBo := &BuildOptions{
		WorkingDirectory: "testdata/config",
	}
	err := allBo.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}

	wantDefaultPlatformsAll := []string{"all"} // matches value in ./testdata/config/.ko.yaml
	if !reflect.DeepEqual(allBo.DefaultPlatforms, wantDefaultPlatformsAll) {
		t.Fatalf("wanted DefaultPlatforms %s, got %s", wantDefaultPlatformsAll, allBo.DefaultPlatforms)
	}

	multipleBo := &BuildOptions{
		WorkingDirectory: "testdata/multiple-platforms",
	}
	err = multipleBo.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}

	wantDefaultPlatformsMultiple := []string{"linux/arm64", "linux/amd64"} // matches value in ./testdata/multiple-platforms/.ko.yaml
	if !reflect.DeepEqual(multipleBo.DefaultPlatforms, wantDefaultPlatformsMultiple) {
		t.Fatalf("wanted DefaultPlatforms %s, got %s", wantDefaultPlatformsMultiple, multipleBo.DefaultPlatforms)
	}
}

func TestBuildConfigWithWorkingDirectoryAndDirAndMain(t *testing.T) {
	bo := &BuildOptions{
		WorkingDirectory: "testdata/paths",
	}
	err := bo.LoadConfig()
	if err != nil {
		t.Fatalf("NewBuilder(): %v", err)
	}

	if len(bo.BuildConfigs) != 1 {
		t.Fatalf("expected 1 build config, got %d", len(bo.BuildConfigs))
	}
	expectedImportPath := "example.com/testapp/cmd/foo" // module from app/go.mod + `main` from .ko.yaml
	if _, exists := bo.BuildConfigs[expectedImportPath]; !exists {
		t.Fatalf("expected build config for import path [%s], got %+v", expectedImportPath, bo.BuildConfigs)
	}
}

func TestCreateBuildConfigs(t *testing.T) {
	compare := func(expected string, actual string) {
		if expected != actual {
			t.Errorf("test case failed: expected '%#v', but actual value is '%#v'", expected, actual)
		}
	}

	buildConfigs := []build.Config{
		{ID: "defaults"},
		{ID: "OnlyMain", Main: "test"},
		{ID: "OnlyMainWithFile", Main: "test/main.go"},
		{ID: "OnlyDir", Dir: "test"},
		{ID: "DirAndMain", Dir: "test", Main: "main.go"},
	}

	for _, b := range buildConfigs {
		buildConfigMap, err := createBuildConfigMap("../../..", []build.Config{b})
		if err != nil {
			t.Fatal(err)
		}
		for importPath, buildCfg := range buildConfigMap {
			switch buildCfg.ID {
			case "defaults":
				compare("github.com/google/ko", importPath)

			case "OnlyMain", "OnlyMainWithFile", "OnlyDir", "DirAndMain":
				compare("github.com/google/ko/test", importPath)

			default:
				t.Fatalf("unknown test case: %s", buildCfg.ID)
			}
		}
	}
}

func TestAddBuildOptionsSetsDefaultsForNonFlagOptions(t *testing.T) {
	cmd := &cobra.Command{}
	bo := &BuildOptions{}
	AddBuildOptions(cmd, bo)
	if !bo.Trimpath {
		t.Error("expected Trimpath=true")
	}
}

func TestOverrideConfigPath(t *testing.T) {
	const envName = "KO_CONFIG_PATH"
	bo := &BuildOptions{}
	for _, tc := range []struct {
		name         string
		koConfigPath string
		err          string
	}{{
		name:         ".ko.yaml does not exist",
		koConfigPath: "testdata",
		err:          "testdata/.ko.yaml: no such file or directory",
	}, {
		name:         "config path does not contain .ko.yaml",
		koConfigPath: "testdata/bad-config",
		err:          "testdata/bad-config/.ko.yaml is not a regular file",
	}, {
		name:         "config path is a directory containing a .ko.yaml",
		koConfigPath: "testdata/config",
	}, {
		name:         "config path points to a file",
		koConfigPath: "testdata/config/my-ko.yaml",
	}} {
		t.Run(tc.name, func(t *testing.T) {
			oldEnv := os.Getenv(envName)
			defer os.Setenv(envName, oldEnv)

			os.Setenv(envName, tc.koConfigPath)
			err := bo.LoadConfig()
			if err == nil {
				if tc.err != "" {
					t.Fatalf("expected error %q, saw nil", tc.err)
				}
			} else {
				if tc.err == "" {
					t.Errorf("expected no error, saw: %v", err)
				}
				if !strings.Contains(err.Error(), tc.err) {
					t.Errorf("expected error to contain %q, saw: %v", tc.err, err)
				}
			}
		})
	}
}
