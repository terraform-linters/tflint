package formatter

import (
	"bytes"
	"fmt"
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
    <error source="test_rule" line="1" column="1" severity="error" message="test" link="https://github.com" rule="test_rule"></error>
  </file>
</checkstyle>`,
		},
		{
			Name:   "error only",
			Issues: tflint.Issues{},
			Error:  fmt.Errorf("Failed to check ruleset"),
			Stdout: `<?xml version="1.0" encoding="UTF-8"?>
<checkstyle>
  <file name="(application)">
    <error source="(application)" line="0" column="0" severity="error" message="Failed to check ruleset" link="" rule="(application)"></error>
  </file>
</checkstyle>`,
		},
		{
			Name: "issues and errors",
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
			Error: fmt.Errorf("Failed to check ruleset"),
			Stdout: `<?xml version="1.0" encoding="UTF-8"?>
<checkstyle>
  <file name="(application)">
    <error source="(application)" line="0" column="0" severity="error" message="Failed to check ruleset" link="" rule="(application)"></error>
  </file>
  <file name="test.tf">
    <error source="test_rule" line="1" column="1" severity="error" message="test" link="https://github.com" rule="test_rule"></error>
  </file>
</checkstyle>`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			formatter := &Formatter{Stdout: stdout, Stderr: stderr}

			formatter.checkstylePrint(tc.Issues, tc.Error, map[string][]byte{})

			if stdout.String() != tc.Stdout {
				t.Fatalf("expected=%s, stdout=%s", tc.Stdout, stdout.String())
			}
		})
	}
}
