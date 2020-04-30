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
	"errors"
	"fmt"
	"os"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/ko/pkg/commands/options"
	"github.com/google/ko/pkg/parameters"
	"github.com/google/ko/pkg/publish"
)

func MakePublisher(po *parameters.PublishParameters) (publish.Interface, error) {
	// Create the publish.Interface that we will use to publish image references
	// to either a docker daemon or a container image registry.
	innerPublisher, err := func() (publish.Interface, error) {
		repoName := os.Getenv("KO_DOCKER_REPO")
		namer := options.MakeNamer(po)
		if repoName == publish.LocalDomain || po.Local {
			// TODO(jonjohnsonjr): I'm assuming that nobody will
			// use local with other publishers, but that might
			// not be true.
			return publish.NewDaemon(namer, po.Tags), nil
		}

		if repoName == "" {
			return nil, errors.New("KO_DOCKER_REPO environment variable is unset")
		}
		if _, err := name.NewRegistry(repoName); err != nil {
			if _, err := name.NewRepository(repoName); err != nil {
				return nil, fmt.Errorf("failed to parse environment variable KO_DOCKER_REPO=%q as repository: %v", repoName, err)
			}
		}

		publishers := []publish.Interface{}
		if po.OCILayoutPath != "" {
			lp, err := publish.NewLayout(po.OCILayoutPath)
			if err != nil {
				return nil, fmt.Errorf("failed to create LayoutPublisher for %q: %v", po.OCILayoutPath, err)
			}
			publishers = append(publishers, lp)
		}
		if po.TarballFile != "" {
			tp := publish.NewTarball(po.TarballFile, repoName, namer, po.Tags)
			publishers = append(publishers, tp)
		}
		if po.Push {
			dp, err := publish.NewDefault(repoName,
				publish.WithTransport(defaultTransport()),
				publish.WithAuthFromKeychain(authn.DefaultKeychain),
				publish.WithNamer(namer),
				publish.WithTags(po.Tags),
				publish.Insecure(po.InsecureRegistry))
			if err != nil {
				return nil, err
			}
			publishers = append(publishers, dp)
		}
		return publish.MultiPublisher(publishers...), nil
	}()
	if err != nil {
		return nil, err
	}

	// Wrap publisher in a memoizing publisher implementation.
	return publish.NewCaching(innerPublisher)
}
