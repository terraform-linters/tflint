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
		Fix    bool
		Stdout string
	}{
		{
			Name:   "no issues",
			Issues: tflint.Issues{},
			Stdout: `{"issues":[],"errors":[]}`,
		},
		{
			Name: "fixable issue without fix",
			Issues: tflint.Issues{
				{
					Rule:    &testRule{},
					Message: "test message",
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1},
						End:      hcl.Pos{Line: 1, Column: 5},
					},
					Fixable: true,
				},
			},
			Fix:    false,
			Stdout: `{"issues":[{"rule":{"name":"test_rule","severity":"error","link":"https://github.com"},"message":"test message","range":{"filename":"test.tf","start":{"line":1,"column":1},"end":{"line":1,"column":5}},"callers":[],"fixable":true,"fixed":false}],"errors":[]}`,
		},
		{
			Name: "fixable issue with fix",
			Issues: tflint.Issues{
				{
					Rule:    &testRule{},
					Message: "test message",
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1},
						End:      hcl.Pos{Line: 1, Column: 5},
					},
					Fixable: true,
				},
			},
			Fix:    true,
			Stdout: `{"issues":[{"rule":{"name":"test_rule","severity":"error","link":"https://github.com"},"message":"test message","range":{"filename":"test.tf","start":{"line":1,"column":1},"end":{"line":1,"column":5}},"callers":[],"fixable":true,"fixed":true}],"errors":[]}`,
		},
		{
			Name: "non-fixable issue",
			Issues: tflint.Issues{
				{
					Rule:    &testRule{},
					Message: "test message",
					Range: hcl.Range{
						Filename: "test.tf",
						Start:    hcl.Pos{Line: 1, Column: 1},
						End:      hcl.Pos{Line: 1, Column: 5},
					},
					Fixable: false,
				},
			},
			Fix:    false,
			Stdout: `{"issues":[{"rule":{"name":"test_rule","severity":"error","link":"https://github.com"},"message":"test message","range":{"filename":"test.tf","start":{"line":1,"column":1},"end":{"line":1,"column":5}},"callers":[],"fixable":false,"fixed":false}],"errors":[]}`,
		},
		{
			Name:   "error",
			Error:  fmt.Errorf("Failed to work; %w", errors.New("I don't feel like working")),
			Stdout: `{"issues":[],"errors":[{"message":"Failed to work; I don't feel like working","severity":"error"}]}`,
		},
		{
			Name: "diagnostics",
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
		{
			Name: "joined errors",
			Error: errors.Join(
				errors.New("an error occurred"),
				errors.New("failed"),
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
			Stdout: `{"issues":[],"errors":[{"message":"an error occurred","severity":"error"},{"message":"failed","severity":"error"},{"summary":"summary","message":"detail","severity":"warning","range":{"filename":"filename","start":{"line":1,"column":1},"end":{"line":5,"column":1}}}]}`,
		},
	}

	for _, tc := range cases {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		formatter := &Formatter{Stdout: stdout, Stderr: stderr, Format: "json", Fix: tc.Fix}

		formatter.Print(tc.Issues, tc.Error, map[string][]byte{})

		if stdout.String() != tc.Stdout {
			t.Fatalf("Failed %s test: expected=%s, stdout=%s", tc.Name, tc.Stdout, stdout.String())
		}
	}
}
