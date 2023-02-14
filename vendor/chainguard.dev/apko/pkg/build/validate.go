// Copyright 2022 Chainguard, Inc.
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

package build

import (
	"fmt"
	"os"
	"path/filepath"
)

type Assertion func(*Context) error

func RequirePasswdFile(optional bool) Assertion {
	return func(bc *Context) error {
		path := filepath.Join(bc.Options.WorkDir, "etc", "passwd")

		_, err := os.Stat(path)
		if err != nil {
			if optional {
				bc.Logger().Warnf("%s is missing", path)
				return nil
			}
			return fmt.Errorf("/etc/passwd file is missing: %w", err)
		}
		return nil
	}
}

func RequireGroupFile(optional bool) Assertion {
	return func(bc *Context) error {
		path := filepath.Join(bc.Options.WorkDir, "etc", "group")

		_, err := os.Stat(path)
		if err != nil {
			if optional {
				bc.Logger().Warnf("%s is missing", path)
				return nil
			}
			return fmt.Errorf("/etc/group file is missing: %w", err)
		}
		return nil
	}
}
