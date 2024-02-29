package storage

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"strings"
)

type FileStore struct {
	dir string
}

func NewFileStore(dir string) FileStore {
	return FileStore{dir: dir}
}

func (fstore FileStore) assetDir(registry, scope, pkg string) string {
	return path.Join(fstore.dir, registry, scope, pkg)
}

func (fstore FileStore) filename(asset Asset) string {
	return path.Join(fstore.assetDir(asset.Registry, asset.Scope, asset.Package), asset.Name)
}

func (fstore FileStore) Put(asset Asset) error {
	dir := fstore.assetDir(asset.Registry, asset.Scope, asset.Package)
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

func (fstore FileStore) Versions(asset Asset) ([]string, error) {
	versions := []string{}
	dir := fstore.assetDir(asset.Registry, asset.Scope, asset.Package)
	fsys := os.DirFS(dir)
	files, err := fs.Glob(fsys, "*.tgz")
	if err != nil {
		return []string{}, nil
	}
	for _, file := range files {
		ver := fileVersion(asset.Package, file)
		if ver != "" {
			versions = append(versions, ver)
		}
	}
	return versions, nil
}

func fileVersion(pkg, filename string) string {
	if filename == "" {
		return ""
	}
	begin := len(pkg) + 1
	end := strings.LastIndex(filename, ".tgz")
	if begin < 0 || end < 0 {
		return ""
	}
	ver := filename[begin:end]
	return ver
}
