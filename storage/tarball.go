package storage

import (
	"net/url"
)

type Tarball struct {
	Asset
}

func NewTarball(registry, scope, pkg, name string) (Tarball, error) {
	u, err := url.Parse(registry)
	if err != nil {
		return Tarball{}, err
	}
	remoteURL, err := url.JoinPath(registry, scope, pkg, "-", name)
	if err != nil {
		return Tarball{}, err
	}
	return Tarball{
		Asset{
			remoteRegistry: registry,
			Registry:       u.Host,
			Scope:          scope,
			Package:        pkg,
			Name:           name,
			RemoteURL:      remoteURL,
		},
	}, nil
}
