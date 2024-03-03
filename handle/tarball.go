package handle

import (
	"enpeeem/config"
	"enpeeem/storage"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

func Tarball(w http.ResponseWriter, r *http.Request) {
	cfg := config.FromContext(r)
	pkg, err := storage.NewPackage(cfg.Registry, r.PathValue("scope"), r.PathValue("pkg"))
	if err != nil {
		logErr(w, r, http.StatusInternalServerError, err)
		return
	}
	tarball := storage.NewTarball(pkg, r.PathValue("tarball"))
	data, err := cfg.Store.GetTarball(tarball)
	if errors.Is(err, storage.ErrNotFound) {
		if !cfg.ProxyStash {
			logErr(w, r, http.StatusNotFound, nil)
			return
		}
		data, err = tarball.FetchRemotely()
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				logErr(w, r, http.StatusNotFound, nil)
				return
			}
			logErr(w, r, http.StatusInternalServerError, err)
			return
		}
		if err := cfg.Store.PutTarball(tarball, data); err != nil {
			logErr(w, r, http.StatusInternalServerError, err)
			return
		}
		if _, err := cfg.Store.Index(pkg); err != nil {
			logErr(w, r, http.StatusInternalServerError, err)
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
	http.Error(w, http.StatusText(status), status)
}

func printHeaders(r *http.Request) {
	for k, v := range r.Header {
		fmt.Printf("%s: %s\n", k, v)
	}
}
