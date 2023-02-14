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

import (
	"os/exec"
)

// ExecuteChroot executes the named program with the given arguments
// inside a chroot.
// TODO(kaniini): Add support for using qemu-binfmt here for multiarch.
func (e *Executor) ExecuteChroot(name string, arg ...string) error {
	var cmd *exec.Cmd

	if e.UseProot {
		baseargs := []string{"-S", e.WorkDir}
		if e.UseQemu != "" {
			baseargs = append(baseargs, "-q", e.UseQemu)
		}
		baseargs = append(baseargs, name)
		arg = append(baseargs, arg...)
		cmd = exec.Command("proot", arg...)
	} else {
		arg = append([]string{e.WorkDir, name}, arg...)
		cmd = exec.Command("chroot", arg...)
	}

	return e.impl.Run(cmd, name, e.Log)
}

// Execute executes the named program with the given arguments.
func (e *Executor) Execute(name string, arg ...string) error {
	logname := name

	if e.UseProot {
		arg = append([]string{"-0", name}, arg...)
		name = "proot"
	}

	cmd := exec.Command(name, arg...)
	return e.impl.Run(cmd, logname, e.Log)
}
