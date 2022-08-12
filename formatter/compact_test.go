package formatter

import (
	"bytes"
	"errors"
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_compactPrint(t *testing.T) {
	cases := []struct {
		Name   string
		Issues tflint.Issues
		Error  error
		Stdout string
		Stderr string
	}{
		{
			Name:   "no issues",
			Issues: tflint.Issues{},
			Stdout: "",
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
			Stdout: `1 issue(s) found:

test.tf:1:1: Error - test (test_rule)
`,
		},
		{
			Name:   "error",
			Error:  errors.New("an error occurred"),
			Stderr: "an error occurred\n",
		},
		{
			Name:   "diagnostics",
			Error:  hclDiags(`resource "foo" "bar" {`),
			Stdout: "main.tf:1:22: error - Unclosed configuration block\n",
		},
	}

	for _, tc := range cases {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		formatter := &Formatter{Stdout: stdout, Stderr: stderr}

		formatter.compactPrint(tc.Issues, tc.Error, map[string][]byte{})

		if stdout.String() != tc.Stdout {
			t.Fatalf("Failed %s test: expected=%s, stdout=%s", tc.Name, tc.Stdout, stdout.String())
		}

		if stderr.String() != tc.Stderr {
			t.Fatalf("Failed %s test: expected=%s, stderr=%s", tc.Name, tc.Stderr, stderr.String())
		}
	}
}

func hclDiags(src string) hcl.Diagnostics {
	parser := hclparse.NewParser()
	_, diags := parser.ParseHCL([]byte(src), "main.tf")
	return diags
}
