package handle

import "testing"

func TestSplitPkg(t *testing.T) {
	type Test struct {
		Test    string
		Scope   string
		Package string
	}

	tests := []Test{
		{"@types/react", "@types", "react"},
		{"react", "", "react"},
		{"", "", ""},
		{"a/b/c", "", "a/b/c"},
		{"@a/b/c", "", ""},
	}

	for _, test := range tests {
		actualScope, actualPackage := splitPkg(test.Test)
		if actualScope != test.Scope {
			t.Errorf("expected scope %s got %s", test.Scope, actualScope)
		}
		if actualPackage != test.Package {
			t.Errorf("expected package %s got %s", test.Package, actualPackage)
		}
	}
}
