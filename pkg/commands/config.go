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

package commands

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/types"

	"github.com/google/ko/pkg/build"
	"github.com/google/ko/pkg/commands/options"
	"github.com/google/ko/pkg/publish"
)

// getBaseImage returns a function that determines the base image for a given import path.
func getBaseImage(bo *options.BuildOptions) build.GetBase {
	cache := map[string]build.Result{}
	fetch := func(ctx context.Context, ref name.Reference) (build.Result, error) {
		// For ko.local, look in the daemon.
		if ref.Context().RegistryStr() == publish.LocalDomain {
			return daemon.Image(ref)
		}

		userAgent := ua()
		if bo.UserAgent != "" {
			userAgent = bo.UserAgent
		}
		ropt := []remote.Option{
			remote.WithAuthFromKeychain(authn.DefaultKeychain),
			remote.WithUserAgent(userAgent),
			remote.WithContext(ctx),
		}

		// Using --platform=all will use an image index for the base,
		// otherwise we'll resolve it to the appropriate platform.
		//
		// Platforms can be comma-separated if we only want a subset of the base
		// image.
		multiplatform := bo.Platform == "all" || strings.Contains(bo.Platform, ",")
		if bo.Platform != "" && !multiplatform {
			var p v1.Platform

			parts := strings.Split(bo.Platform, ":")
			if len(parts) == 2 {
				p.OSVersion = parts[1]
			}

			parts = strings.Split(parts[0], "/")
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
				return nil, fmt.Errorf("too many slashes in platform spec: %s", bo.Platform)
			}
			ropt = append(ropt, remote.WithPlatform(p))
		}

		desc, err := remote.Get(ref, ropt...)
		if err != nil {
			return nil, err
		}
		switch desc.MediaType {
		case types.OCIImageIndex, types.DockerManifestList:
			if multiplatform {
				return desc.ImageIndex()
			}
			return desc.Image()
		default:
			return desc.Image()
		}
	}
	return func(ctx context.Context, s string) (name.Reference, build.Result, error) {
		s = strings.TrimPrefix(s, build.StrictScheme)
		// Viper configuration file keys are case insensitive, and are
		// returned as all lowercase.  This means that import paths with
		// uppercase must be normalized for matching here, e.g.
		//    github.com/GoogleCloudPlatform/foo/cmd/bar
		// comes through as:
		//    github.com/googlecloudplatform/foo/cmd/bar
		baseImage, ok := bo.BaseImageOverrides[strings.ToLower(s)]
		if !ok || baseImage == "" {
			baseImage = bo.BaseImage
		}
		var nameOpts []name.Option
		if bo.InsecureRegistry {
			nameOpts = append(nameOpts, name.Insecure)
		}
		ref, err := name.ParseReference(baseImage, nameOpts...)
		if err != nil {
			return nil, nil, fmt.Errorf("parsing base image (%q): %w", baseImage, err)
		}

		if cached, ok := cache[ref.String()]; ok {
			return ref, cached, nil
		}

		log.Printf("Using base %s for %s", ref, s)
		result, err := fetch(ctx, ref)
		if err != nil {
			return ref, result, err
		}

		cache[ref.String()] = result
		return ref, result, nil
	}
}

func getTimeFromEnv(env string) (*v1.Time, error) {
	epoch := os.Getenv(env)
	if epoch == "" {
		return nil, nil
	}

	seconds, err := strconv.ParseInt(epoch, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("the environment variable %s should be the number of seconds since January 1st 1970, 00:00 UTC, got: %w", env, err)
	}
	return &v1.Time{Time: time.Unix(seconds, 0)}, nil
}

func getCreationTime() (*v1.Time, error) {
	return getTimeFromEnv("SOURCE_DATE_EPOCH")
}

func getKoDataCreationTime() (*v1.Time, error) {
	return getTimeFromEnv("KO_DATA_DATE_EPOCH")
}
