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

// GroupEntry describes a single line in /etc/group.
type GroupEntry struct {
	GroupName string
	Password  string
	GID       uint32
	Members   []string
}

// GroupFile describes an entire /etc/group file's contents.
type GroupFile struct {
	Entries []GroupEntry
}

// ReadOrCreateGroupFile parses an /etc/group file into a GroupFile.
// An empty file is created if /etc/group is missing.
func ReadOrCreateGroupFile(filePath string) (GroupFile, error) {
	gf := GroupFile{}

	file, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0o644)
	if err != nil {
		return gf, fmt.Errorf("failed to open %s: %w", filePath, err)
	}
	defer file.Close()

	if err := gf.Load(file); err != nil {
		return gf, err
	}

	return gf, nil
}

// Load loads an /etc/passwd file into a GroupFile from an io.Reader.
func (gf *GroupFile) Load(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		ge := GroupEntry{}

		if err := ge.Parse(scanner.Text()); err != nil {
			return fmt.Errorf("unable to parse: %w", err)
		}

		gf.Entries = append(gf.Entries, ge)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("unable to parse: %w", err)
	}

	return nil
}

// WriteFile writes an /etc/passwd file from a GroupFile.
func (gf *GroupFile) WriteFile(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("unable to open %s for writing: %w", filePath, err)
	}
	defer file.Close()

	return gf.Write(file)
}

// Write writes an /etc/passwd file into an io.Writer.
func (gf *GroupFile) Write(w io.Writer) error {
	for _, ge := range gf.Entries {
		if err := ge.Write(w); err != nil {
			return fmt.Errorf("unable to write group entry: %w", err)
		}
	}

	return nil
}

// Parse parses an /etc/group line into a GroupEntry.
func (ge *GroupEntry) Parse(line string) error {
	line = strings.TrimSpace(line)

	parts := strings.Split(line, ":")
	if len(parts) != 4 {
		return fmt.Errorf("malformed line, contains %d parts, expecting 4", len(parts))
	}

	ge.GroupName = parts[0]
	ge.Password = parts[1]

	gid, err := strconv.Atoi(parts[2])
	if err != nil {
		return fmt.Errorf("failed to parse UID %s", parts[2])
	}
	ge.GID = uint32(gid)

	ge.Members = strings.Split(parts[3], ",")

	return nil
}

// Write writes an /etc/group line into an io.Writer.
func (ge *GroupEntry) Write(w io.Writer) error {
	members := strings.Join(ge.Members, ",")
	_, err := fmt.Fprintf(w, "%s:%s:%d:%s\n", ge.GroupName, ge.Password, ge.GID, members)
	return err
}
