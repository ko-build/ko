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
	"context"
	"fmt"
	gb "go/build"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/ko/pkg/build"
	"github.com/google/ko/pkg/publish"

	"golang.org/x/tools/go/packages"
)

func qualifyLocalImport(importpath string) (string, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName,
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

func publishImages(ctx context.Context, importpaths []string, pub publish.Interface, b build.Interface) (map[string]name.Reference, error) {
	imgs := make(map[string]name.Reference)
	for _, importpath := range importpaths {
		if gb.IsLocalImport(importpath) {
			var err error
			importpath, err = qualifyLocalImport(importpath)
			if err != nil {
				return nil, err
			}
		}
		if !strings.HasPrefix(importpath, build.StrictScheme) {
			importpath = build.StrictScheme + importpath
		}

		if err := b.IsSupportedReference(importpath); err != nil {
			return nil, fmt.Errorf("importpath %q is not supported: %v", importpath, err)
		}

		img, err := b.Build(ctx, importpath)
		if err != nil {
			return nil, fmt.Errorf("error building %q: %v", importpath, err)
		}
		ref, err := pub.Publish(ctx, img, importpath)
		if err != nil {
			return nil, fmt.Errorf("error publishing %s: %v", importpath, err)
		}
		imgs[importpath] = ref
	}
	return imgs, nil
}
