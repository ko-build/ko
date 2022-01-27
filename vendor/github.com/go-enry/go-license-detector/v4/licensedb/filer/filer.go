package filer

import (
	"archive/zip"
	"bytes"
	"io/ioutil"
	"os"
	xpath "path"
	"path/filepath"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/pkg/errors"
)

// File represents a file in the virtual file system: every node is either a regular file
// or a directory. Symlinks are dereferenced in the implementations.
type File struct {
	Name  string
	IsDir bool
}

// A Filer provides a list of files.
type Filer interface {
	// ReadFile returns the contents of a file given it's path.
	ReadFile(path string) (content []byte, err error)
	// ReadDir lists a directory.
	ReadDir(path string) ([]File, error)
	// Close frees all the resources allocated by this Filer.
	Close()
	// PathsAreAlwaysSlash indicates whether the path separator is platform-independent ("/") or
	// OS-specific.
	PathsAreAlwaysSlash() bool
}

type localFiler struct {
	root string
}

// FromDirectory returns a Filer that allows accessing over all the files contained in a directory.
func FromDirectory(path string) (Filer, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create Filer from %s", path)
	}
	if !fi.IsDir() {
		return nil, errors.New("not a directory")
	}
	path, _ = filepath.Abs(path)
	return &localFiler{root: path}, nil
}

func (filer *localFiler) resolvePath(path string) (string, error) {
	path, err := filepath.Abs(filepath.Join(filer.root, path))
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(path, filer.root) {
		return "", errors.New("path is out of scope")
	}
	return path, nil
}

func (filer *localFiler) ReadFile(path string) ([]byte, error) {
	path, err := filer.resolvePath(path)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read file %s", path)
	}
	buffer, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read file %s", path)
	}
	return buffer, nil
}

func (filer *localFiler) ReadDir(path string) ([]File, error) {
	path, err := filer.resolvePath(path)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read directory %s", path)
	}
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read directory %s", path)
	}
	result := make([]File, 0, len(files))
	for _, file := range files {
		result = append(result, File{
			Name:  file.Name(),
			IsDir: file.IsDir(),
		})
	}
	return result, nil
}

func (filer *localFiler) Close() {}

func (filer *localFiler) PathsAreAlwaysSlash() bool {
	return false
}

type gitFiler struct {
	root *object.Tree
}

// FromGitURL returns a Filer that allows to access all the files in a Git repository's default branch given its URL.
// It keeps a shallow single-branch clone of the repository in memory.
func FromGitURL(url string) (Filer, error) {
	repo, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{URL: url, Depth: 1})
	if err != nil {
		return nil, errors.Wrapf(err, "could not clone repo from %s", url)
	}
	return FromGit(repo, "")
}

// FromGit returns a Filer that allows accessing all the files in a Git repository
func FromGit(repo *git.Repository, headRef plumbing.ReferenceName) (Filer, error) {
	var head *plumbing.Reference
	var err error
	if headRef == "" {
		head, err = repo.Head()
	} else {
		head, err = repo.Reference(headRef, true)
	}
	if err != nil {
		return nil, errors.Wrap(err, "could not fetch HEAD from repo")
	}
	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return nil, errors.Wrap(err, "could not fetch commit for HEAD")
	}
	tree, err := commit.Tree()
	if err != nil {
		return nil, errors.Wrap(err, "could not fetch root for HEAD commit")
	}
	return &gitFiler{root: tree}, nil
}

func (filer gitFiler) ReadFile(path string) ([]byte, error) {
	entry, err := filer.root.FindEntry(path)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot find file %s", path)
	}
	if entry.Mode == filemode.Symlink {
		file, err := filer.root.File(path)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot find file %s", path)
		}
		path, err = file.Contents()
		if err != nil {
			return nil, errors.Wrapf(err, "cannot read file %s", path)
		}
	}
	file, err := filer.root.File(path)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read file %s", path)
	}
	reader, err := file.Reader()
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read file %s", path)
	}
	defer func() { err = reader.Close() }()

	buf := new(bytes.Buffer)
	if _, err = buf.ReadFrom(reader); err != nil {
		return nil, errors.Wrapf(err, "cannot read file %s", path)
	}
	return buf.Bytes(), err
}

