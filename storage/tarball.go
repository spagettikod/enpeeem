package storage

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

var ErrParseTarball = errors.New("parse error")

type Tarball struct {
	pkg  Package
	Name string
}

func NewTarball(pkg Package, name string) Tarball {
	return Tarball{
		pkg:  pkg,
		Name: name,
	}
}

// TarballFromURI parses an uri and returns a Tarball object without data. For example registry.npmjs.org/@babel/parser/parser-7.24.0.tgz.
func TarballFromURI(uri string) (Tarball, error) {
	els := strings.Split(uri, string(filepath.Separator))
	if len(els) < 3 || len(els) > 4 {
		return Tarball{}, fmt.Errorf("%w: could not parse path %s as Tarball", ErrParseTarball, uri)
	}
	var registry, scope, pkg, name string
	for i, el := range els {
		switch i {
		case 0:
			registry = el
		case 1:
			if len(els) == 3 {
				pkg = el
				continue
			}
			if len(els) == 4 {
				scope = el
			}
		case 2:
			if len(els) == 3 {
				name = el
				continue
			}
			if len(els) == 4 {
				pkg = el
			}
		case 3:
			name = el
		}
	}
	npkg, err := NewPackage(registry, scope, pkg)
	if err != nil {
		return Tarball{}, err
	}
	return NewTarball(npkg, name), nil
}

func (tarball Tarball) FetchRemotely() ([]byte, error) {
	resp, err := http.Get(tarball.RemoteURL())
	if err != nil {
		return []byte{}, err
	}

	switch resp.StatusCode {
	case http.StatusNotFound:
		return []byte{}, ErrNotFound
	case http.StatusOK:
		defer resp.Body.Close()
		return io.ReadAll(resp.Body)
	default:
		return []byte{}, fmt.Errorf("error calling %s responded with: %v %s", tarball.RemoteURL(), resp.StatusCode, resp.Status)
	}
}

func (tarball Tarball) Package() Package {
	return tarball.pkg
}

func (tarball Tarball) RemoteURL() string {
	remoteURL, _ := url.JoinPath("https://", tarball.pkg.Registry, tarball.pkg.Scope, tarball.pkg.Name, "-", tarball.Name)
	return remoteURL
}

func (tarball Tarball) String() string {
	return fmt.Sprintf("%s/%s", tarball.pkg.String(), tarball.Name)
}

func (tarball Tarball) PackageJsonFromTar(tgz []byte) ([]byte, error) {
	if len(tgz) == 0 {
		return []byte{}, fmt.Errorf("can not extract package/package.json, from %s, empty data", tarball.String())
	}
	buffer := bytes.NewBuffer(tgz)
	gzipReader, err := gzip.NewReader(buffer)
	if err != nil {
		return []byte{}, fmt.Errorf("error in gzip reader opening %s: %w", tarball.pkg.String(), err)
	}

	tr := tar.NewReader(gzipReader)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return []byte{}, fmt.Errorf("could not find package.json in %s", tarball.pkg.String())
		}
		if err != nil {
			return []byte{}, err
		}
		buf := bytes.NewBuffer([]byte{})
		matched, err := filepath.Match("*/package.json", hdr.Name)
		if err != nil {
			return []byte{}, fmt.Errorf("could not match */package.json in %s: %w", tarball.pkg.String(), err)
		}
		if matched {
			if _, err := io.Copy(buf, tr); err != nil {
				return []byte{}, err
			}
			return buf.Bytes(), nil
		}
	}
}
