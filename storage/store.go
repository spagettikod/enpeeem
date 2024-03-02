package storage

import (
	"errors"
)

var (
	ErrNotFound = errors.New("object not found")
)

type Store interface {
	GetPackageMetadata(Package) ([]byte, error)
	PutPackage(Package, []byte) error
	PutTarball(Tarball, []byte) error
	Packages() ([]Package, error)
	Tarballs(Package) ([]Tarball, error)
	GetTarball(Tarball) ([]byte, error)
	Index(Package) error
}
