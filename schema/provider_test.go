package schema

import (
	"testing"

	"os"
	"path/filepath"
	"reflect"

	"github.com/hashicorp/hcl/hcl/token"
	"github.com/k0kubun/pp"
)

func TestLoad(t *testing.T) {
	cases := []struct {
		Name   string
		Input  string
		Result []*Template
		Error  bool
	}{
		{
			Name: "provider",
			Input: `
		provider "aws" {
			region = "ap-southeast-2"
		}`,
			Result: Provider{
				Id: 1,
				Type: "aws",
				Source: &Source{
					File:  "test.tf",
					Pos:   3,
					Attrs: map[string]*Attribute{
						region: "ap-southeast-2"
					},
				},
			},
			Error: false,
		}
	}

	for _, tc := range cases {
		prev, _ := filepath.Abs(".")
		dir, _ := os.Getwd()
		defer os.Chdir(prev)
		testDir := dir + "/test-fixtures"
		os.Chdir(testDir)
		files := map[string][]byte{"test.tf": []byte(tc.Input)}
		schema, _ := Make(files)

		provider := schema[0].Providers[0]
		err := provider.Load()

		if tc.Error && err == nil {
			t.Fatalf("\nshould be happen error.\n\ntestcase: %s", tc.Name)
			continue
		}
		if !tc.Error && err != nil {
			t.Fatalf("\nshould not be happen error.\nError: %s\n\ntestcase: %s", err, tc.Name)
			continue
		}

		if !reflect.DeepEqual(provider, tc.Result) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(module.Templates), pp.Sprint(tc.Result), tc.Name)
		}
	}
}
