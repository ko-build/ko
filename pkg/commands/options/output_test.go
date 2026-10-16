// Copyright 2026 ko Build Authors All Rights Reserved.
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
	"testing"

	"github.com/spf13/cobra"
)

func TestOutputFlag(t *testing.T) {
	for _, tc := range []struct {
		name    string
		args    []string
		want    string
		wantErr bool
	}{
		{name: "default", args: nil, want: OutputText},
		{name: "long", args: []string{"--output", "json"}, want: OutputJSON},
		{name: "alias", args: []string{"--out=json"}, want: OutputJSON},
		{name: "short", args: []string{"-O", "text"}, want: OutputText},
		{name: "invalid", args: []string{"-O", "yaml"}, want: "yaml", wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			oo := &OutputOptions{}
			cmd := &cobra.Command{Use: "test"}
			AddOutputArg(cmd, oo)
			if err := cmd.ParseFlags(tc.args); err != nil {
				t.Fatalf("ParseFlags() = %v", err)
			}
			if oo.Output != tc.want {
				t.Errorf("Output = %q, want %q", oo.Output, tc.want)
			}
			if err := oo.Validate(); (err != nil) != tc.wantErr {
				t.Errorf("Validate() = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
