/*
Copyright 2022 The Kubernetes Authors.

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

package version

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/common-nighthawk/go-figure"
)

const unknown = "unknown"

// Base version information.
//
// This is the fallback data used when version information from git is not
// provided via go ldflags.
var (
	// Output of "git describe". The prerequisite is that the
	// branch should be tagged using the correct versioning strategy.
	gitVersion = "devel"
	// SHA1 from git, output of $(git rev-parse HEAD)
	gitCommit = unknown
	// State of git tree, either "clean" or "dirty"
	gitTreeState = unknown
	// Build date in ISO8601 format, output of $(date -u +'%Y-%m-%dT%H:%M:%SZ')
	buildDate = unknown
	// flag to print the ascii name banner
	asciiName = "true"
	// goVersion is the used golang version.
	goVersion = unknown
	// compiler is the used golang compiler.
	compiler = unknown
	// platform is the used os/arch identifier.
	platform = unknown
)

type Info struct {
	GitVersion   string `json:"gitVersion"`
	GitCommit    string `json:"gitCommit"`
	GitTreeState string `json:"gitTreeState"`
	BuildDate    string `json:"buildDate"`
	GoVersion    string `json:"goVersion"`
	Compiler     string `json:"compiler"`
	Platform     string `json:"platform"`

	ASCIIName   string `json:"-"`
	FontName    string `json:"-"`
	Name        string `json:"-"`
	Description string `json:"-"`
}

func getBuildInfo() *debug.BuildInfo {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return nil
	}
	return bi
}

func getGitVersion(bi *debug.BuildInfo) string {
	if bi == nil {
		return unknown
	}

	// TODO: remove this when the issue https://github.com/golang/go/issues/29228 is fixed
	if bi.Main.Version == "(devel)" || bi.Main.Version == "" {
		return gitVersion
	}

	return bi.Main.Version
}

func getCommit(bi *debug.BuildInfo) string {
	return getKey(bi, "vcs.revision")
}

func getDirty(bi *debug.BuildInfo) string {
	modified := getKey(bi, "vcs.modified")
	if modified == "true" {
		return "dirty"
	}
	if modified == "false" {
		return "clean"
	}
	return unknown
}

func getBuildDate(bi *debug.BuildInfo) string {
	buildTime := getKey(bi, "vcs.time")
	t, err := time.Parse("2006-01-02T15:04:05Z", buildTime)
	if err != nil {
		return unknown
	}
	return t.Format("2006-01-02T15:04:05")
}

func getKey(bi *debug.BuildInfo, key string) string {
	if bi == nil {
		return unknown
	}
	for _, iter := range bi.Settings {
		if iter.Key == key {
			return iter.Value
		}
	}
	return unknown
}

// GetVersionInfo represents known information on how this binary was built.
func GetVersionInfo() Info {
	buildInfo := getBuildInfo()
	gitVersion = getGitVersion(buildInfo)
	if gitCommit == unknown {
		gitCommit = getCommit(buildInfo)
	}

	if gitTreeState == unknown {
		gitTreeState = getDirty(buildInfo)
	}

	if buildDate == unknown {
		buildDate = getBuildDate(buildInfo)
	}

	if goVersion == unknown {
		goVersion = runtime.Version()
	}

	if compiler == unknown {
		compiler = runtime.Compiler
	}

	if platform == unknown {
		platform = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	}

	return Info{
		ASCIIName:    asciiName,
		GitVersion:   gitVersion,
		GitCommit:    gitCommit,
		GitTreeState: gitTreeState,
		BuildDate:    buildDate,
		GoVersion:    goVersion,
		Compiler:     compiler,
		Platform:     platform,
	}
}

// String returns the string representation of the version info
func (i *Info) String() string {
	b := strings.Builder{}
	w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)

	// name and description are optional.
	if i.Name != "" {
		if i.ASCIIName == "true" {
			f := figure.NewFigure(strings.ToUpper(i.Name), i.FontName, true)
			_, _ = fmt.Fprint(w, f.String())
		}
		_, _ = fmt.Fprint(w, i.Name)
		if i.Description != "" {
			_, _ = fmt.Fprintf(w, ": %s", i.Description)
		}
		_, _ = fmt.Fprint(w, "\n\n")
	}

	_, _ = fmt.Fprintf(w, "GitVersion:\t%s\n", i.GitVersion)
	_, _ = fmt.Fprintf(w, "GitCommit:\t%s\n", i.GitCommit)
	_, _ = fmt.Fprintf(w, "GitTreeState:\t%s\n", i.GitTreeState)
	_, _ = fmt.Fprintf(w, "BuildDate:\t%s\n", i.BuildDate)
	_, _ = fmt.Fprintf(w, "GoVersion:\t%s\n", i.GoVersion)
	_, _ = fmt.Fprintf(w, "Compiler:\t%s\n", i.Compiler)
	_, _ = fmt.Fprintf(w, "Platform:\t%s\n", i.Platform)

	_ = w.Flush()
	return b.String()
}

// JSONString returns the JSON representation of the version info
func (i *Info) JSONString() (string, error) {
	b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (i *Info) CheckFontName(fontName string) bool {
	assetNames := figure.AssetNames()

	for _, font := range assetNames {
		if strings.Contains(font, fontName) {
			return true
		}
	}

	fmt.Fprintln(os.Stderr, "font not valid, using default")
	return false
}
