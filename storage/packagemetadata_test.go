package storage

import (
	"encoding/json"
	"os"
	"testing"
)

func TestReduceVersions(t *testing.T) {
	data, err := os.ReadFile("testdata/create-vite_metadata.json")
	if err != nil {
		t.Fatal(err)
	}
	pkmt, err := NewPackageMetadata("https://registry.npmjs.org", "", "create-vite")
	if err != nil {
		t.Fatal(err)
	}
	pkmt.Data = data

	myVersions := []string{"5.2.0", "5.0.0-beta.1", "4.4.0", "4.3.0"}
	pkmt.ReduceVersions(myVersions)

	meta := map[string]interface{}{}
	if err := json.Unmarshal(pkmt.Data, &meta); err != nil {
		t.Fatal(err)
	}
	metaVersions, ok := meta["versions"].(map[string]interface{})
	if !ok {
		t.Fatal("error fetching versions")
	}

	keptVersions := []string{}
	for k := range metaVersions {
		keptVersions = append(keptVersions, k)
	}

	if len(keptVersions) != len(myVersions) {
		t.Fatal("number of versions in package metadata differs from expected")
	}
}

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
		actual := LatestStableVersion(test.Values)
		if actual != test.Expected {
			t.Errorf("expected %s but got %s", test.Expected, actual)
		}
	}

}
