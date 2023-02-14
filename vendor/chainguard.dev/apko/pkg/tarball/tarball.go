// Copyright 2022 Chainguard, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tarball

import (
	"time"
)

type Context struct {
	SourceDateEpoch time.Time
	OverrideUIDGID  bool
	UID             int
	GID             int
	OverrideUname   string
	OverrideGname   string
	SkipClose       bool
	UseChecksums    bool
}

type Option func(*Context) error

// Generates a new context from a set of options.
func NewContext(opts ...Option) (*Context, error) {
	ctx := Context{}

	for _, opt := range opts {
		if err := opt(&ctx); err != nil {
			return nil, err
		}
	}

	return &ctx, nil
}

// Sets SourceDateEpoch for Context.
func WithSourceDateEpoch(t time.Time) Option {
	return func(ctx *Context) error {
		ctx.SourceDateEpoch = t
		return nil
	}
}

// WithOverrideUIDGID sets the UID/GID to override with for Context.
func WithOverrideUIDGID(uid, gid int) Option {
	return func(ctx *Context) error {
		ctx.OverrideUIDGID = true
		ctx.UID = uid
		ctx.GID = gid
		return nil
	}
}

// WithOverrideUname sets the Uname to use with Context.
func WithOverrideUname(uname string) Option {
	return func(ctx *Context) error {
		ctx.OverrideUname = uname
		return nil
	}
}

// WithOverrideGname sets the Gname to use with Context.
func WithOverrideGname(gname string) Option {
	return func(ctx *Context) error {
		ctx.OverrideGname = gname
		return nil
	}
}

// WithSkipClose is used to determine whether the tar stream
// should be closed.  For concatenated tar streams such as APKv2
// containers, only the final tar stream should be closed.
func WithSkipClose(skipClose bool) Option {
	return func(ctx *Context) error {
		ctx.SkipClose = skipClose
		return nil
	}
}

// WithUseChecksums is used to determine whether the tar stream
// should have the APK-TOOLS.checksum.SHA1 extension.
func WithUseChecksums(useChecksums bool) Option {
	return func(ctx *Context) error {
		ctx.UseChecksums = useChecksums
		return nil
	}
}
