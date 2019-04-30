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

package commands

import (
	"fmt"
	gb "go/build"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/ko/pkg/build"
	"github.com/google/ko/pkg/publish"
)

func qualifyLocalImport(importpath, gopathsrc, pwd string) (string, error) {
	if !strings.HasPrefix(pwd, gopathsrc) {
		return "", fmt.Errorf("pwd (%q) must be on $GOPATH/src (%q) to support local imports", pwd, gopathsrc)
	}
	// Given $GOPATH/src and $PWD (which must be within $GOPATH/src), trim
	// off $GOPATH/src/ from $PWD and append local importpath to get the
	// fully-qualified importpath.
	return filepath.Join(strings.TrimPrefix(pwd, gopathsrc+string(filepath.Separator)), importpath), nil
}

func publishImages(importpaths []string, pub publish.Interface, b build.Interface) (map[string]name.Reference, error) {
	imgs := make(map[string]name.Reference)
	for _, importpath := range importpaths {
		if gb.IsLocalImport(importpath) {
			// Qualify relative imports to their fully-qualified
			// import path, assuming $PWD is within $GOPATH/src.
			gopathsrc := filepath.Join(gb.Default.GOPATH, "src")
			pwd, err := os.Getwd()
			if err != nil {
				return nil, fmt.Errorf("error getting current working directory: %v", err)
			}
			importpath, err = qualifyLocalImport(importpath, gopathsrc, pwd)
			if err != nil {
				return nil, err
			}
		}

		if !b.IsSupportedReference(importpath) {
			return nil, fmt.Errorf("importpath %q is not supported", importpath)
		}

		img, err := b.Build(importpath)
		if err != nil {
			return nil, fmt.Errorf("error building %q: %v", importpath, err)
		}
		ref, err := pub.Publish(img, importpath)
		if err != nil {
			return nil, fmt.Errorf("error publishing %s: %v", importpath, err)
		}
		imgs[importpath] = ref
	}
	return imgs, nil
}
