package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"sort"

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

func ParsePackageJson(tarball Tarball, data []byte) (string, map[string]interface{}, error) {
	raw := map[string]interface{}{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return "", raw, err
	}
	raw["dist"] = map[string]string{"tarball": tarball.RemoteURL()}
	version, _ := raw["version"].(string)
	return version, raw, nil
}

// LatestStableVersion returns the latest stable version from an array of semver versions.
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
