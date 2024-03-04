package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
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
	dir     string
	metadir string
}

func NewFileStore(dir, metadir string) *FileStore {
	return &FileStore{dir: dir, metadir: metadir}
}

func (fstore FileStore) tarballDir(pkg Package) string {
	return path.Join(fstore.dir, pkg.Registry, pkg.Scope, pkg.Name)
}

func (fstore FileStore) tarballFilename(tarball Tarball) string {
	return path.Join(fstore.tarballDir(tarball.Package()), tarball.Name)
}

func (fstore FileStore) packageDir(pkg Package) string {
	return path.Join(fstore.metadir, pkg.Registry, pkg.Scope, pkg.Name)
}

func (fstore FileStore) packageFilename(pkg Package) string {
	return path.Join(fstore.packageDir(pkg), PackageMetadataAssetName)
}

func (fstore FileStore) PutPackage(pkg Package, data []byte) error {
	dir := fstore.packageDir(pkg)
	file := fstore.packageFilename(pkg)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}
	return os.WriteFile(file, data, 0644)
}

func (fstore FileStore) PutTarball(tarball Tarball, data []byte) error {
	dir := fstore.tarballDir(tarball.Package())
	file := fstore.tarballFilename(tarball)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}
	return os.WriteFile(file, data, 0644)
}

func (fstore FileStore) GetPackageMetadataRaw(pkg Package) ([]byte, error) {
	filename := fstore.packageFilename(pkg)
	data, err := os.ReadFile(filename)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return data, ErrNotFound
		}
	}
	return data, err
}

func (fstore FileStore) GetPackageMetadata(pkg Package) (PackageMetadata, error) {
	pkmt := PackageMetadata{}
	raw, err := fstore.GetPackageMetadataRaw(pkg)
	if err != nil {
		return pkmt, err
	}
	err = json.Unmarshal(raw, &pkmt)
	return pkmt, err
}

func (fstore FileStore) GetTarball(tarball Tarball) ([]byte, error) {
	data, err := os.ReadFile(fstore.tarballFilename(tarball))
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
	root := os.DirFS(fstore.tarballDir(pkg))
	files, err := fs.Glob(root, "*.tgz")
	if err != nil {
		return tarballs, err
	}
	for _, file := range files {
		tarballs = append(tarballs, NewTarball(pkg, file))
	}

	return tarballs, nil
}

func (fstore FileStore) Index(pkg Package) (PackageMetadata, error) {
	tarballs, err := fstore.Tarballs(pkg)
	if err != nil {
		return PackageMetadata{}, err
	}

	slog.Debug("number of tarballs found", "tarballs", len(tarballs), "pkg", pkg.String())
	pkmt, err := fstore.GetPackageMetadata(pkg)
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			return pkmt, fmt.Errorf("error opening existing package metadata file for %s: %w", pkg.String(), err)
		}
		slog.Debug("creating new package metadata, existing not found", "pkg", pkg.String())
		pkmt = NewPackageMetadata("", pkg.Name, map[string]interface{}{})
	} else {
		// remove metadata for tarballs that no longer exist on disk
		pkmt.PruneVersions(tarballs)
	}

	// we don't need to process tarballs already indexed in the package metadata file
	tarballs = slices.DeleteFunc(tarballs, func(tarball Tarball) bool {
		v := fileVersion(pkg.Name, tarball.Name)
		_, found := pkmt.Versions[v]
		return found
	})

	slog.Debug("unindexed tarballs", "tarballs", len(tarballs), "pkg", pkg.String())

	for _, tarball := range tarballs {
		slog.Debug("loading tarball", "tarball", tarball.String(), "pkg", pkg.String())
		data, err := fstore.GetTarball(tarball)
		if err != nil {
			slog.Error("could not load tarball, skipping", "tarball", tarball.String(), "error", err)
			continue
		}
		if err := pkmt.AddVersion(tarball, data); err != nil {
			slog.Error("error parsing tarball, skipping", "tarball", tarball.String(), "error", err)
			continue
		}
	}
	jb, err := json.MarshalIndent(pkmt, "", "   ")
	if err != nil {
		return pkmt, err
	}
	slog.Debug("create directory if not exists", "pkg", pkg.String(), "dir", fstore.packageDir(pkg))
	if err := os.MkdirAll(fstore.packageDir(pkg), 0750); err != nil {
		return pkmt, err
	}
	slog.Debug("writing new package metadata file", "pkg", pkg.String(), "file", fstore.packageFilename(pkg))
	return pkmt, os.WriteFile(fstore.packageFilename(pkg), jb, 0644)
}

func fileVersion(pkgName, filename string) string {
	if filename == "" || pkgName == "" {
		return ""
	}
	begin := len(pkgName) + 1
	end := strings.LastIndex(filename, ".tgz")
	if begin < 0 || end < 0 {
		return ""
	}
	ver := filename[begin:end]
	return ver
}