func (filer *gitFiler) ReadDir(path string) ([]File, error) {
	var tree *object.Tree
	if path != "" {
		var err error
		tree, err = filer.root.Tree(path)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot read directory %s", path)
		}
	} else {
		tree = filer.root
	}
	result := make([]File, 0, len(tree.Entries))
	for _, entry := range tree.Entries {
		switch entry.Mode {
		case filemode.Dir:
			result = append(result, File{
				Name:  entry.Name,
				IsDir: true,
			})
		case filemode.Regular, filemode.Executable, filemode.Deprecated, filemode.Symlink:
			result = append(result, File{
				Name:  entry.Name,
				IsDir: false,
			})
		}
	}
	return result, nil
}

func (filer *gitFiler) Close() {
	filer.root = nil
}

func (filer *gitFiler) PathsAreAlwaysSlash() bool {
	return true
}

type zipNode struct {
	children map[string]*zipNode
	file     *zip.File
}

type zipFiler struct {
	arch *zip.ReadCloser
	tree *zipNode
}

// FromZIP returns a Filer that allows accessing all the files in a ZIP archive given its path.
func FromZIP(path string) (Filer, error) {
	arch, err := zip.OpenReader(path)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read ZIP archive %s", path)
	}
	root := &zipNode{children: map[string]*zipNode{}}
	for _, f := range arch.File {
		path := strings.Split(f.Name, "/") // zip always has "/"
		node := root
		for _, part := range path {
			if part == "" {
				continue
			}
			child := node.children[part]
			if child == nil {
				child = &zipNode{children: map[string]*zipNode{}}
				node.children[part] = child
			}
			node = child
		}
		node.file = f
	}
	return &zipFiler{arch: arch, tree: root}, nil
}

func (filer *zipFiler) ReadFile(path string) ([]byte, error) {
	parts := strings.Split(path, string("/"))
	node := filer.tree
	for _, part := range parts {
		if part == "" {
			continue
		}
		node = node.children[part]
		if node == nil {
			return nil, errors.Errorf("does not exist: %s", path)
		}
	}
	reader, err := node.file.Open()
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open %s", path)
	}
	defer reader.Close()
	buffer, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read %s", path)
	}
	return buffer, nil
}

func (filer *zipFiler) ReadDir(path string) ([]File, error) {
	parts := strings.Split(path, string("/"))
	node := filer.tree
	for _, part := range parts {
		if part == "" {
			continue
		}
		node = node.children[part]
		if node == nil {
			return nil, errors.Errorf("does not exist: %s", path)
		}
	}
	if path != "" && !node.file.FileInfo().IsDir() {
		return nil, errors.Errorf("not a directory: %s", path)
	}
	result := make([]File, 0, len(node.children))
	for name, child := range node.children {
		result = append(result, File{
			Name:  name,
			IsDir: child.file.FileInfo().IsDir(),
		})
	}
	return result, nil
}

func (filer *zipFiler) Close() {
	filer.arch.Close()
}

func (filer *zipFiler) PathsAreAlwaysSlash() bool {
	return true
}

type nestedFiler struct {
	origin Filer
	offset string
}

// NestFiler wraps an existing Filer. It prepends the specified prefix to every path.
func NestFiler(filer Filer, prefix string) Filer {
	return &nestedFiler{origin: filer, offset: prefix}
}

func (filer *nestedFiler) ReadFile(path string) ([]byte, error) {
	var fullPath string
	if filer.origin.PathsAreAlwaysSlash() {
		fullPath = xpath.Join(filer.offset, path)
	} else {
		fullPath = filepath.Join(filer.offset, path)
	}
	return filer.origin.ReadFile(fullPath)
}

func (filer *nestedFiler) ReadDir(path string) ([]File, error) {
	var fullPath string
	if filer.origin.PathsAreAlwaysSlash() {
		fullPath = xpath.Join(filer.offset, path)
	} else {
		fullPath = filepath.Join(filer.offset, path)
	}
	return filer.origin.ReadDir(fullPath)
}

func (filer *nestedFiler) Close() {
	filer.origin.Close()
}

func (filer *nestedFiler) PathsAreAlwaysSlash() bool {
	return filer.origin.PathsAreAlwaysSlash()
}
