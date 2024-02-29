package storage

import (
	"errors"
)

var (
	ErrNotFound = errors.New("object not found")
)

type Store interface {
	// Get fetches data for the given Asset from the Store. If not
	// found it should return ErrNotFound, other errors are
	// considered internal server errors.
	Get(*Asset) error
	// Put saves the given Asset in the Store. Errors returned
	// should be considered as internal server errors.
	Put(Asset) error
	// Versions returns a list of versions available in the Store
	// for the given Asset.
	Versions(Asset) ([]string, error)
}
