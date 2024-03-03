package storage

import (
	"slices"
	"testing"
)

func TestLatestStableVersion(t *testing.T) {
	type Test struct {
		Values   []string
		Expected string
	}
	tests := []Test{
		{
			Values:   []string{"", "5.0.0", "2.4.5", "5.0.0-alpha", "6.0.0-beta.1"},
			Expected: "5.0.0",
		},
		{
			Values:   []string{},
			Expected: "",
		},
		{
			Values:   []string{"a", "b", "c"},
			Expected: "",
		},
	}

	for _, test := range tests {
		actual := latestStableVersion(test.Values)
		if actual != test.Expected {
			t.Errorf("expected %s but got %s", test.Expected, actual)
		}
	}

}

func TestPruneVersions(t *testing.T) {
	pkg := Package{Registry: "registry.npmjs.org", Scope: "", Name: "create-vite"}
	type Test struct {
		PackageMetadata PackageMetadata
		Tarballs        []Tarball
		Expected        []string
	}

	tests := []Test{
		{
			PackageMetadata: NewPackageMetadata("3.0.0", "create-vite",
				map[string]interface{}{
					"1.0.0": "",
					"2.0.0": "",
					"3.0.0": "",
				},
			),
			Tarballs: []Tarball{NewTarball(pkg, "create-vite-1.0.0.tgz"), NewTarball(pkg, "create-vite-3.0.0.tgz")},
			Expected: []string{"1.0.0", "3.0.0"},
		},
	}

	for _, test := range tests {
		test.PackageMetadata.PruneVersions(test.Tarballs)
		vl := test.PackageMetadata.VersionList()
		if len(vl) != len(test.Expected) {
			t.Errorf("expected %v version but found %v", len(test.Expected), len(vl))
		}
		for _, v := range test.Expected {
			if !slices.Contains(vl, v) {
				t.Errorf("expected to find version %s in PackageMeta data but did not", v)
			}
		}
	}
}
