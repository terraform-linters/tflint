package printer

import (
	"bytes"
	"testing"

	"github.com/wata727/tflint/issue"
)

func TestCheckstylePrint(t *testing.T) {
	cases := []struct {
		Name   string
		Input  []*issue.Issue
		Result string
	}{
		{
			Name:  "no issues",
			Input: []*issue.Issue{},
			Result: `<?xml version="1.0" encoding="UTF-8"?>
<checkstyle></checkstyle>`,
		},
		{
			Name: "multi files",
			Input: []*issue.Issue{
				{
					Detector: "error detector",
					File:     "template.tf",
					Line:     1,
					Type:     "ERROR",
					Message:  "example error message",
				},
				{
					Detector: "notice detector",
					File:     "application.tf",
					Line:     10,
					Type:     "NOTICE",
					Message:  "example notice message",
					Link:     "https://github.com/wata727/tflint",
				},
				{
					Detector: "warning detector",
					File:     "template.tf",
					Line:     3,
					Type:     "WARNING",
					Message:  "example warning message",
				},
			},
			Result: `<?xml version="1.0" encoding="UTF-8"?>
<checkstyle>
  <file name="application.tf">
    <error detector="notice detector" line="10" severity="info" message="example notice message" link="https://github.com/wata727/tflint"></error>
  </file>
  <file name="template.tf">
    <error detector="error detector" line="1" severity="error" message="example error message" link=""></error>
    <error detector="warning detector" line="3" severity="warning" message="example warning message" link=""></error>
  </file>
</checkstyle>`,
		},
	}

	for _, tc := range cases {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		p := NewPrinter(stdout, stderr)
		p.CheckstylePrint(tc.Input)
		result := stdout.String()

		if result != tc.Result {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", result, tc.Result, tc.Name)
		}
	}
}
