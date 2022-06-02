package formatter

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_codeClimatePrint(t *testing.T) {
	cases := []struct {
		Name   string
		Issues tflint.Issues
		Error  error
		Stdout string
	}{
		{
			Name:   "no issues",
			Issues: tflint.Issues{},
			Stdout: `[]`,
		},
		{
			Name:   "error",
			Error:  fmt.Errorf("Failed to work; %w", errors.New("I don't feel like working")),
			Stdout: `[{"type":"issue","check_name":"TFLint Error","description":"Failed to work; I don't feel like working","categories":["Bug Risk"],"location":{"path":"","positions":{"begin":{"line":0},"end":{"line":0}}},"fingerprint":"57394bce0549274282f603462359aad7","severity":"critical"}]`,
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
			Stdout: `[{"type":"issue","check_name":"TFLint Error","description":"detail","content":"summary","categories":["Bug Risk"],"location":{"path":"filename","positions":{"begin":{"line":1,"column":1},"end":{"line":5,"column":1}}},"fingerprint":"33c423ea5450bd2d6209c131fad9398d","severity":"minor"}]`,
		},
	}

	for _, tc := range cases {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		formatter := &Formatter{Stdout: stdout, Stderr: stderr, Format: "codeclimate"}

		formatter.Print(tc.Issues, tc.Error, map[string][]byte{})

		if stdout.String() != tc.Stdout {
			t.Fatalf("Failed %s test: expected=%s, stdout=%s", tc.Name, tc.Stdout, stdout.String())
		}
	}
}
