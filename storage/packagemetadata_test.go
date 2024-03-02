package storage

import "testing"

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
