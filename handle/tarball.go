package handle

import (
	"enpeeem/config"
	"enpeeem/storage"
	"errors"
	"log/slog"
	"net/http"
)

func Tarball(w http.ResponseWriter, r *http.Request) {
	cfg := config.FromContext(r)
	pkg, err := storage.NewPackage(cfg.Registry, r.PathValue("scope"), r.PathValue("pkg"))
	if err != nil {
		logErr(w, r, http.StatusInternalServerError, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	tarball := storage.NewTarball(pkg, r.PathValue("tarball"))
	data, err := cfg.Store.GetTarball(tarball)
	if errors.Is(err, storage.ErrNotFound) {
		if !cfg.ProxyStash {
			logErr(w, r, http.StatusNotFound, nil)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		data, err = tarball.FetchRemotely()
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				logErr(w, r, http.StatusNotFound, nil)
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
			logErr(w, r, http.StatusInternalServerError, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if err := cfg.Store.PutTarball(tarball, data); err != nil {
			logErr(w, r, http.StatusInternalServerError, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		if _, err := cfg.Store.Index(pkg); err != nil {
			logErr(w, r, http.StatusInternalServerError, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		logOK(r, "tarball fetched remotely")
	} else {
		logOK(r, "tarball found locally")
	}
	w.Write(data)
}

/*

===============================================================================

UTILS

===============================================================================

*/

func logOK(r *http.Request, msg string) {
	slog.Debug(msg, "method", r.Method, "url", r.URL, "http_status", http.StatusOK)
}

func logErr(w http.ResponseWriter, r *http.Request, status int, err error) {
	if err != nil {
		slog.Error("error occurred", "method", r.Method, "url", r.URL, "http_status", status, "cause", err)
	} else {
		slog.Error("failed request", "method", r.Method, "url", r.URL, "http_status", status)
	}
}
