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

package passwd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// UserEntry contains the parsed data from an /etc/passwd entry.
type UserEntry struct {
	UserName string
	Password string
	UID      uint32
	GID      uint32
	Info     string
	HomeDir  string
	Shell    string
}

// UserFile contains the entries from an /etc/passwd file.
type UserFile struct {
	Entries []UserEntry
}

// ReadOrCreateUserFile parses an /etc/passwd file into a UserFile.
// An empty file is created if /etc/passwd is missing.
func ReadOrCreateUserFile(filePath string) (UserFile, error) {
	uf := UserFile{}

	file, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0o644)
	if err != nil {
		return uf, fmt.Errorf("failed to open %s: %w", filePath, err)
	}
	defer file.Close()

	if err := uf.Load(file); err != nil {
		return uf, err
	}

	return uf, nil
}

// Load loads an /etc/passwd file into a UserFile from an io.Reader.
func (uf *UserFile) Load(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		ue := UserEntry{}

		if err := ue.Parse(scanner.Text()); err != nil {
			return fmt.Errorf("unable to parse: %w", err)
		}

		uf.Entries = append(uf.Entries, ue)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("unable to parse: %w", err)
	}

	return nil
}

// WriteFile writes an /etc/passwd file from a UserFile.
func (uf *UserFile) WriteFile(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("unable to open %s for writing: %w", filePath, err)
	}
	defer file.Close()

	return uf.Write(file)
}

// Write writes an /etc/passwd file into an io.Writer.
func (uf *UserFile) Write(w io.Writer) error {
	for _, ue := range uf.Entries {
		if err := ue.Write(w); err != nil {
			return fmt.Errorf("unable to write passwd entry: %w", err)
		}
	}

	return nil
}

// Parse parses an /etc/passwd line into a UserEntry.
func (ue *UserEntry) Parse(line string) error {
	line = strings.TrimSpace(line)

	parts := strings.Split(line, ":")
	if len(parts) != 7 {
		return fmt.Errorf("malformed line, contains %d parts, expecting 7", len(parts))
	}

	ue.UserName = parts[0]
	ue.Password = parts[1]

	uid, err := strconv.Atoi(parts[2])
	if err != nil {
		return fmt.Errorf("failed to parse UID %s", parts[2])
	}
	ue.UID = uint32(uid)

	gid, err := strconv.Atoi(parts[3])
	if err != nil {
		return fmt.Errorf("failed to parse GID %s", parts[3])
	}
	ue.GID = uint32(gid)

	ue.Info = parts[4]
	ue.HomeDir = parts[5]
	ue.Shell = parts[6]

	return nil
}

// Write writes an /etc/passwd line into an io.Writer.
func (ue *UserEntry) Write(w io.Writer) error {
	_, err := fmt.Fprintf(w, "%s:%s:%d:%d:%s:%s:%s\n", ue.UserName, ue.Password, ue.UID, ue.GID, ue.Info, ue.HomeDir, ue.Shell)
	return err
}
