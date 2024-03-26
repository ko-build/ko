// Copyright 2018 ko Build Authors All Rights Reserved.
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
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	// Give this an interesting import
	_ "github.com/google/go-containerregistry/pkg/registry"
)

var (
	f    = flag.String("f", "kenobi", "File in kodata to print")
	wait = flag.Bool("wait", false, "Whether to wait for SIGTERM")
)

// This is defined so we can test build-time variable setting using ldflags.
var version = "default"

func main() {
	flag.Parse()

	log.Println("version =", version)

	if runtime.GOOS == "windows" {
		// Go seems to not load location data from Windows, so timezone
		// conversion fails unless tzdata is embedded in the Go program
		// with the go build tag `timetzdata`. Since we want to test
		// loading tzdata provided by the base image below, we'll just
		// skip that for Windows here.
		// See https://github.com/ko-build/ko/issues/739
		log.Println("skipping timezone conversion on Windows")
	} else {
		// Exercise timezone conversions, which demonstrates tzdata is provided
		// by the base image.
		now := time.Now()
		loc, _ := time.LoadLocation("UTC")
		fmt.Printf("UTC Time: %s\n", now.In(loc))
		loc, _ = time.LoadLocation("America/New_York")
		fmt.Printf("New York Time: %s\n", now.In(loc))
	}

	dp := os.Getenv("KO_DATA_PATH")
	file := filepath.Join(dp, *f)
	bytes, err := os.ReadFile(file)
	if err != nil {
		log.Fatalf("Error reading %q: %v", file, err)
	}
	fmt.Println(string(bytes))

	// Cause the pod to "hang" to allow us to check for a readiness state.
	if *wait {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGTERM)
		<-sigs
	}
}
