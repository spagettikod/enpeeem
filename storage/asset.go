package storage

import (
	"fmt"
	"io"
	"net/http"
)

type Asset struct {
	remoteRegistry string
	Registry       string
	Scope          string
	Package        string
	Name           string
	Data           []byte
	RemoteURL      string
}

func (asset *Asset) FetchRemotely() error {
	resp, err := http.Get(asset.RemoteURL)
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusOK:
		defer resp.Body.Close()
		asset.Data, err = io.ReadAll(resp.Body)
		return err
	default:
		return fmt.Errorf("error calling %s responded with: %v %s", asset.RemoteURL, resp.StatusCode, resp.Status)
	}
}
