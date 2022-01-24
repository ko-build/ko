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

	"github.com/go-logr/logr"
)

// L returns a logr.Logger from the provided context.
// If the context is nil or doesn't contain a logr.Logger, it returns a logr.Logger that discards all messages.
func L(ctx context.Context) logr.Logger {
	if ctx == nil {
		fmt.Fprintln(os.Stderr, "nil context, discarding log messages")
		return logr.Discard()
	}
	l := logr.FromContextOrDiscard(ctx)
	if l == logr.Discard() { // the Discard() return value is comparable
		fmt.Fprintln(os.Stderr, "no logger in context, discarding log messages")
		return l
	}
	return l
}
