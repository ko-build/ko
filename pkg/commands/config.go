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
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/google/ko/pkg/build"
	"github.com/spf13/viper"
)

var (
	defaultBaseImage   name.Reference
	baseImageOverrides map[string]name.Reference
)

func getBaseImage(platform string) build.GetBase {
	return func(ctx context.Context, s string) (build.Result, error) {
		s = strings.TrimPrefix(s, build.StrictScheme)
		// Viper configuration file keys are case insensitive, and are
		// returned as all lowercase.  This means that import paths with
		// uppercase must be normalized for matching here, e.g.
		//    github.com/GoogleCloudPlatform/foo/cmd/bar
		// comes through as:
		//    github.com/googlecloudplatform/foo/cmd/bar
		ref, ok := baseImageOverrides[strings.ToLower(s)]
		if !ok {
			ref = defaultBaseImage
		}
		ropt := []remote.Option{
			remote.WithAuthFromKeychain(authn.DefaultKeychain),
			remote.WithUserAgent(ua()),
			remote.WithContext(ctx),
		}

		// Using --platform=all will use an image index for the base,
		// otherwise we'll resolve it to the appropriate platform.
		//
		// Platforms can be comma-separated if we only want a subset of the base
		// image.
		multiplatform := platform == "all" || strings.Contains(platform, ",")
		var p v1.Platform
		if platform != "" && !multiplatform {
			parts := strings.Split(platform, "/")
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
			ropt = append(ropt, remote.WithPlatform(p))
		}

		log.Printf("Using base %s for %s", ref, s)
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
}

func getCreationTime() (*v1.Time, error) {
	epoch := os.Getenv("SOURCE_DATE_EPOCH")
	if epoch == "" {
		return nil, nil
	}

	seconds, err := strconv.ParseInt(epoch, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("the environment variable SOURCE_DATE_EPOCH should be the number of seconds since January 1st 1970, 00:00 UTC, got: %v", err)
	}
	return &v1.Time{Time: time.Unix(seconds, 0)}, nil
}

func createCancellableContext() context.Context {
	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-signals
		cancel()
	}()

	return ctx
}

func init() {
	// If omitted, use this base image.
	viper.SetDefault("defaultBaseImage", "gcr.io/distroless/static:nonroot")
	viper.SetConfigName(".ko") // .yaml is implicit
	viper.SetEnvPrefix("KO")
	viper.AutomaticEnv()

	if override := os.Getenv("KO_CONFIG_PATH"); override != "" {
		viper.AddConfigPath(override)
	}

	viper.AddConfigPath("./")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Fatalf("error reading config file: %v", err)
		}
	}

	ref := viper.GetString("defaultBaseImage")
	dbi, err := name.ParseReference(ref)
	if err != nil {
		log.Fatalf("'defaultBaseImage': error parsing %q as image reference: %v", ref, err)
	}
	defaultBaseImage = dbi

	baseImageOverrides = make(map[string]name.Reference)
	overrides := viper.GetStringMapString("baseImageOverrides")
	for k, v := range overrides {
		bi, err := name.ParseReference(v)
		if err != nil {
			log.Fatalf("'baseImageOverrides': error parsing %q as image reference: %v", v, err)
		}
		baseImageOverrides[k] = bi
	}
}
