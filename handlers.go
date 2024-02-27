package main

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"
)

func pkgHandler(w http.ResponseWriter, r *http.Request) {
	pkg := r.PathValue("pkg")
	file := path.Join(storageDir, pkg, "package.json")
	data, err := os.ReadFile(file)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if proxyNStash {
				url := fmt.Sprintf("%s/%s", registry, pkg)
				proxyStash(w, r, url, path.Join(storageDir, pkg), "package.json")
			} else {
				logErr(w, r, http.StatusNotFound, nil)
			}
		} else {
			logErr(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	logOK(r, file, "found locally")
	w.Write(data)
}

func tarballHandler(w http.ResponseWriter, r *http.Request) {
	pkg := r.PathValue("pkg")
	commonTarball(w, r, pkg)
}

func subpackageTarballHandler(w http.ResponseWriter, r *http.Request) {
	pkg := r.PathValue("pkg")
	subpkg := r.PathValue("subpkg")
	commonTarball(w, r, path.Join(pkg, subpkg))
}

func commonTarball(w http.ResponseWriter, r *http.Request, pkg string) {
	tarball := r.PathValue("tarball")
	file := path.Join(storageDir, pkg, tarball)
	data, err := os.ReadFile(file)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if proxyNStash {
				url := fmt.Sprintf("%s/%s/-/%s", registry, pkg, tarball)
				proxyStash(w, r, url, path.Join(storageDir, pkg), tarball)
			} else {
				logErr(w, r, http.StatusNotFound, nil)
			}
		} else {
			logErr(w, r, http.StatusInternalServerError, err)
		}
		return
	}
	logOK(r, file, "found locally")
	w.Write(data)
}

func logOK(r *http.Request, file string, msg string) {
	logger.Info(msg, "method", r.Method, "url", r.URL, "http_status", http.StatusOK, "file", file)
}

func logErr(w http.ResponseWriter, r *http.Request, status int, err error) {
	if err != nil {
		logger.Error("error occurred", "method", r.Method, "url", r.URL, "http_status", status, "cause", err)
	} else {
		logger.Error("failed request", "method", r.Method, "url", r.URL, "http_status", status)
	}
	http.Error(w, http.StatusText(status), status)
}
