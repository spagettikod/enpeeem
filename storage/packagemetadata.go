package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"sort"

	"github.com/Masterminds/semver/v3"
)

const (
	PackageMetadataAssetName = "metadata.json"
)

type PackageMetadata struct {
	Asset
}

func NewPackageMetadata(registry, scope, pkg string) (PackageMetadata, error) {
	u, err := url.Parse(registry)
	if err != nil {
		return PackageMetadata{}, err
	}
	return PackageMetadata{
		Asset{
			remoteRegistry: registry,
			Registry:       u.Host,
			Scope:          scope,
			Package:        pkg,
			Name:           PackageMetadataAssetName,
			RemoteURL:      fmt.Sprintf("%s/%s", registry, pkg),
		},
	}, nil
}

func (pm *PackageMetadata) ReduceVersions(versions []string) error {
	meta := map[string]interface{}{}
	if err := json.Unmarshal(pm.Data, &meta); err != nil {
		return err
	}
	metaDistTags, ok := meta["dist-tags"].(map[string]interface{})
	if !ok {
		return errors.New("versions error")
	}
	metaVersions, ok := meta["versions"].(map[string]interface{})
	if !ok {
		return errors.New("versions error")
	}
	var err error

	versionsToKeep := map[string]interface{}{}
	for version, data := range metaVersions {
		if slices.Contains(versions, version) {
			versionsToKeep[version] = data
		}
	}
	meta["versions"] = versionsToKeep

	// npm expected the latest version to match available versions
	metaDistTags["latest"] = LatestStableVersion(versions)
	meta["dist-tags"] = metaDistTags

	pm.Data, err = json.MarshalIndent(meta, "", "   ")
	if err != nil {
		return err
	}
	return nil
}

// LatestStableVersion returns the latest stable version from an array of semver versions.
func LatestStableVersion(versions []string) string {
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
