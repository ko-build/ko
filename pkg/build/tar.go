package build

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"golang.org/x/tools/go/packages"
)

// userOwnerAndGroupSID is a magic value needed to make the binary executable
// in a Windows container.
//
// owner: BUILTIN/Users group: BUILTIN/Users ($sddlValue="O:BUG:BU")
const userOwnerAndGroupSID = "AQAAgBQAAAAkAAAAAAAAAAAAAAABAgAAAAAABSAAAAAhAgAAAQIAAAAAAAUgAAAAIQIAAA=="

// Where kodata lives in the image.
const kodataRoot = "/var/run/ko"

func tarBinary(name, binary string, platform *v1.Platform) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	// Write the parent directories to the tarball archive.
	// For Windows, the layer must contain a Hives/ directory, and the root
	// of the actual filesystem goes in a Files/ directory.
	// For Linux, the binary goes into /ko-app/
	dirs := []string{"ko-app"}
	if platform.OS == "windows" {
		dirs = []string{
			"Hives",
			"Files",
			"Files/ko-app",
		}
		name = "Files" + name
	}
	for _, dir := range dirs {
		if err := tw.WriteHeader(&tar.Header{
			Name:     dir,
			Typeflag: tar.TypeDir,
			// Use a fixed Mode, so that this isn't sensitive to the directory and umask
			// under which it was created. Additionally, windows can only set 0222,
			// 0444, or 0666, none of which are executable.
			Mode: 0555,
		}); err != nil {
			return nil, fmt.Errorf("writing dir %q: %w", dir, err)
		}
	}

	file, err := os.Open(binary)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	header := &tar.Header{
		Name:     name,
		Size:     stat.Size(),
		Typeflag: tar.TypeReg,
		// Use a fixed Mode, so that this isn't sensitive to the directory and umask
		// under which it was created. Additionally, windows can only set 0222,
		// 0444, or 0666, none of which are executable.
		Mode: 0555,
	}
	if platform.OS == "windows" {
		// This magic value is for some reason needed for Windows to be
		// able to execute the binary.
		header.PAXRecords = map[string]string{
			"MSWINDOWS.rawsd": userOwnerAndGroupSID,
		}
	}
	// write the header to the tarball archive
	if err := tw.WriteHeader(header); err != nil {
		return nil, err
	}
	// copy the file data to the tarball
	if _, err := io.Copy(tw, file); err != nil {
		return nil, err
	}

	return buf, nil
}

func kodataPath(dir string, ref reference) (string, error) {
	dir = filepath.Clean(dir)
	if dir == "." {
		dir = ""
	}
	pkgs, err := packages.Load(&packages.Config{Dir: dir, Mode: packages.NeedFiles}, ref.Path())
	if err != nil {
		return "", fmt.Errorf("error loading package from %s: %w", ref.Path(), err)
	}
	if len(pkgs) != 1 {
		return "", fmt.Errorf("found %d local packages, expected 1", len(pkgs))
	}
	if len(pkgs[0].GoFiles) == 0 {
		return "", fmt.Errorf("package %s contains no Go files", pkgs[0])
	}
	return filepath.Join(filepath.Dir(pkgs[0].GoFiles[0]), "kodata"), nil
}


// walkRecursive performs a filepath.Walk of the given root directory adding it
// to the provided tar.Writer with root -> chroot.  All symlinks are dereferenced,
// which is what leads to recursion when we encounter a directory symlink.
func walkRecursive(tw *tar.Writer, root, chroot string, creationTime v1.Time, platform *v1.Platform) error {
	return filepath.Walk(root, func(hostPath string, info os.FileInfo, err error) error {
		if hostPath == root {
			return nil
		}
		if err != nil {
			return fmt.Errorf("filepath.Walk(%q): %w", root, err)
		}
		// Skip other directories.
		if info.Mode().IsDir() {
			return nil
		}
		newPath := path.Join(chroot, filepath.ToSlash(hostPath[len(root):]))

		// Don't chase symlinks on Windows, where cross-compiled symlink support is not possible.
		if platform.OS == "windows" {
			if info.Mode()&os.ModeSymlink != 0 {
				log.Println("skipping symlink in kodata for windows:", info.Name())
				return nil
			}
		}

		evalPath, err := filepath.EvalSymlinks(hostPath)
		if err != nil {
			return fmt.Errorf("filepath.EvalSymlinks(%q): %w", hostPath, err)
		}

		// Chase symlinks.
		info, err = os.Stat(evalPath)
		if err != nil {
			return fmt.Errorf("os.Stat(%q): %w", evalPath, err)
		}
		// Skip other directories.
		if info.Mode().IsDir() {
			return walkRecursive(tw, evalPath, newPath, creationTime, platform)
		}

		// Open the file to copy it into the tarball.
		file, err := os.Open(evalPath)
		if err != nil {
			return fmt.Errorf("os.Open(%q): %w", evalPath, err)
		}
		defer file.Close()

		// Copy the file into the image tarball.
		header := &tar.Header{
			Name:     newPath,
			Size:     info.Size(),
			Typeflag: tar.TypeReg,
			// Use a fixed Mode, so that this isn't sensitive to the directory and umask
			// under which it was created. Additionally, windows can only set 0222,
			// 0444, or 0666, none of which are executable.
			Mode:    0555,
			ModTime: creationTime.Time,
		}
		if platform.OS == "windows" {
			// This magic value is for some reason needed for Windows to be
			// able to execute the binary.
			header.PAXRecords = map[string]string{
				"MSWINDOWS.rawsd": userOwnerAndGroupSID,
			}
		}
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("tar.Writer.WriteHeader(%q): %w", newPath, err)
		}
		if _, err := io.Copy(tw, file); err != nil {
			return fmt.Errorf("io.Copy(%q, %q): %w", newPath, evalPath, err)
		}
		return nil
	})
}

func tarKoData(dir string, ref reference, platform *v1.Platform, creationTime v1.Time) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	root, err := kodataPath(dir, ref)
	if err != nil {
		return nil, err
	}

	// Write the parent directories to the tarball archive.
	// For Windows, the layer must contain a Hives/ directory, and the root
	// of the actual filesystem goes in a Files/ directory.
	// For Linux, kodata starts at /var/run/ko.
	chroot := kodataRoot
	dirs := []string{
		"/var",
		"/var/run",
		"/var/run/ko",
	}
	if platform.OS == "windows" {
		chroot = "Files" + kodataRoot
		dirs = []string{
			"Hives",
			"Files",
			"Files/var",
			"Files/var/run",
			"Files/var/run/ko",
		}
	}
	for _, dir := range dirs {
		if err := tw.WriteHeader(&tar.Header{
			Name:     dir,
			Typeflag: tar.TypeDir,
			// Use a fixed Mode, so that this isn't sensitive to the directory and umask
			// under which it was created. Additionally, windows can only set 0222,
			// 0444, or 0666, none of which are executable.
			Mode:    0555,
			ModTime: creationTime.Time,
		}); err != nil {
			return nil, fmt.Errorf("writing dir %q: %w", dir, err)
		}
	}

	return buf, walkRecursive(tw, root, chroot, creationTime, platform)
}
