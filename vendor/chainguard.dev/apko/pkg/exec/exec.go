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

package exec

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"fmt"
	"os/exec"

	"github.com/sirupsen/logrus"
)

type Executor struct {
	impl     executorImplementation
	WorkDir  string
	UseProot bool
	UseQemu  string
	Log      *logrus.Entry
}

type Option func(*Executor) error

func New(workDir string, logger *logrus.Entry, opts ...Option) (*Executor, error) {
	e := &Executor{
		impl:    &defaultBuildImplementation{},
		WorkDir: workDir,
	}

	for _, opt := range opts {
		if err := opt(e); err != nil {
			return nil, err
		}
	}

	e.Log = logger.WithFields(logrus.Fields{"use-proot": e.UseProot, "use-qemu": e.UseQemu})

	return e, nil
}

func WithProot(proot bool) Option {
	return func(e *Executor) error {
		e.UseProot = proot
		return nil
	}
}

func WithQemu(qemuArch string) Option {
	return func(e *Executor) error {
		paths := []string{
			fmt.Sprintf("qemu-%s", qemuArch),
			fmt.Sprintf("qemu-%s-static", qemuArch),
		}

		for _, path := range paths {
			emu, err := exec.LookPath(path)
			if err == nil {
				e.UseQemu = emu
				return nil
			}
		}

		return fmt.Errorf("unable to find qemu emulator for %s on $PATH, is qemu-user or qemu-user-static installed?", qemuArch)
	}
}

func (e *Executor) SetImplementation(i executorImplementation) {
	e.impl = i
}
