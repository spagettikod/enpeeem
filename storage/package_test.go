package storage

import (
	"testing"
)

func TestNewPackage(t *testing.T) {
	type Test struct {
		Registry string
		Scope    string
		Name     string
	}
	tests := []Test{
		{
			Registry: "registry.npmjs.org",
			Scope:    "",
			Name:     "ansi-styles",
		},
		{
			Registry: "registry.npmjs.org",
			Scope:    "@babel",
			Name:     "parser",
		},
	}

	for _, test := range tests {
		pkg, err := NewPackage(test.Registry, test.Scope, test.Name)
		if err != nil {
			t.Fatal(err)
		}
		if pkg.Registry != test.Registry {
			t.Errorf("expected %s got %s", test.Registry, pkg.Registry)
		}
		if pkg.Scope != test.Scope {
			t.Errorf("expected %s got %s", test.Scope, pkg.Scope)
		}
		if pkg.Name != test.Name {
			t.Errorf("expected %s got %s", test.Name, pkg.Name)
		}
	}
}

func TestPackageMetadataFromURI(t *testing.T) {
	type Test struct {
		Test     string
		Expected Package
	}
	tests := []Test{
		{
			Test:     "registry.npmjs.org/ansi-styles",
			Expected: Package{Registry: "registry.npmjs.org", Scope: "", Name: "ansi-styles"},
		},
		{
			Test:     "registry.npmjs.org/@babel/parser",
			Expected: Package{Registry: "registry.npmjs.org", Scope: "@babel", Name: "parser"},
		},
	}

	for _, test := range tests {
		actual, err := PackageMetadataFromURI(test.Test)
		if err != nil {
			t.Fatal(err)
		}
		if test.Expected.Registry != actual.Registry {
			t.Errorf("%s, expected registry %s but got %s", test.Test, test.Expected.Registry, actual.Registry)
		}
		if test.Expected.Scope != actual.Scope {
			t.Errorf("%s, expected scope %s but got %s", test.Test, test.Expected.Scope, actual.Scope)
		}
		if test.Expected.Name != actual.Name {
			t.Errorf("%s, expected name %s but got %s", test.Test, test.Expected.Name, actual.Name)
		}
		if test.Expected.Name != actual.Name {
			t.Errorf("%s, expected name %s but got %s", test.Test, test.Expected.Name, actual.Name)
		}
	}
}
