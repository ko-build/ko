// Copyright 2025 ko Build Authors All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package options

import (
	"os"

	"github.com/google/ko/pkg/build"
	"github.com/spf13/cobra"
)

type OutputOptions struct {
	Json   bool
	Indent bool
}

func AddOutputOptions(cmd *cobra.Command, oo *OutputOptions) {
	cmd.Flags().BoolVar(&oo.Json, "json", false, "Enable json-structured output")
	cmd.Flags().BoolVar(&oo.Indent, "pretty", true, "Indent json output (pretty-print)")
}

func (oo *OutputOptions) Printer() build.Printer {
	if oo.Json {
		return build.NewJSONPrinter(os.Stdout, oo.Indent)
	}

	return build.NewPrinter(os.Stdout)
}
