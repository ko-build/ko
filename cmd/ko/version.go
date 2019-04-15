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

package main

import (
	"fmt"
	"log"
	"os/exec"
)

// provided by govvv in compile-time
var Version string

func version() {
	if Version == "" {
		hash, err := gitRevParseHead()
		if err != nil {
			log.Fatalf("error during command execution: %v", err)
		}
		fmt.Printf("version: %v", string(hash))
	} else {
		fmt.Printf("version: %v\n", Version)
	}
}

func gitRevParseHead() ([]byte, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	return cmd.Output()
}
