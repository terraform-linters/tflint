package loader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	cases := []struct {
		Name  string
		Input string
		Error bool
	}{
		{
			Name:  "return parsed object",
			Input: "template.tf",
			Error: false,
		},
	}

	for _, tc := range cases {
		prev, _ := filepath.Abs(".")
		dir, _ := os.Getwd()
		defer os.Chdir(prev)
		testDir := dir + "/test-fixtures"
		os.Chdir(testDir)

		_, err := load(tc.Input)
		if tc.Error == true && err == nil {
			t.Fatalf("should be happen error.\n\ntestcase: %s", tc.Name)
		}
		if tc.Error == false && err != nil {
			t.Fatalf("should not be happen error.\nError: %s\n\ntestcase: %s", err, tc.Name)
		}
	}
}
