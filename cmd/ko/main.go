// Copyright 2018 Google LLC All Rights Reserved.
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

package main

// HEY! YOU! This file has moved to the root of the project.
// !! PLEASE DO NOT ADD NEW FEATURES HERE !!
// Only sync with the root/main.go.

import (
	"log"

	"github.com/google/ko/pkg/commands"
)

const Deprecation258 = `NOTICE!
-----------------------------------------------------------------
Please install ko from github.com/google/ko.

For more information see:
   https://github.com/google/ko/issues/258
-----------------------------------------------------------------
`

func main() {
	log.Print(Deprecation258)

	if err := commands.Root.Execute(); err != nil {
		log.Fatalf("error during command execution: %v", err)
	}
}
