// Copyright 2019 Google LLC All Rights Reserved.
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

package options

import (
	"github.com/spf13/cobra"
)

// DebugOptions holds options to improve debugging containers.
type DebugOptions struct {
	DisableOptimizations bool
}

func AddDebugArg(cmd *cobra.Command, do *DebugOptions) {
	cmd.Flags().BoolVar(&do.DisableOptimizations, "disable-optimizations", do.DisableOptimizations,
		"Disable optimizations when building Go code. Useful when you want to interactively debug the created container.")
}
