package main

import (
	"enpeeem/storage"
	"errors"
	"net/http"
)

func packageMetadataHandler(w http.ResponseWriter, r *http.Request) {
	pkmt, err := storage.NewPackageMetadata(registry, r.PathValue("scope"), r.PathValue("pkg"))
	if err != nil {
		logErr(w, r, http.StatusInternalServerError, err)
		return
	}

	assetHandler(w, r, &pkmt.Asset)

	if !proxystash {
		versions, err := store.Versions(pkmt.Asset)
		if err != nil {
			logErr(w, r, http.StatusInternalServerError, err)
			return
		}
		pkmt.ReduceVersions(versions)
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(pkmt.Data)
}

func tarballHandler(w http.ResponseWriter, r *http.Request) {
	tarball, err := storage.NewTarball(registry, r.PathValue("scope"), r.PathValue("pkg"), r.PathValue("tarball"))
	if err != nil {
		logErr(w, r, http.StatusInternalServerError, err)
		return
	}
	assetHandler(w, r, &tarball.Asset)
	w.Write(tarball.Data)
}

// func scopedTarballHandler(w http.ResponseWriter, r *http.Request) {
// 	tarball, err := storage.NewTarball(registry, r.PathValue("scope"), r.PathValue("pkg"), r.PathValue("tarball"))
// 	if err != nil {
// 		logErr(w, r, http.StatusInternalServerError, err)
// 		return
// 	}
// 	assetHandler(w, r, &tarball.Asset)
// 	w.Write(tarball.Data)
// }

func assetHandler(w http.ResponseWriter, r *http.Request, asset *storage.Asset) {
	if err := store.Get(asset); err != nil {
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
			if err := store.Put(*asset); err != nil {
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
