package formatter

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_junitPrint(t *testing.T) {
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
<testsuites>
  <testsuite tests="0" failures="0" time="0" name="">
    <properties></properties>
  </testsuite>
</testsuites>`,
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
<testsuites>
  <testsuite tests="1" failures="1" time="0" name="">
    <properties></properties>
    <testcase classname="test.tf" name="test_rule" time="0">
      <failure message="test" type="">line 1, col 1, Error - test (test_rule)</failure>
    </testcase>
  </testsuite>
</testsuites>`,
		},
	}

	for _, tc := range cases {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		formatter := &Formatter{Stdout: stdout, Stderr: stderr}

		formatter.junitPrint(tc.Issues, tc.Error, map[string][]byte{})

		if stdout.String() != tc.Stdout {
			t.Fatalf("%s: stdout did not match expected:\n%s", tc.Name, cmp.Diff(tc.Stdout, stdout.String()))
		}
	}
}
