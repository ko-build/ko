/*
Copyright 2019 Google LLC All Rights Reserved.

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

package options

import (
	"github.com/google/ko/pkg/build"
	"github.com/spf13/cobra"
)

// BuildOptions represents options for the ko builder.
type BuildOptions struct {
	// BaseImage enables setting the default base image programmatically.
	// If non-empty, this takes precedence over the value in `.ko.yaml`.
	BaseImage string

	// WorkingDirectory allows for setting the working directory for invocations of the `go` tool.
	// Empty string means the current working directory.
	WorkingDirectory string

	ConcurrentBuilds     int
	DisableOptimizations bool
	Platform             string
	Labels               []string
	// UserAgent enables overriding the default value of the `User-Agent` HTTP
	// request header used when retrieving the base image.
	UserAgent string

	InsecureRegistry bool

	// BuildConfigs enables programmatic overriding of build config set in `.ko.yaml`.
	BuildConfigs map[string]build.Config
}

func AddBuildOptions(cmd *cobra.Command, bo *BuildOptions) {
	cmd.Flags().IntVarP(&bo.ConcurrentBuilds, "jobs", "j", 0,
		"The maximum number of concurrent builds (default GOMAXPROCS)")
	cmd.Flags().BoolVar(&bo.DisableOptimizations, "disable-optimizations", bo.DisableOptimizations,
		"Disable optimizations when building Go code. Useful when you want to interactively debug the created container.")
	cmd.Flags().StringVar(&bo.Platform, "platform", "",
		"Which platform to use when pulling a multi-platform base. Format: all | <os>[/<arch>[/<variant>]][,platform]*")
	cmd.Flags().StringSliceVar(&bo.Labels, "image-label", []string{},
		"Which labels (key=value) to add to the image.")
}
