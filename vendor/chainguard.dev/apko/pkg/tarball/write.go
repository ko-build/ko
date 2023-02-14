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
	"archive/tar"
	"crypto/sha1" // nolint:gosec
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"syscall"

	gzip "golang.org/x/build/pargzip"

	apkofs "chainguard.dev/apko/pkg/fs"
)

func (ctx *Context) writeArchiveFromFS(dst io.Writer, fsys fs.FS) error {
	gzw := gzip.NewWriter(dst)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	if !ctx.SkipClose {
		defer tw.Close()
	} else {
		defer tw.Flush()
	}

	return ctx.writeTar(tw, fsys)
}

func hasHardlinks(fi fs.FileInfo) bool {
	if stat := fi.Sys(); stat != nil {
		si := stat.(*syscall.Stat_t)

		// if we don't have inodes, we just assume the filesystem
		// does not support hardlinks
		if si == nil {
			return false
		}

		return si.Nlink > 1
	}

	return false
}

func getInodeFromFileInfo(fi fs.FileInfo) (uint64, error) {
	if stat := fi.Sys(); stat != nil {
		si := stat.(*syscall.Stat_t)

		// if we don't have inodes, we just assume the filesystem
		// does not support hardlinks
		if si == nil {
			return 0, fmt.Errorf("unable to stat underlying file")
		}

		return si.Ino, nil
	}

	return 0, fmt.Errorf("unable to stat underlying file")
}

func (ctx *Context) writeTar(tw *tar.Writer, fsys fs.FS) error {
	seenFiles := map[uint64]string{}

	if err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		// skip the root path, superfluous
		if path == "." {
			return nil
		}

		if err != nil {
			return err
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		var link string
		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			rlfs, ok := fsys.(apkofs.ReadLinkFS)
			if !ok {
				return fmt.Errorf("readlink not supported by this fs: path (%s)", path)
			}

			if link, err = rlfs.Readlink(path); err != nil {
				return err
			}
		}

		header, err := tar.FileInfoHeader(info, link)
		if err != nil {
			return err
		}
		// work around some weirdness, without this we wind up with just the basename
		header.Name = path

		// zero out timestamps for reproducibility
		header.AccessTime = ctx.SourceDateEpoch
		header.ModTime = ctx.SourceDateEpoch
		header.ChangeTime = ctx.SourceDateEpoch

		if ctx.OverrideUIDGID {
			header.Uid = ctx.UID
			header.Gid = ctx.GID
		}

		if ctx.OverrideUname != "" {
			header.Uname = ctx.OverrideUname
		}

		if ctx.OverrideGname != "" {
			header.Gname = ctx.OverrideGname
		}

		if !info.IsDir() && hasHardlinks(info) {
			inode, err := getInodeFromFileInfo(info)
			if err != nil {
				return err
			}

			if oldpath, ok := seenFiles[inode]; ok {
				header.Typeflag = tar.TypeLink
				header.Linkname = oldpath
				header.Size = 0
			} else {
				seenFiles[inode] = header.Name
			}
		}

		if ctx.UseChecksums {
			header.PAXRecords = map[string]string{}

			if link != "" {
				linkDigest := sha1.Sum([]byte(link)) // nolint:gosec
				linkChecksum := hex.EncodeToString(linkDigest[:])
				header.PAXRecords["APK-TOOLS.checksum.SHA1"] = linkChecksum
			} else if info.Mode().IsRegular() {
				data, err := fsys.Open(path)
				if err != nil {
					return err
				}
				defer data.Close()

				fileDigest := sha1.New() // nolint:gosec
				if _, err := io.Copy(fileDigest, data); err != nil {
					return err
				}

				fileChecksum := hex.EncodeToString(fileDigest.Sum(nil))
				header.PAXRecords["APK-TOOLS.checksum.SHA1"] = fileChecksum
			}
		}

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if info.Mode().IsRegular() && header.Size > 0 {
			data, err := fsys.Open(path)
			if err != nil {
				return err
			}

			defer data.Close()

			if _, err := io.Copy(tw, data); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

// WriteArchive writes a tarball to the provided io.Writer.
func (ctx *Context) WriteArchive(dst io.Writer, src fs.FS) error {
	if err := ctx.writeArchiveFromFS(dst, src); err != nil {
		return fmt.Errorf("writing TAR archive failed: %w", err)
	}

	return nil
}
