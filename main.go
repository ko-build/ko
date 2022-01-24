/*
Copyright 2020 Google LLC All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	stdlog "log"
	"os"
	"os/signal"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"github.com/google/go-containerregistry/pkg/logs"

	"github.com/google/ko/pkg/commands"
	"github.com/google/ko/pkg/log"
)

const (
	defaultVerbosity = 1
)

func main() {
	ctx := context.Background()
	ctx, stop := signal.NotifyContext(configureLogging(ctx), os.Interrupt)
	defer stop()
	if err := commands.Root.ExecuteContext(ctx); err != nil {
		log.Fatal(ctx, "error during command execution:", err)
	}
}

// configureLogging log logr logs logger logging
func configureLogging(ctx context.Context) context.Context {
	logger := defaultLogger()
	configureGGCRLogging(logger)
	return logr.NewContext(ctx, logger)
}

// defaultLogger returns a logr.Logger instance that wraps log.Logger from Go's standard library.
// This function is placed here so `pkg` does not depend on the `stdr` module.
func defaultLogger() logr.Logger {
	stdLogger := stdlog.Default()
	logger := stdr.New(stdLogger).V(defaultVerbosity).WithName("ko")
	stdr.SetVerbosity(defaultVerbosity)
	return logger
}

// configureGGCRLogging maps ggcr log levels as follows:
//
// ggcr Warn     -> logr V(0)
// ggcr Progress -> logr V(defaultVerbosity)
// ggcr Debug    -> logr V(defaultVerbosity + 1)
func configureGGCRLogging(logger logr.Logger) {
	ggcrLogger := logger.WithName("go-containerregistry")
	logs.Warn.SetFlags(0)
	logs.Warn.SetOutput(log.NewWriter(ggcrLogger.V(-1)))
	logs.Progress.SetFlags(0)
	logs.Progress.SetOutput(log.NewWriter(ggcrLogger))
	logs.Debug.SetFlags(0)
	logs.Debug.SetOutput(log.NewWriter(ggcrLogger.V(1)))
}
