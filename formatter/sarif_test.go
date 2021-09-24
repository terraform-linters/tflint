package formatter

import (
	"bytes"
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_sarifPrint(t *testing.T) {
	cases := []struct {
		Name   string
		Issues tflint.Issues
		Error  *tflint.Error
		Stdout string
	}{
		{
			Name:   "no issues",
			Issues: tflint.Issues{},
			Stdout: `{
  "version": "2.1.0",
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "tflint",
          "informationUri": "https://github.com/terraform-linters/tflint"
        }
      },
      "results": []
    }
  ]
}`,
		},
		{
			Name: "issues",
			Issues: tflint.Issues{
				{
					Rule:    &testRule{},
					Message: "test",
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
						End:      hcl.Pos{Line: 1, Column: 4, Byte: 3},
					},
				},
			},
			Stdout: `{
  "version": "2.1.0",
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "tflint",
          "informationUri": "https://github.com/terraform-linters/tflint",
          "rules": [
            {
              "id": "test_rule",
              "shortDescription": null,
              "helpUri": "https://github.com"
            }
          ]
        }
      },
      "results": [
        {
          "ruleId": "test_rule",
          "level": "error",
          "message": {
            "text": "test"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "test.tf"
                },
                "region": {
                  "startLine": 1,
                  "startColumn": 1,
                  "endLine": 1,
                  "endColumn": 4
                }
              }
            }
          ]
        }
      ]
    }
  ]
}`,
		},
	}

	for _, tc := range cases {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		formatter := &Formatter{Stdout: stdout, Stderr: stderr}

		formatter.sarifPrint(tc.Issues, tc.Error, map[string][]byte{})

		if stdout.String() != tc.Stdout {
			t.Fatalf("Failed %s test: expected=%s, stdout=%s", tc.Name, tc.Stdout, stdout.String())
		}
	}
}
