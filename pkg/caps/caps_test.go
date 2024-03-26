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
	"encoding/base64"
	"fmt"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		arg      string
		res      Mask
		mustFail bool
	}{
		{arg: "chown", res: 1},
		{arg: "cap_chown", res: 1},
		{arg: "cAp_cHoWn", res: 1},
		{arg: "unknown", mustFail: true},
		{arg: "63", res: 1 << 63},
		{arg: "64", mustFail: true},
		{arg: "all", res: allKnownCaps()},
	}
	for _, tc := range tests {
		t.Run(tc.arg, func(t *testing.T) {
			mask, err := Parse(tc.arg)
			if err == nil && tc.mustFail {
				t.Fatal("invalid input accepted")
			}
			if err != nil && !tc.mustFail {
				t.Fatal(err)
			}
			if mask != tc.res {
				t.Fatalf("unexpected result: %x", mask)
			}
		})
	}
}

//go:generate ./gen.sh

type ddTest struct {
	permitted, inheritable string
	effective              bool
	res                    string
}

func TestDd(t *testing.T) {
	for _, test := range ddTests {
		label := fmt.Sprintf("%s,%s,%v", test.permitted, test.inheritable, test.effective)
		t.Run(label, func(t *testing.T) {
			var permitted, inheritable Mask
			var flags Flags

			if test.permitted != "" {
				mask, err := Parse(test.permitted)
				if err != nil {
					t.Fatal(err)
				}
				permitted = mask
			}

			if test.inheritable != "" {
				mask, err := Parse(test.inheritable)
				if err != nil {
					t.Fatal(err)
				}
				inheritable = mask
			}

			if test.effective {
				flags = FlagEffective
			}

			res, err := XattrBytes(permitted, inheritable, flags)
			if err != nil {
				t.Fatal(err)
			}

			resBase64 := make([]byte, base64.StdEncoding.EncodedLen(len(res)))
			base64.StdEncoding.Encode(resBase64, res)
			if string(resBase64) != test.res {
				t.Fatalf("expected %s, result %s", test.res, resBase64)
			}
		})
	}
}
