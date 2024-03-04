package handle

import (
	"enpeeem/config"
	"enpeeem/storage"
	"errors"
	"log/slog"
	"net/http"
)

func Tarball(w http.ResponseWriter, r *http.Request) (int, error) {
	cfg := config.FromContext(r)
	pkg, err := storage.NewPackage(cfg.Registry, r.PathValue("scope"), r.PathValue("pkg"))
	if err != nil {
		return http.StatusInternalServerError, err
	}
	tarball := storage.NewTarball(pkg, r.PathValue("tarball"))
	data, err := cfg.Store.GetTarball(tarball)
	if errors.Is(err, storage.ErrNotFound) {
		if !cfg.ProxyStash {
			return http.StatusNotFound, nil
		}
		data, err = tarball.FetchRemotely()
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return http.StatusNotFound, nil
			}
			return http.StatusInternalServerError, err
		}
		if err := cfg.Store.PutTarball(tarball, data); err != nil {
			return http.StatusInternalServerError, err
		}
		if _, err := cfg.Store.Index(pkg); err != nil {
			return http.StatusInternalServerError, err
		}
		slog.Debug("tarball fetched remotely", "method", r.Method, "url", r.URL, "http_status", http.StatusOK)
	} else {
		slog.Debug("tarball found locally", "method", r.Method, "url", r.URL, "http_status", http.StatusOK)
	}
	w.Write(data)
	return http.StatusOK, nil
}
