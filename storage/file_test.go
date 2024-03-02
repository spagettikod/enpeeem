package storage

import (
	"testing"
)

func TestFileVersion(t *testing.T) {
	type Test struct {
		PackageName string
		Filename    string
		Expected    string
	}
	tests := []Test{
		{PackageName: "create-vite", Filename: "create-vite-5.0.0.tgz", Expected: "5.0.0"},
		{PackageName: "create-vite", Filename: "", Expected: ""},
		{PackageName: "create-vite", Filename: "create-vite-5.0.0", Expected: ""},
		{PackageName: "create-vite", Filename: "create", Expected: ""},
		{PackageName: "create-vite", Filename: "create-vite-5.0.0-beta.1.tgz", Expected: "5.0.0-beta.1"},
		{PackageName: "@types/react", Filename: "react-0.0.0.tgz", Expected: "0.0.0"},
	}

	for _, test := range tests {
		actual := fileVersion(test.PackageName, test.Filename)
		if actual != test.Expected {
			t.Errorf("expected %s but got %s", test.Expected, actual)
		}
	}
}
