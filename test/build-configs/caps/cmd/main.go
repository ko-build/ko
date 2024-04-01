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

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

func permittedCaps() (uint64, error) {
	data, err := ioutil.ReadFile("/proc/self/status")
	if err != nil {
		return 0, err
	}
	const prefix = "CapPrm:"
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, prefix) {
			return strconv.ParseUint(strings.TrimSpace(line[len(prefix):]), 16, 64)
		}
	}
	return 0, fmt.Errorf("didn't find %#v in /proc/self/status", prefix)
}

func main() {
	caps, err := permittedCaps()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if caps == 0 {
		fmt.Println("No capabilities")
	} else {
		fmt.Printf("Has capabilities (%x)\n", caps)
	}
}
