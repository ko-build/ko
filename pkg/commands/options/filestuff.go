// Copyright 2018 ko Build Authors All Rights Reserved.
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

package options

import (
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// FilenameOptions is from pkg/kubectl.
type FilenameOptions struct {
	Filenames []string
	Recursive bool
}

func AddFileArg(cmd *cobra.Command, fo *FilenameOptions) {
	// From pkg/kubectl
	cmd.Flags().StringSliceVarP(&fo.Filenames, "filename", "f", fo.Filenames,
		"Filename, directory, or URL to files to use to create the resource")
	cmd.Flags().BoolVarP(&fo.Recursive, "recursive", "R", fo.Recursive,
		"Process the directory used in -f, --filename recursively. Useful when you want to manage related manifests organized within the same directory.")
}

// Based heavily on pkg/kubectl
func EnumerateFiles(fo *FilenameOptions) chan string {
	files := make(chan string)
	go func() {
		// When we're done enumerating files, close the channel
		defer close(files)
		for _, paths := range fo.Filenames {
			// Just pass through '-' as it is indicative of stdin.
			if paths == "-" {
				files <- paths
				continue
			}
			if err := enumerateFiles(paths, paths, fo.Recursive, map[string]struct{}{}, files); err != nil {
				log.Fatalf("Error enumerating files: %v", err)
			}
		}
	}()
	return files
}

func enumerateFiles(root, filename string, recursive bool, visited map[string]struct{}, files chan<- string) error {
	fi, err := os.Lstat(filename)
	if err != nil {
		return err
	}

	isDir := fi.IsDir()
	if fi.Mode()&os.ModeSymlink != 0 {
		targetInfo, err := os.Stat(filename)
		if err != nil {
			// Preserve the old behavior for file-like symlinks: if the entry was
			// passed explicitly, stream it and let later file handling report errors.
			if filename == root {
				files <- filename
			}
			return nil
		}
		isDir = targetInfo.IsDir()
	}

	if isDir {
		if filename != root && !recursive {
			return nil
		}
		resolved, err := filepath.EvalSymlinks(filename)
		if err != nil {
			return err
		}
		abs, err := filepath.Abs(resolved)
		if err != nil {
			return err
		}
		if _, ok := visited[abs]; ok {
			return nil
		}
		visited[abs] = struct{}{}

		entries, err := os.ReadDir(filename)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if err := enumerateFiles(root, filepath.Join(filename, entry.Name()), recursive, visited, files); err != nil {
				return err
			}
		}
		return nil
	}

	// Don't check extension if the filepath was passed explicitly.
	if filename != root {
		switch filepath.Ext(filename) {
		case ".json", ".yaml":
			// Process these.
		default:
			return nil
		}
	}

	files <- filename
	return nil
}
