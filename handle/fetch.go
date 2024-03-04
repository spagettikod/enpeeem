package handle

import (
	"encoding/json"
	"enpeeem/config"
	"enpeeem/storage"
	"log/slog"
	"path"
	"slices"
)

func FetchAll(cfg config.Config, pkg storage.Package, packageMetadata []byte) error {
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
	existingTarballs, err := cfg.Store.Tarballs(pkg)
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

		slog.Info("downloading tarball", "url", tarball.RemoteURL())
		if err := fetchAndSave(cfg, tarball); err != nil {
			slog.Error("failed to download tarball", "cause", err, "url", tarball.RemoteURL())
		}
	}
	return nil
}

func fetchAndSave(cfg config.Config, tarball storage.Tarball) error {
	data, err := tarball.FetchRemotely()
	if err != nil {
		return err
	}
	return cfg.Store.PutTarball(tarball, data)
}
