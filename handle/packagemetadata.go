package handle

import (
	"enpeeem/config"
	"enpeeem/storage"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

const (
	AbbreviatedPackageMetadataContentType = "application/vnd.npm.install-v1+json"
)

func fetchOrIndexLocally(store storage.Store, pkg storage.Package) ([]byte, error) {
	data, err := store.GetPackageMetadata(pkg)
	if errors.Is(err, storage.ErrNotFound) {
		tarballs, err := store.Tarballs(pkg)
		if err != nil {
			return []byte{}, err
		}
		slog.Info("metadata not found, will index it now", "pkg", pkg.String(), "tarballs", len(tarballs))
		if len(tarballs) > 0 {
			slog.Debug("indexing package", "pkg", pkg.String())
			if _, err := store.Index(pkg); err != nil {
				return []byte{}, err
			}
			slog.Debug("loading package", "pkg", pkg.String())
			return store.GetPackageMetadata(pkg)
		}
		return []byte{}, storage.ErrNotFound
	}
	return data, nil
}

func splitPkg(s string) (string, string) {
	if len(s) < 1 {
		return "", ""
	}
	if s[0:1] == "@" {
		split := strings.Split(s, "/")
		if len(split) != 2 {
			return "", ""
		}
		return split[0], split[1]
	}
	return "", s
}

func PackageMetadata(w http.ResponseWriter, r *http.Request) {
	cfg := config.FromContext(r)
	s, p := splitPkg(r.PathValue("pkg"))
	pkg, err := storage.NewPackage(cfg.Registry, s, p)
	if err != nil {
		logErr(w, r, http.StatusInternalServerError, err)
		return
	}

	// don't use local storage when proxying, otherwise we won't be able to
	// fetch packages we don't have in the local storage
	if cfg.ProxyStash {
		data, err := storage.FetchPackageMetadataRemotely(pkg)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				logErr(w, r, http.StatusNotFound, nil)
				return
			}
			logErr(w, r, http.StatusInternalServerError, err)
			return
		}
		logOK(r, "metadata fetched remotely")
		if cfg.FetchAll {
			go func() {
				if err := FetchAll(cfg.Store, pkg, data); err != nil {
					slog.Error("error while fetching all tarballs", "cause", err)
				}
			}()
		}
		w.Header().Add("Content-Type", "application/json")
		w.Write(data)
		return
	}

	data, err := fetchOrIndexLocally(cfg.Store, pkg)
	if errors.Is(err, storage.ErrNotFound) {
		logErr(w, r, http.StatusNotFound, nil)
		return
	}
	if err != nil {
		logErr(w, r, http.StatusNotFound, err)
		return
	}
	logOK(r, "metadata found locally")
	w.Header().Add("Content-Type", AbbreviatedPackageMetadataContentType)

	w.Write(data)
}
