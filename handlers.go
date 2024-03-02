package main

import (
	"encoding/json"
	"enpeeem/storage"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
)

const (
	AbbreviatedPackageMetadataContentType = "application/vnd.npm.install-v1+json"
)

func fetchOrIndexLocally(pkg storage.Package) ([]byte, error) {
	data, err := store.GetPackageMetadata(pkg)
	if errors.Is(err, storage.ErrNotFound) {
		tarballs, err := store.Tarballs(pkg)
		if err != nil {
			return []byte{}, err
		}
		slog.Info("metadata not found, will index it now", "pkg", pkg.String(), "tarballs", len(tarballs))
		if len(tarballs) > 0 {
			slog.Debug("indexing package", "pkg", pkg.String())
			if pm, err := store.Index(pkg); err != nil {
				// even if indexing is incomplete we send what we have
				if errors.Is(err, storage.ErrIndexIncomplete) {
					slog.Debug("indexing was incomplete, sending what we have", "pkg", pkg.String())
					return json.MarshalIndent(pm, "", "   ")
				}
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

func packageMetadataHandler(w http.ResponseWriter, r *http.Request) {
	s, p := splitPkg(r.PathValue("pkg"))
	pkg, err := storage.NewPackage(registry, s, p)
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
		if fetchAll {
			go func() {
				if err := FetchAll(pkg, data); err != nil {
					slog.Error("error while fetching all tarballs", "cause", err)
				}
			}()
		}
		w.Header().Add("Content-Type", "application/json")
		w.Write(data)
		return
	}

	data, err := fetchOrIndexLocally(pkg)
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
		if _, err := store.Index(pkg); err != nil {
			// ignore if indexing failed, errors have been logged
			if !errors.Is(err, storage.ErrIndexIncomplete) {
				logErr(w, r, http.StatusInternalServerError, err)
				return
			}
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
