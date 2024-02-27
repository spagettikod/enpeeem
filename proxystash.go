package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
)

// proxyStash will request the given url from the remote registry. The result from
// the request is stored locally at the folder named in dir as filename.
func proxyStash(w http.ResponseWriter, r *http.Request, url, dir, filename string) {
	resp, err := http.Get(url)
	if err != nil {
		logErr(w, r, http.StatusInternalServerError, err)
		return
	}

	log.Printf("REMOTE %s %s %v", resp.Request.Method, url, resp.StatusCode)
	switch resp.StatusCode {
	case http.StatusNotFound:
		logErr(w, r, http.StatusNotFound, nil)
	case http.StatusOK:
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			logErr(w, r, http.StatusInternalServerError, err)
			return
		}
		if err := os.MkdirAll(dir, 0750); err != nil {
			logErr(w, r, http.StatusInternalServerError, err)
			return
		}
		file := path.Join(dir, filename)
		if err := os.WriteFile(file, data, 0644); err != nil {
			logErr(w, r, http.StatusInternalServerError, err)
			return
		}
		logOK(r, file)
		w.Write(data)
	default:
		logErr(w, r, http.StatusInternalServerError, fmt.Errorf("error calling %s responded with: %v %s", url, resp.StatusCode, resp.Status))
	}
}
