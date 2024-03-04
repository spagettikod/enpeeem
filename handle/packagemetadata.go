package handle

import (
	"encoding/json"
	"enpeeem/config"
	"enpeeem/storage"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"text/template"
)

const (
	AbbreviatedPackageMetadataContentType = "application/vnd.npm.install-v1+json"
)

func PackageMetadata(w http.ResponseWriter, r *http.Request) (int, error) {
	cfg := config.FromContext(r)
	s, p := splitPkg(r.PathValue("pkg"))
	pkg, err := storage.NewPackage(cfg.Registry, s, p)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// don't use local storage when proxying, otherwise we won't be able to
	// fetch packages we don't have in the local storage
	if cfg.ProxyStash {
		remotePackageMetadata(w, r, cfg, pkg)
		return http.StatusOK, nil
	}

	data, err := localPackageMetadata(cfg.Store, pkg)
	if errors.Is(err, storage.ErrNotFound) {
		return http.StatusNotFound, err
	}
	if err != nil {
		return http.StatusInternalServerError, err
	}
	slog.Debug("metadata found locally", "method", r.Method, "url", r.URL, "http_status", http.StatusOK)
	w.Header().Add("Content-Type", AbbreviatedPackageMetadataContentType)

	if cfg.URLTemplate != nil {
		data, err = rewriteURLs(data, cfg.URLTemplate)
		if err != nil {
			return http.StatusInternalServerError, err
		}
	}
	w.Write(data)
	return http.StatusOK, nil
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

func rewriteURLs(data []byte, tmpl *template.Template) ([]byte, error) {
	pkmt := storage.PackageMetadata{}
	if err := json.Unmarshal(data, &pkmt); err != nil {
		return data, err
	}
	if err := pkmt.RewriteURLs(tmpl); err != nil {
		return data, err
	}
	return json.Marshal(pkmt)
}

func localPackageMetadata(store storage.Store, pkg storage.Package) ([]byte, error) {
	data, err := store.GetPackageMetadataRaw(pkg)
	if errors.Is(err, storage.ErrNotFound) {
		var tarballs []storage.Tarball
		tarballs, err = store.Tarballs(pkg)
		if err != nil {
			return []byte{}, err
		}
		slog.Info("metadata not found, will index it now", "pkg", pkg.String(), "tarballs", len(tarballs))
		if len(tarballs) == 0 {
			slog.Debug("no tarballs to index, exiting", "pkg", pkg.String())
			return []byte{}, storage.ErrNotFound
		}
		slog.Debug("indexing package", "pkg", pkg.String())
		if _, err := store.Index(pkg); err != nil {
			return []byte{}, err
		}
		slog.Debug("loading package", "pkg", pkg.String())
		data, err = store.GetPackageMetadataRaw(pkg)
	}
	return data, err
}

func remotePackageMetadata(w http.ResponseWriter, r *http.Request, cfg config.Config, pkg storage.Package) (int, error) {
	data, err := storage.FetchPackageMetadataRemotely(pkg)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return http.StatusNotFound, nil
		}
		return http.StatusInternalServerError, err
	}
	slog.Debug("metadata fetched remotely", "method", r.Method, "url", r.URL, "http_status", http.StatusOK)
	if cfg.FetchAll {
		go func() {
			if err := FetchAll(cfg, pkg, data); err != nil {
				slog.Error("error while fetching all tarballs", "cause", err)
			}
		}()
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(data)
	return http.StatusOK, nil
}
