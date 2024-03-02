package main

import (
	"encoding/json"
	"enpeeem/storage"
	"path"
	"slices"
)

func FetchAll(pkg storage.Package, packageMetadata []byte) error {
	type PackageMetadata struct {
		Versions map[string]struct {
			Dist struct {
				Tarball string
			}
		}
	}
	jsn := PackageMetadata{}
	err := json.Unmarshal(packageMetadata, &jsn)
	if err != nil {
		return err
	}
	existingTarballs, err := store.Tarballs(pkg)
	if err != nil {
		return err
	}
	for k := range jsn.Versions {
		file := path.Base(jsn.Versions[k].Dist.Tarball)
		tarball := storage.NewTarball(pkg, file)

		// skip if tarball exist
		if slices.Contains(existingTarballs, tarball) {
			continue
		}

		logger.Info("downloading tarball", "url", tarball.RemoteURL())
		if err := fetchAndSave(tarball); err != nil {
			logger.Error("failed to download tarball", "cause", err, "url", tarball.RemoteURL())
		}
	}
	return nil
	// return store.Index(pkg)
}

func fetchAndSave(tarball storage.Tarball) error {
	data, err := tarball.FetchRemotely()
	if err != nil {
		return err
	}
	return store.PutTarball(tarball, data)
}
