package storage

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

type Package struct {
	Registry string
	Scope    string
	Name     string
}

func NewPackage(registry, scope, name string) (Package, error) {
	// parse scheme and only keep host if registry is a URL
	if len(registry) > 4 && registry[:4] == "http" {
		u, err := url.Parse(registry)
		if err != nil {
			return Package{}, err
		}
		registry = u.Host
	}
	return Package{
		Registry: registry,
		Scope:    scope,
		Name:     name,
	}, nil
}

func (pkg Package) String() string {
	if pkg.Scope == "" {
		return fmt.Sprintf("%s/%s", pkg.Registry, pkg.Name)
	} else {
		return fmt.Sprintf("%s/%s/%s", pkg.Registry, pkg.Scope, pkg.Name)
	}
}

// PackageMetadataFromURI parses an uri and returns a Tarball object without data. For example registry.npmjs.org/@babel/parser/parser-7.24.0.tgz.
func PackageMetadataFromURI(uri string) (Package, error) {
	pkmt := Package{}
	els := strings.Split(uri, string(filepath.Separator))
	if len(els) < 2 || len(els) > 3 {
		return pkmt, fmt.Errorf("%w: could not parse path %s as PackageMetadata", ErrParseTarball, uri)
	}
	var registry, scope, pkg string
	for i, el := range els {
		switch i {
		case 0:
			registry = el
		case 1:
			if len(els) == 2 {
				pkg = el
			}
			if len(els) == 3 {
				scope = el
			}
		case 2:
			pkg = el
		}
	}
	return NewPackage(registry, scope, pkg)
}

func (pkg Package) RemoteURL() string {
	remoteURL, _ := url.JoinPath("https://", pkg.Registry, pkg.Scope, pkg.Name)
	return remoteURL
}
