package storage

import (
	"testing"
)

func TestFileVersion(t *testing.T) {
	type Test struct {
		Test     string
		Expected string
	}
	tests := []Test{
		{Test: "create-vite-5.0.0.tgz", Expected: "5.0.0"},
		{Test: "", Expected: ""},
		{Test: "create-vite-5.0.0", Expected: ""},
		{Test: "create", Expected: ""},
		{Test: "create-vite-5.0.0-beta.1.tgz", Expected: "5.0.0-beta.1"},
	}

	for _, test := range tests {
		actual := fileVersion("create-vite", test.Test)
		if actual != test.Expected {
			t.Errorf("expected %s but got %s", test.Expected, actual)
		}
	}
}
