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

// Package version provides an importable cobra command and a fixed package
// path location to set compile time version information. To override the
// default values, set the `-ldflags` flags with the following strings:
//
// sigs.k8s.io/release-utils/version.gitVersion=<GitVersion>
// sigs.k8s.io/release-utils/version.gitCommit=<GitCommit>
// sigs.k8s.io/release-utils/version.gitTreeState=<GitTreeState>
// sigs.k8s.io/release-utils/version.buildDate=<BuildDate>
//
// Example: `go build -ldflags " -X sigs.k8s.io/release-utils/version.gitVersion=v0.4.0-1-g040f53c -X sigs.k8s.io/release-utils/version.gitCommit=040f53c -X sigs.k8s.io/release-utils/version.gitTreeState=dirty -X sigs.k8s.io/release-utils/version.buildDate=2022-02-03T17:30:01Z" .`
package version
