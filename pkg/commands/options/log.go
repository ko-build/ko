// Copyright 2025 ko Build Authors All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package options

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/google/go-containerregistry/pkg/logs"
	"github.com/spf13/cobra"
)

func AddLogOptions(root *cobra.Command, oo *LogOptions) {
	root.PersistentFlags().BoolVarP(&oo.Verbose, "verbose", "v", false, "Enable debug logs")
	root.PersistentFlags().BoolVarP(&oo.Structured, "structured", "s", false, "Enable structured logs")
}

type LogOptions struct {
	Verbose    bool
	Structured bool
}

func (lo *LogOptions) SetOutput(cmd *cobra.Command) {
	opts := &slog.HandlerOptions{
		Level:       slog.LevelDebug,
		ReplaceAttr: replaceAttr,
	}

	var handler slog.Handler

	if lo.Structured {
		handler = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		handler = slog.NewTextHandler(os.Stderr, opts)
	}

	slog.SetDefault(slog.New(handler))

	if lo.Verbose {
		logs.Warn = slog.NewLogLogger(handler, slog.LevelWarn)
		logs.Debug = slog.NewLogLogger(handler, slog.LevelDebug)
	}

	logs.Progress = slog.NewLogLogger(handler, slog.LevelInfo)

	cmd.SetErr(&slogWriter{slog.New(handler)})
}

type slogWriter struct {
	*slog.Logger
}

func (w *slogWriter) Write(p []byte) (n int, err error) {
	w.Log(context.Background(), slog.LevelError, string(p))
	return len(p), nil
}

var replacer = strings.NewReplacer(
	"\r", " ",
	"\n", "",
	"\"", "'",
)

func replaceAttr(groups []string, a slog.Attr) slog.Attr {
	switch a.Key {
	case "msg":
		v := replacer.Replace(a.Value.String())
		v = strings.TrimSpace(v)
		return slog.Attr{
			Key:   a.Key,
			Value: slog.StringValue(v),
		}
	}

	return a
}
