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

package log

import (
	"context"
	"fmt"
	"os"
)

// The functions below are provided to help transition from the log package in
// Go's standard library to go-logr.
// For new code, please use the L() function to obtain a logr.Logger and use
// its methods to log.

// Fatal is intended as a replacement for log.Fatal.
// It takes an extra Context argument that it uses to find a logr.Logger.
func Fatal(ctx context.Context, v ...interface{}) {
	err := getError(v...)
	L(ctx).Error(err, fmt.Sprint(v...))
	os.Exit(1)
}

// Fatalf is intended as a replacement for log.Fatalf.
// It takes an extra Context argument that it uses to find a logr.Logger.
func Fatalf(ctx context.Context, format string, v ...interface{}) {
	err := getError(v...)
	L(ctx).Error(err, fmt.Sprintf(format, v...))
	os.Exit(1)
}

// Fatalln is intended as a replacement for log.Fatalln.
// It takes an extra Context argument that it uses to find a logr.Logger.
func Fatalln(ctx context.Context, v ...interface{}) {
	err := getError(v...)
	L(ctx).Error(err, fmt.Sprintln(v...))
	os.Exit(1)
}

// Print is intended as a replacement for log.Print.
// It takes an extra Context argument that it uses to find a logr.Logger.
func Print(ctx context.Context, v ...interface{}) {
	L(ctx).Info(fmt.Sprint(v...))
}

// Printf is intended as a replacement for log.Printf.
// It takes an extra Context argument that it uses to find a logr.Logger.
func Printf(ctx context.Context, format string, v ...interface{}) {
	L(ctx).Info(fmt.Sprintf(format, v...))
}

// Println is intended as a replacement for log.Println.
// It takes an extra Context argument that it uses to find a logr.Logger.
func Println(ctx context.Context, v ...interface{}) {
	L(ctx).Info(fmt.Sprintln(v...))
}

// getError returns the first error from the arguments.
// If there are no errors, the function returns nil.
func getError(v ...interface{}) error {
	var err error
	for _, val := range v {
		switch value := val.(type) {
		case error:
			err = value
		}
	}
	return err
}
