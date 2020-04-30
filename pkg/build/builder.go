// Copyright 2020 Google LLC All Rights Reserved.
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

package build

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/ko/pkg/parameters"
)

var (
	DefaultBaseImage   name.Reference
	BaseImageOverrides map[string]name.Reference
)

type useragentTransport struct {
	useragent string
	inner     http.RoundTripper
}

// RoundTrip implements http.RoundTripper
func (ut *useragentTransport) RoundTrip(in *http.Request) (*http.Response, error) {
	in.Header.Set("User-Agent", ut.useragent)
	return ut.inner.RoundTrip(in)
}

func ua() string {
	if v := GetVersion(); v != "" {
		return "ko/" + v
	}
	return "ko"
}

func defaultTransport() http.RoundTripper {
	return &useragentTransport{
		inner:     http.DefaultTransport,
		useragent: ua(),
	}
}

func getBaseImage(s string) (v1.Image, error) {
	// Viper configuration file keys are case insensitive, and are
	// returned as all lowercase.  This means that import paths with
	// uppercase must be normalized for matching here, e.g.
	//    github.com/GoogleCloudPlatform/foo/cmd/bar
	// comes through as:
	//    github.com/googlecloudplatform/foo/cmd/bar
	ref, ok := BaseImageOverrides[strings.ToLower(s)]
	if !ok {
		ref = DefaultBaseImage
	}
	log.Printf("Using base %s for %s", ref, s)
	return remote.Image(ref,
		remote.WithTransport(defaultTransport()),
		remote.WithAuthFromKeychain(authn.DefaultKeychain))
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
	return &v1.Time{time.Unix(seconds, 0)}, nil
}

func gobuildOptions(bo *parameters.BuildParameters) ([]Option, error) {
	creationTime, err := getCreationTime()
	if err != nil {
		return nil, err
	}
	opts := []Option{
		WithBaseImages(getBaseImage),
	}
	if creationTime != nil {
		opts = append(opts, WithCreationTime(*creationTime))
	}
	if bo.DisableOptimizations {
		opts = append(opts, WithDisabledOptimizations())
	}
	return opts, nil
}

func MakeBuilder(bo *parameters.BuildParameters) (*Caching, error) {
	opt, err := gobuildOptions(bo)
	if err != nil {
		return nil, fmt.Errorf("error setting up builder options: %v", err)
	}
	innerBuilder, err := NewGo(opt...)
	if err != nil {
		return nil, err
	}

	innerBuilder = NewLimiter(innerBuilder, bo.ConcurrentBuilds)

	// tl;dr Wrap builder in a caching builder.
	//
	// The caching builder should on Build calls:
	//  - Check for a valid Build future
	//    - if a valid Build future exists at the time of the request,
	//      then block on it.
	//    - if it does not, then initiate and record a Build future.
	//  - When import paths are "affected" by filesystem changes during a
	//    Watch, then invalidate their build futures *before* we put the
	//    affected yaml files onto the channel
	//
	// This will benefit the following key cases:
	// 1. When the same import path is referenced across multiple yaml files
	//    we can elide subsequent builds by blocking on the same image future.
	// 2. When an affected yaml file has multiple import paths (mostly unaffected)
	//    we can elide the builds of unchanged import paths.
	return NewCaching(innerBuilder)
}
