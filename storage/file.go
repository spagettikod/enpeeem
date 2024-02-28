package storage

import (
	"errors"
	"io/fs"
	"os"
	"path"
)

type FileStore struct {
	dir string
}

func NewFileStore(dir string) FileStore {
	return FileStore{dir: dir}
}

func (fstore FileStore) assetDir(asset Asset) string {
	return path.Join(fstore.dir, asset.Registry, asset.Package)
}

func (fstore FileStore) filename(asset Asset) string {
	return path.Join(fstore.assetDir(asset), asset.Name)
}

func (fstore FileStore) Put(asset Asset) error {
	dir := fstore.assetDir(asset)
	file := fstore.filename(asset)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}
	return os.WriteFile(file, asset.Data, 0644)
}

func (fstore FileStore) Get(asset *Asset) error {
	var err error
	asset.Data, err = os.ReadFile(fstore.filename(*asset))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return ErrNotFound
		}
	}
	return err
}
