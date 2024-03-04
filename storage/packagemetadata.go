package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"sort"
	"text/template"

	"github.com/Masterminds/semver/v3"
)

type PackageMetadata struct {
	DistTags map[string]string      `json:"dist-tags"`
	Name     string                 `json:"name"`
	Versions map[string]interface{} `json:"versions"`
}

func NewPackageMetadata(latestVersion, packageName string, versions map[string]interface{}) PackageMetadata {
	return PackageMetadata{
		DistTags: map[string]string{"latest": latestVersion},
		Name:     packageName,
		Versions: versions,
	}
}

// PruneVersions takes a list of tarballs and validates their versions against versions found in the metadata.
// Metadata versions not found in the list of tarballs are removed from the metadata file. Package metadata
// latest version field is also updated to reflect the version list.
func (pm *PackageMetadata) PruneVersions(tarballs []Tarball) {
	pmVers := pm.VersionList()
	for _, v := range pmVers {
		tarballExists := slices.ContainsFunc(tarballs, func(t Tarball) bool {
			return t.Version() == v
		})
		if !tarballExists {
			delete(pm.Versions, v)
		}
	}
	pm.SetLatestVersion()
}

// VersionList returns an array with all version numbers in the metadata.
func (pm *PackageMetadata) VersionList() []string {
	versions := []string{}
	for k := range pm.Versions {
		versions = append(versions, k)
	}
	return versions
}

// SetLatestVersion upates the metadata field pointing to the latest
// stable version for this package.
func (pm *PackageMetadata) SetLatestVersion() {
	versions := pm.VersionList()
	pm.DistTags["latest"] = latestStableVersion(versions)
}

func (pm *PackageMetadata) RewriteURLs(tmpl *template.Template) error {
	newvers := map[string]interface{}{}
	fmt.Println(len(pm.Versions))
	for k, v := range pm.Versions {
		version := v.(map[string]interface{})
		dist := version["dist"].(map[string]interface{})
		tbl, err := TarballFromURI(dist["tarball"].(string))
		if err != nil {
			return err
		}
		nurl, err := tbl.RewrittenURL(tmpl)
		if err != nil {
			return err
		}
		v.(map[string]interface{})["dist"] = map[string]string{"tarball": nurl}
		newvers[k] = v
	}

	pm.Versions = newvers
	return nil
}

// AddVersion unpacks and parses package.json metadata from raw tarball bytes and adds it
// as a version to the package metadata. Package metadata latest version field is also
// updated to reflect the new version.
func (pm *PackageMetadata) AddVersion(tarball Tarball, data []byte) error {
	pkgJson, err := tarball.PackageJsonFromTar(data)
	if err != nil {
		return fmt.Errorf("could not fetch package.json from tarball: %w", err)
	}
	verNo, version, err := parsePackageJson(tarball, pkgJson)
	if err != nil {
		return fmt.Errorf("could not parse package.json: %w", err)
	}
	pm.Versions[verNo] = version
	pm.SetLatestVersion()
	return nil
}

func parsePackageJson(tarball Tarball, data []byte) (string, map[string]interface{}, error) {
	raw := map[string]interface{}{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return "", raw, err
	}
	raw["dist"] = map[string]string{"tarball": tarball.RemoteURL()}
	version, _ := raw["version"].(string)
	return version, raw, nil
}

// latestStableVersion returns the latest stable version from an array
// of semver versions.
func latestStableVersion(versions []string) string {
	if len(versions) == 0 {
		return ""
	}
	vs := []*semver.Version{}
	for _, v := range versions {
		v, err := semver.NewVersion(v)
		// simply skip adding versions we can not parse
		if err != nil {
			continue
		}
		vs = append(vs, v)
	}

	vs = slices.DeleteFunc(vs, func(v *semver.Version) bool {
		return v.Prerelease() != ""
	})
	sort.Sort(semver.Collection(vs))

	if len(vs) == 0 {
		return ""
	}

	return vs[len(vs)-1].String()
}

// FetchPackageMetadataRemotely downloads package metadata from remote registry,
func FetchPackageMetadataRemotely(pkg Package) ([]byte, error) {
	resp, err := http.Get(pkg.RemoteURL())
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
		return []byte{}, fmt.Errorf("error calling %s responded with: %v %s", pkg.RemoteURL(), resp.StatusCode, resp.Status)
	}
}
