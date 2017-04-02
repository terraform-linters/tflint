package printer

import (
	"testing"
)

func TestValidateFormat(t *testing.T) {
	cases := []struct {
		Name   string
		Input  string
		Result bool
	}{
		{
			Name:   "default is valid format",
			Input:  "default",
			Result: true,
		},
		{
			Name:   "json is valid format",
			Input:  "json",
			Result: true,
		},
		{
			Name:   "checkstyle is valid format",
			Input:  "checkstyle",
			Result: true,
		},
		{
			Name:   "yaml is invalid format",
			Input:  "yaml",
			Result: false,
		},
	}

	for _, tc := range cases {
		result := ValidateFormat(tc.Input)
		if result != tc.Result {
			t.Fatalf("\nBad: %t\nExpected: %t\n\ntestcase: %s", result, tc.Result, tc.Name)
		}
	}
}
