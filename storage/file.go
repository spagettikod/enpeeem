package storage

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

const (
	PackageMetadataAssetName = "metadata.json"
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

func (fstore FileStore) assetFilename(registry, scope, pkg, name string) string {
	return path.Join(fstore.assetDir(registry, scope, pkg), name)
}
func (fstore FileStore) PutPackage(pkg Package, data []byte) error {
	dir := fstore.assetDir(pkg.Registry, pkg.Scope, pkg.Name)
	file := fstore.assetFilename(pkg.Registry, pkg.Scope, pkg.Name, PackageMetadataAssetName)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}
	return os.WriteFile(file, data, 0644)
}

func (fstore FileStore) PutTarball(tarball Tarball, data []byte) error {
	dir := fstore.assetDir(tarball.Package().Registry, tarball.Package().Scope, tarball.Package().Name)
	file := fstore.assetFilename(tarball.Package().Registry, tarball.Package().Scope, tarball.Package().Name, tarball.Name)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}
	if err := os.WriteFile(file, data, 0644); err != nil {
		return err
	}
	return fstore.Index(tarball.Package())
}

// func (fstore FileStore) Get(asset Asset) ([]byte, error) {
// 	data, err := os.ReadFile(fstore.filename(asset))
// 	if err != nil {
// 		if errors.Is(err, fs.ErrNotExist) {
// 			return data, ErrNotFound
// 		}
// 	}
// 	return data, err
// }

func (fstore FileStore) GetPackageMetadata(pkg Package) ([]byte, error) {
	filename := fstore.assetFilename(pkg.Registry, pkg.Scope, pkg.Name, PackageMetadataAssetName)
	data, err := os.ReadFile(filename)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return data, ErrNotFound
		}
	}
	return data, err
}

func (fstore FileStore) GetTarball(tarball Tarball) ([]byte, error) {
	data, err := os.ReadFile(fstore.assetFilename(tarball.Package().Registry, tarball.Package().Scope, tarball.Package().Name, tarball.Name))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return data, ErrNotFound
		}
	}
	return data, err
}

func (fstore FileStore) Packages() ([]Package, error) {
	pkgs := []Package{}
	root := os.DirFS(fstore.dir)
	files, err := fs.Glob(root, "*/*")
	if err != nil {
		return pkgs, err
	}
	// remove scoped packages that were globbed
	unscopedFiles := slices.DeleteFunc(files, func(f string) bool {
		if f == "" {
			return true
		}
		return filepath.Base(f)[0:1] == "@"
	})
	// scoped packages
	files, err = fs.Glob(root, "*/@*/*")
	if err != nil {
		return pkgs, err
	}

	files = append(files, unscopedFiles...)

	for _, file := range files {
		pkg, err := PackageMetadataFromURI(file)
		if err != nil {
			return pkgs, err
		}
		pkgs = append(pkgs, pkg)
	}
	return pkgs, nil
}

func (fstore FileStore) Tarballs(pkg Package) ([]Tarball, error) {
	tarballs := []Tarball{}
	root := os.DirFS(fstore.assetDir(pkg.Registry, pkg.Scope, pkg.Name))
	files, err := fs.Glob(root, "*.tgz")
	if err != nil {
		return tarballs, err
	}
	for _, file := range files {
		tarballs = append(tarballs, NewTarball(pkg, file))
	}

	return tarballs, nil
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

func (fstore FileStore) Index(pkg Package) error {
	tarballs, err := fstore.Tarballs(pkg)
	if err != nil {
		return err
	}

	verNos := []string{}
	versions := map[string]interface{}{}
	for _, tarball := range tarballs {
		data, err := fstore.GetTarball(tarball)
		if err != nil {
			return err
		}
		pkgJson, err := tarball.PackageJsonFromTar(data)
		if err != nil {
			return err
		}
		verNo, version, err := ParsePackageJson(tarball, pkgJson)
		if err != nil {
			return err
		}
		verNos = append(verNos, verNo)
		versions[verNo] = version
	}
	pm := NewPackageMetadata(latestStableVersion(verNos), pkg.Name, versions)
	jb, err := json.MarshalIndent(pm, "", "   ")
	if err != nil {
		return err
	}
	tmpfile, err := os.CreateTemp(os.TempDir(), "enpeeem_*.json")
	if err != nil {
		return err
	}
	if _, err := tmpfile.Write(jb); err != nil {
		tmpfile.Close()
		return err
	}
	if err := tmpfile.Close(); err != nil {
		return err
	}

	return os.Rename(tmpfile.Name(), fstore.assetFilename(pkg.Registry, pkg.Scope, pkg.Name, PackageMetadataAssetName))
}
