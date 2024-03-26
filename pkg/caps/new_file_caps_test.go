// Copyright 2024 ko Build Authors All Rights Reserved.
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

package caps

import (
	"reflect"
	"strings"
	"testing"
)

func TestNewFileCaps(t *testing.T) {
	tests := []struct {
		args     []string
		res      *FileCaps
		mustFail bool
	}{
		{},
		{
			args: []string{"chown", "dac_override", "dac_read_search"},
			res:  &FileCaps{permitted: 7},
		},
		{
			args: []string{"chown,dac_override,dac_read_search=p"},
			res:  &FileCaps{permitted: 7},
		},
		{
			args: []string{"chown,dac_override,dac_read_search=i"},
			res:  &FileCaps{inheritable: 7},
		},
		{
			args: []string{"chown,dac_override,dac_read_search=e"},
		},
		{
			args: []string{"chown,dac_override,dac_read_search=pe"},
			res:  &FileCaps{permitted: 7, flags: FlagEffective},
		},
		{
			args: []string{"=pe"},
			res:  &FileCaps{permitted: allKnownCaps(), flags: FlagEffective},
		},
		{
			args: []string{"chown=ie", "chown=p"},
			res:  &FileCaps{permitted: 1},
		},
		{
			args: []string{"chown=ie", "chown="},
		},
		{
			args: []string{"chown=ie", "chown+p"},
			res:  &FileCaps{permitted: 1, inheritable: 1, flags: FlagEffective},
		},
		{
			args: []string{"chown=pie", "dac_override,chown-p"},
			res:  &FileCaps{inheritable: 1, flags: FlagEffective},
		},
		{args: []string{"chown,=pie"}, mustFail: true},
		{args: []string{"-pie"}, mustFail: true},
		{args: []string{"+pie"}, mustFail: true},
		{args: []string{"="}},
	}
	for _, tc := range tests {
		label := strings.Join(tc.args, ":")
		t.Run(label, func(t *testing.T) {
			res, err := NewFileCaps(tc.args...)
			if tc.mustFail && err == nil {
				t.Fatal("didn't fail")
			}
			if !tc.mustFail && err != nil {
				t.Fatalf("unexpectedly failed: %v", err)
			}
			if !reflect.DeepEqual(res, tc.res) {
				t.Fatalf("got %v expected %v", res, tc.res)
			}
		})
	}
}
