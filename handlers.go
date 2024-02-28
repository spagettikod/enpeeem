package main

import (
	"enpeeem/storage"
	"errors"
	"net/http"
	"path"
)

func pkgHandler(w http.ResponseWriter, r *http.Request) {
	asset, err := storage.NewPackument(registry, r.PathValue("pkg"))
	if err != nil {
		logErr(w, r, http.StatusInternalServerError, err)
		return
	}
	assetHandler(w, r, asset)
}

func tarballHandler(w http.ResponseWriter, r *http.Request) {
	tarball, err := storage.NewTarball(registry, r.PathValue("pkg"), r.PathValue("tarball"))
	if err != nil {
		logErr(w, r, http.StatusInternalServerError, err)
		return
	}
	assetHandler(w, r, tarball)
}

func subpackageTarballHandler(w http.ResponseWriter, r *http.Request) {
	jointPkg := path.Join(r.PathValue("pkg"), r.PathValue("subpkg"))
	tarball, err := storage.NewTarball(registry, jointPkg, r.PathValue("tarball"))
	if err != nil {
		logErr(w, r, http.StatusInternalServerError, err)
		return
	}
	assetHandler(w, r, tarball)
}

func assetHandler(w http.ResponseWriter, r *http.Request, asset storage.Asset) {
	if err := store.Get(&asset); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			if !proxystash {
				logErr(w, r, http.StatusNotFound, nil)
				return
			}
			if err := asset.FetchRemotely(); err != nil {
				if errors.Is(err, storage.ErrNotFound) {
					logErr(w, r, http.StatusNotFound, nil)
					return
				}
				logErr(w, r, http.StatusInternalServerError, err)
				return
			}
			if err := store.Put(asset); err != nil {
				logErr(w, r, http.StatusInternalServerError, err)
				return
			}
			logOK(r, "fetched remotely")
		} else {
			logErr(w, r, http.StatusInternalServerError, err)
			return
		}
	} else {
		logOK(r, "found locally")
	}
	w.Write(asset.Data)
}

func logOK(r *http.Request, msg string) {
	logger.Info(msg, "method", r.Method, "url", r.URL, "http_status", http.StatusOK)
}

func logErr(w http.ResponseWriter, r *http.Request, status int, err error) {
	if err != nil {
		logger.Error("error occurred", "method", r.Method, "url", r.URL, "http_status", status, "cause", err)
	} else {
		logger.Error("failed request", "method", r.Method, "url", r.URL, "http_status", status)
	}
	http.Error(w, http.StatusText(status), status)
}
