// Package repository provides implements parsing of apk repositories
package repository

import (
	"fmt"
	"strings"
)

type Repository struct {
	Uri string
}

// NewRepositoryFromComponents creates a new Repository with the uri constructed
// from the individual components
func NewRepositoryFromComponents(baseUri, release, repo, arch string) Repository {
	return Repository{
		Uri: fmt.Sprintf("%s/%s/%s/%s", baseUri, release, repo, arch),
	}
}

// WithIndex returns a RepositoryWithIndex object with the
func (r *Repository) WithIndex(index *ApkIndex) *RepositoryWithIndex {
	return &RepositoryWithIndex{
		Repository: r,
		index:      index,
	}
}

// IndexUri returns the uri of the APKINDEX for this repository
func (r *Repository) IndexUri() string {
	return fmt.Sprintf("%s/APKINDEX.tar.gz", r.Uri)
}

// IsRemote returns whether the repository is considered remote and needs to be
// fetched over http(s)
func (r *Repository) IsRemote() bool {
	return !strings.HasPrefix(r.Uri, "/")
}

// RepositoryWithIndex represents a repository with the index read and parsed
type RepositoryWithIndex struct {
	*Repository
	index *ApkIndex
}

// Packages returns a list of RepositoryPackage in this repository
func (r *RepositoryWithIndex) Packages() (pkgs []*RepositoryPackage) {
	for _, pkg := range r.index.Packages {
		rp := &RepositoryPackage{
			Package:    pkg,
			repository: r,
		}
		pkgs = append(pkgs, rp)
	}

	return
}

// Count returns the amout of packages that are available in this repository
func (r *RepositoryWithIndex) Count() int {
	return len(r.index.Packages)
}

// RepoAbbr returns a short name of this repository consiting of the repo name
// and the architecture.
func (r *RepositoryWithIndex) RepoAbbr() string {
	parts := strings.Split(r.Uri, "/")
	return strings.Join(parts[len(parts)-2:], "/")
}

type RepositoryPackage struct {
	*Package
	repository *RepositoryWithIndex
}

func NewRepositoryPackage(pkg *Package, repo *RepositoryWithIndex) *RepositoryPackage {
	return &RepositoryPackage{
		Package:    pkg,
		repository: repo,
	}
}

func (rp *RepositoryPackage) Url() string {
	return fmt.Sprintf("%s/%s", rp.repository.Uri, rp.Filename())
}

func (rp *RepositoryPackage) Repository() *RepositoryWithIndex {
	return rp.repository
}
