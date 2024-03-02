package main

import (
	"enpeeem/storage"
	"errors"
	"fmt"
	"net/http"
)

const (
	AbbreviatedPackageMetadataContentType = "application/vnd.npm.install-v1+json"
)

func packageMetadataHandler(w http.ResponseWriter, r *http.Request) {
	pkg, err := storage.NewPackage(registry, r.PathValue("scope"), r.PathValue("pkg"))
	if err != nil {
		logErr(w, r, http.StatusInternalServerError, err)
		return
	}

	// don't use local storage when proxying, otherwise we won't be able to
	// fetch packages we don't have in the local storage
	if proxystash {
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
		w.Header().Add("Content-Type", "application/json")
		w.Write(data)
		return
	}

	data, err := store.GetPackageMetadata(pkg)
	if errors.Is(err, storage.ErrNotFound) {
		logErr(w, r, http.StatusNotFound, nil)
		return
	}
	logOK(r, "metadata found locally")
	w.Header().Add("Content-Type", AbbreviatedPackageMetadataContentType)

	w.Write(data)
}

func tarballHandler(w http.ResponseWriter, r *http.Request) {
	pkg, err := storage.NewPackage(registry, r.PathValue("scope"), r.PathValue("pkg"))
	if err != nil {
		logErr(w, r, http.StatusInternalServerError, err)
		return
	}
	tarball := storage.NewTarball(pkg, r.PathValue("tarball"))
	data, err := store.GetTarball(tarball)
	if errors.Is(err, storage.ErrNotFound) {
		if !proxystash {
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
		if err := store.PutTarball(tarball, data); err != nil {
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

func printHeaders(r *http.Request) {
	for k, v := range r.Header {
		fmt.Printf("%s: %s\n", k, v)
	}
}
