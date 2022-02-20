package formatter

import (
	"bytes"
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_checkstylePrint(t *testing.T) {
	cases := []struct {
		Name   string
		Issues tflint.Issues
		Error  error
		Stdout string
	}{
		{
			Name:   "no issues",
			Issues: tflint.Issues{},
			Stdout: `<?xml version="1.0" encoding="UTF-8"?>
<checkstyle></checkstyle>`,
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
			Stdout: `<?xml version="1.0" encoding="UTF-8"?>
<checkstyle>
  <file name="test.tf">
    <error rule="test_rule" line="1" column="1" severity="error" message="test" link="https://github.com"></error>
  </file>
</checkstyle>`,
		},
	}

	for _, tc := range cases {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		formatter := &Formatter{Stdout: stdout, Stderr: stderr}

		formatter.checkstylePrint(tc.Issues, tc.Error, map[string][]byte{})

		if stdout.String() != tc.Stdout {
			t.Fatalf("Failed %s test: expected=%s, stdout=%s", tc.Name, tc.Stdout, stdout.String())
		}
	}
}
