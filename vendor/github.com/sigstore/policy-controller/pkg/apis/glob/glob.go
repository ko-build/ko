//
// Copyright 2022 The Sigstore Authors.
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

package glob

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
)

const (
	ResolvedDockerhubHost = "index.docker.io/"
	// Images such as "busybox" reside in the dockerhub "library" repository
	// The full resolved image reference would be index.docker.io/library/busybox
	DockerhubPublicRepository = "library/"
)

var validGlob = regexp.MustCompile(`^[a-zA-Z0-9-_:\/\*\.@]+$`)

// Compile attempts to normalize the glob and turn it into a regular expression
// that we can use for matching image names.
func Compile(glob string) (*regexp.Regexp, error) {
	if glob == "*/*" {
		// TODO: Warn that the glob match "*/*" should be "index.docker.io/*/*".
		glob = "index.docker.io/*/*"
	}
	if glob == "*" {
		// TODO: Warn that the glob match "*" should be "index.docker.io/library/*".
		glob = "index.docker.io/library/*"
	}

	// Reject that glob doesn't look like a regexp
	if !validGlob.MatchString(glob) {
		return nil, fmt.Errorf("invalid glob %q", glob)
	}

	// Translate glob to regexp.
	glob = strings.ReplaceAll(glob, ".", `\.`) // . in glob means \. in regexp
	glob = strings.ReplaceAll(glob, "**", "#") // ** in glob means 0+ of any character in regexp
	// We replace ** with a placeholder here, rather than `.*` directly, because the next line
	// would replace that `*` again, breaking the regexp. So we stash the change with a placeholder,
	// then replace the placeholder later to preserve the original intent.
	glob = strings.ReplaceAll(glob, "*", "[^/]*") // * in glob means 0+ of any non-`/` character in regexp
	glob = strings.ReplaceAll(glob, "#", ".*")
	glob = fmt.Sprintf("^%s$", glob) // glob must match the whole string

	return regexp.Compile(glob)
}

// Match will return true if the image reference matches the requested glob pattern.
//
// If the image reference is invalid, an error will be returned.
//
// In the glob pattern, the "*" character matches any non-"/" character, and "**" matches any character, including "/".
//
// If the image is a DockerHub official image like "ubuntu" or "debian", the glob that matches it must be something like index.docker.io/library/ubuntu.
// If the image is a DockerHub used-owned image like "myuser/myapp", then the glob that matches it must be something like index.docker.io/myuser/myapp.
// This means that the glob patterns "*" will not match the image name "ubuntu", and "*/*" will not match "myuser/myapp"; the "index.docker.io" prefix is required.
//
// If the image does not specify a tag (e.g., :latest or :v1.2.3), the tag ":latest" will be assumed.
//
// Note that the tag delimiter (":") does not act as a breaking separator for the purposes of a "*" glob.
// To match any tag, the glob should end with ":**".
func Match(glob, image string) (bool, error) {
	re, err := Compile(glob)
	if err != nil {
		return false, err
	}

	// TODO: do we want ":" to count as a separator like "/" is?

	ref, err := name.ParseReference(image, name.WeakValidation)
	if err != nil {
		return false, err
	}

	match := re.MatchString(ref.Name())
	if !match && ref.Name() != image {
		// If the image was not fully qualified, try matching the glob against the original non-fully-qualified.
		// This should be a warning and this behavior should eventually be removed.
		match = re.MatchString(image)
	}
	return match, nil
}
