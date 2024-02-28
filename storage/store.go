package storage

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type AssetType string

const (
	PackumentAssetType AssetType = "packument"
	TarballAssetType   AssetType = "tarball"
)

var (
	ErrNotFound = errors.New("object not found")
)

type Store interface {
	Get(*Asset) error
	Put(Asset) error
}

type Asset struct {
	remoteRegistry string
	Registry       string
	Package        string
	Name           string
	Data           []byte
	RemoteURL      string
	Type           AssetType
}

func NewPackument(registry, pkg string) (Asset, error) {
	u, err := url.Parse(registry)
	if err != nil {
		return Asset{}, err
	}
	return Asset{
		remoteRegistry: registry,
		Registry:       u.Host,
		Package:        pkg,
		Name:           "packument.json",
		RemoteURL:      fmt.Sprintf("%s/%s", registry, pkg),
		Type:           PackumentAssetType,
	}, nil
}

func NewTarball(registry, pkg, name string) (Asset, error) {
	u, err := url.Parse(registry)
	if err != nil {
		return Asset{}, err
	}
	return Asset{
		remoteRegistry: registry,
		Registry:       u.Host,
		Package:        pkg,
		Name:           name,
		RemoteURL:      fmt.Sprintf("%s/%s/-/%s", registry, pkg, name),
		Type:           TarballAssetType,
	}, nil
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
