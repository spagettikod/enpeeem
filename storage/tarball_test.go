package storage

import "testing"

func TestNewTarball(t *testing.T) {
	type Test struct {
		Tarball     Tarball
		ExpectedURL string
	}
	tests := []Test{
		{
			Tarball: func() Tarball {
				t, _ := NewTarball("http://registry.npmjs.org", "", "create-vite", "create-vite-5.0.0.tgz")
				return t
			}(),
			ExpectedURL: "http://registry.npmjs.org/create-vite/-/create-vite-5.0.0.tgz",
		},
		{
			Tarball: func() Tarball {
				t, _ := NewTarball("http://registry.npmjs.org", "@babel", "plugin-transform-react-jsx-self", "plugin-transform-react-jsx-self-7.23.3.tgz")
				return t
			}(),
			ExpectedURL: "@babel/plugin-transform-react-jsx-self/-/plugin-transform-react-jsx-self-7.23.3.tgz",
		},
	}

	for _, test := range tests {
		if test.Tarball.RemoteURL != test.ExpectedURL {
			t.Errorf("expected %s but got %s", test.ExpectedURL, test.Tarball.RemoteURL)
		}
	}
}
