package formatter

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_jsonPrint(t *testing.T) {
	cases := []struct {
		Name   string
		Issues tflint.Issues
		Error  error
		Stdout string
	}{
		{
			Name:   "no issues",
			Issues: tflint.Issues{},
			Stdout: `{"issues":[],"errors":[]}`,
		},
		{
			Name:   "error",
			Error:  fmt.Errorf("Failed to work; %w", errors.New("I don't feel like working")),
			Stdout: `{"issues":[],"errors":[{"message":"Failed to work; I don't feel like working","severity":"error"}]}`,
		},
		{
			Name: "error",
			Error: fmt.Errorf(
				"babel fish confused; %w",
				hcl.Diagnostics{
					&hcl.Diagnostic{
						Severity: hcl.DiagWarning,
						Summary:  "summary",
						Detail:   "detail",
						Subject: &hcl.Range{
							Filename: "filename",
							Start:    hcl.Pos{Line: 1, Column: 1, Byte: 0},
							End:      hcl.Pos{Line: 5, Column: 1, Byte: 4},
						},
					},
				},
			),
			Stdout: `{"issues":[],"errors":[{"summary":"summary","message":"detail","severity":"warning","range":{"filename":"filename","start":{"line":1,"column":1},"end":{"line":5,"column":1}}}]}`,
		},
	}

	for _, tc := range cases {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		formatter := &Formatter{Stdout: stdout, Stderr: stderr, Format: "json"}

		formatter.Print(tc.Issues, tc.Error, map[string][]byte{})

		if stdout.String() != tc.Stdout {
			t.Fatalf("Failed %s test: expected=%s, stdout=%s", tc.Name, tc.Stdout, stdout.String())
		}
	}
}
