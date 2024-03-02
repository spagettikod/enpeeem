package storage

import "testing"

func TestString(t *testing.T) {
	type Test struct {
		Tarball  Tarball
		Expected string
	}
	tests := []Test{
		{
			Tarball: func() Tarball {
				pkg, err := NewPackage("https://registry.npmjs.org", "", "create-vite")
				if err != nil {
					t.Fatal(err)
				}
				return NewTarball(pkg, "create-vite-5.0.0.tgz")
			}(),
			Expected: "registry.npmjs.org/create-vite/create-vite-5.0.0.tgz",
		},
		{
			Tarball: func() Tarball {
				pkg, err := NewPackage("https://registry.npmjs.org", "@babel", "plugin-transform-react-jsx-self")
				if err != nil {
					t.Fatal(err)
				}
				return NewTarball(pkg, "plugin-transform-react-jsx-self-7.23.3.tgz")
			}(),
			Expected: "registry.npmjs.org/@babel/plugin-transform-react-jsx-self/plugin-transform-react-jsx-self-7.23.3.tgz",
		},
	}

	for _, test := range tests {
		if test.Tarball.String() != test.Expected {
			t.Errorf("expected %s but got %s", test.Expected, test.Tarball.RemoteURL())
		}
	}
}

func TestRemoteURL(t *testing.T) {
	type Test struct {
		Tarball     Tarball
		ExpectedURL string
	}
	tests := []Test{
		{
			Tarball: func() Tarball {
				pkg, err := NewPackage("https://registry.npmjs.org", "", "create-vite")
				if err != nil {
					t.Fatal(err)
				}
				return NewTarball(pkg, "create-vite-5.0.0.tgz")
			}(),
			ExpectedURL: "https://registry.npmjs.org/create-vite/-/create-vite-5.0.0.tgz",
		},
		{
			Tarball: func() Tarball {
				pkg, err := NewPackage("https://registry.npmjs.org", "@babel", "plugin-transform-react-jsx-self")
				if err != nil {
					t.Fatal(err)
				}
				return NewTarball(pkg, "plugin-transform-react-jsx-self-7.23.3.tgz")
			}(),
			ExpectedURL: "https://registry.npmjs.org/@babel/plugin-transform-react-jsx-self/-/plugin-transform-react-jsx-self-7.23.3.tgz",
		},
	}

	for _, test := range tests {
		if test.Tarball.RemoteURL() != test.ExpectedURL {
			t.Errorf("expected %s but got %s", test.ExpectedURL, test.Tarball.RemoteURL())
		}
	}
}

func TestTarballFromURI(t *testing.T) {
	type Test struct {
		Test     string
		Expected Tarball
	}
	tests := []Test{
		{
			Test: "registry.npmjs.org/ansi-styles/ansi-styles-3.2.1.tgz",
			Expected: Tarball{
				pkg:  Package{Registry: "registry.npmjs.org", Scope: "", Name: "ansi-styles"},
				Name: "ansi-styles-3.2.1.tgz",
			},
		},
		{
			Test: "registry.npmjs.org/@babel/parser/parser-7.24.0.tgz",
			Expected: Tarball{
				pkg:  Package{Registry: "registry.npmjs.org", Scope: "@babel", Name: "parser"},
				Name: "parser-7.24.0.tgz",
			},
		},
	}

	for _, test := range tests {
		actual, err := TarballFromURI(test.Test)
		if err != nil {
			t.Fatal(err)
		}
		if test.Expected.Package().Registry != actual.Package().Registry {
			t.Errorf("%s, expected registry %s but got %s", test.Test, test.Expected.Package().Registry, actual.Package().Registry)
		}
		if test.Expected.Package().Scope != actual.Package().Scope {
			t.Errorf("%s, expected scope %s but got %s", test.Test, test.Expected.Package().Scope, actual.Package().Scope)
		}
		if test.Expected.Package().Name != actual.Package().Name {
			t.Errorf("%s, expected package %s but got %s", test.Test, test.Expected.Package().Name, actual.Package().Name)
		}
		if test.Expected.Name != actual.Name {
			t.Errorf("%s, expected name %s but got %s", test.Test, test.Expected.Name, actual.Name)
		}
	}
}
