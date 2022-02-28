// Copyright 2022 Google LLC All Rights Reserved.
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

import "log"

const bareBaseFlagsWarning = `WARNING!
-----------------------------------------------------------------
Both --base-import-paths and --bare were set.

--base-import-paths will take precedence and ignore --bare flag.

In a future release this will be an error.
-----------------------------------------------------------------
`

func Validate(po *PublishOptions, bo *BuildOptions) error {
	if po.Bare && po.BaseImportPaths {
		log.Print(bareBaseFlagsWarning)
		// TODO: return error when we decided to make this an error, for now it is a warning
	}

	return nil
}
