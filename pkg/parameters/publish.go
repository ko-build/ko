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

package parameters

// PublishParameters encapsulates parameters when publishing.
type PublishParameters struct {
	Tags []string

	// Push publishes images to a registry.
	Push bool

	// Local publishes images to a local docker daemon.
	Local            bool
	InsecureRegistry bool

	OCILayoutPath string
	TarballFile   string

	// PreserveImportPaths preserves the full import path after KO_DOCKER_REPO.
	PreserveImportPaths bool
	// BaseImportPaths uses the base path without MD5 hash after KO_DOCKER_REPO.
	BaseImportPaths bool
}
