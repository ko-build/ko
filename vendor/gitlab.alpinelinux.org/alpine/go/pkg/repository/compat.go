// This package is for backwards compatibility. Use
// gitlab.alpinelinux.org/alpine/go/repository instead.
package repository

import "gitlab.alpinelinux.org/alpine/go/repository"

type (
	ApkIndex            = repository.ApkIndex
	Package             = repository.Package
	Repository          = repository.Repository
	RepositoryPackage   = repository.RepositoryPackage
	RepositoryWithIndex = repository.RepositoryWithIndex
)

var (
	IndexFromArchive            = repository.IndexFromArchive
	NewRepositoryFromComponents = repository.NewRepositoryFromComponents
	NewRepositoryPackage        = repository.NewRepositoryPackage
	ParsePackageIndex           = repository.ParsePackageIndex
)
