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
	"io"

	"github.com/go-logr/logr"
)

// Writer wraps a logr.Logger and implements io.Writer.
// This is useful for redirecting output of other loggers,
// e.g., Go's standard library logger.
type Writer struct {
	io.Writer

	logger logr.Logger
}

// NewWriter returns a log.Writer that wraps the provided logr.Logger
func NewWriter(logger logr.Logger) *Writer {
	return &Writer{
		logger: logger,
	}
}

// Write implements io.Writer
func (w *Writer) Write(p []byte) (n int, err error) {
	numBytes := len(p)
	if numBytes > 0 && p[numBytes-1] == '\n' {
		p = p[:numBytes-1]
	}
	w.logger.Info(string(p))
	return len(p), nil
}
