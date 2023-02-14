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

package s6

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func (sc *Context) CreateSupervisionDirectory(name string) (string, error) {
	svbase := filepath.Join(sc.WorkDir, "sv")
	svcdir := filepath.Join(svbase, name)
	sc.Log.Debugf("  supervision dir: %s", svcdir)

	if err := os.MkdirAll(svcdir, 0755); err != nil {
		return svcdir, fmt.Errorf("could not make supervision directory: %w", err)
	}

	if err := os.Chmod(svcdir, 0777); err != nil { // nolint:gosec
		return svcdir, fmt.Errorf("could not set permissions on supervision dir: %w", err)
	}

	if err := os.Chmod(svbase, 0777); err != nil { // nolint:gosec
		return svcdir, fmt.Errorf("could not set permissions on base supervision dir: %w", err)
	}

	return svcdir, nil
}

func (sc *Context) WriteSupervisionTemplate(svcdir string, command string) error {
	file, err := os.Create(filepath.Join(svcdir, "run"))
	if err != nil {
		return fmt.Errorf("could not create runfile: %w", err)
	}
	defer file.Close()

	if err := os.Chmod(file.Name(), 0755); err != nil {
		return fmt.Errorf("could not set permissions on runfile: %w", err)
	}

	fmt.Fprintf(file, "#!/bin/execlineb\n%s\n", command)

	return nil
}

func (sc *Context) WriteSupervisionServiceSimple(name string, command string) error {
	sc.Log.Debugf("simple service: %s => %s", name, command)

	svcdir, err := sc.CreateSupervisionDirectory(name)
	if err != nil {
		return err
	}

	if err := sc.WriteSupervisionTemplate(svcdir, command); err != nil {
		return err
	}

	return nil
}

func (sc *Context) WriteSupervisionTree(services Services) error {
	sc.Log.Infof("generating supervision tree")

	// generate the leaves
	for service, descriptor := range services {
		service, ok := service.(string)
		if !ok {
			return errors.New("service name is not string")
		}

		if svccmd, ok := descriptor.(string); ok {
			if err := sc.WriteSupervisionServiceSimple(service, svccmd); err != nil {
				return err
			}
		} else {
			return errors.New("complex services are not yet supported")
		}
	}

	return nil
}
