// Copyright 2022 Chainguard, Inc.
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

package options

import (
	"fmt"
	"os"
	"time"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"

	"chainguard.dev/apko/pkg/build/types"
)

type Options struct {
	UseDockerMediaTypes     bool
	WantSBOM                bool
	UseProot                bool
	WithVCS                 bool
	WorkDir                 string
	TarballPath             string
	Tags                    []string
	SourceDateEpoch         time.Time
	SBOMPath                string
	SBOMFormats             []string
	ExtraKeyFiles           []string
	ExtraRepos              []string
	Arch                    types.Architecture
	Log                     *logrus.Logger
	TempDirPath             string
	PackageVersionTag       string
	PackageVersionTagStem   bool
	PackageVersionTagPrefix string
	TagSuffix               string
	Local                   bool
}

var Default = Options{
	Log: &logrus.Logger{
		Out: os.Stderr,
		Formatter: &nested.Formatter{
			ShowFullLevel: true,
		},
		Hooks: make(logrus.LevelHooks),
		Level: logrus.InfoLevel,
	},
}

func (o *Options) Summarize(logger *logrus.Entry) {
	logger.Printf("  working directory: %s", o.WorkDir)
	logger.Printf("  tarball path: %s", o.TarballPath)
	logger.Printf("  use proot: %t", o.UseProot)
	logger.Printf("  source date: %s", o.SourceDateEpoch)
	logger.Printf("  Docker mediatypes: %t", o.UseDockerMediaTypes)
	logger.Printf("  SBOM output path: %s", o.SBOMPath)
	logger.Printf("  arch: %v", o.Arch.ToAPK())
}

func (o *Options) Logger() *logrus.Entry {
	fields := logrus.Fields{}
	emptyArch := types.Architecture{}

	if o.Arch != emptyArch {
		fields["arch"] = o.Arch.ToAPK()
	}

	return o.Log.WithFields(fields)
}

// Tempdir returns the temporary directory where apko will create
// the layer blobs
func (o *Options) TempDir() string {
	if o.TempDirPath != "" {
		return o.TempDirPath
	}

	path, err := os.MkdirTemp(os.TempDir(), "apko-temp-*")
	if err != nil {
		o.Logger().Fatalf(fmt.Errorf("creating tempdir: %w", err).Error())
	}
	o.TempDirPath = path
	return o.TempDirPath
}

// TarballFileName returns a deterministic filename for the layer taball
func (o Options) TarballFileName() string {
	tarName := "apko.tar.gz"
	if o.Arch.String() != "" {
		tarName = fmt.Sprintf("apko-%s.tar.gz", o.Arch.ToAPK())
	}
	return tarName
}
