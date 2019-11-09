package formatter

import (
	"bytes"
	"errors"
	"testing"

	"github.com/terraform-linters/tflint/tflint"
)

func Test_jsonPrint(t *testing.T) {
	cases := []struct {
		Name   string
		Issues tflint.Issues
		Error  *tflint.Error
		Stdout string
	}{
		{
			Name:   "no issues",
			Issues: tflint.Issues{},
			Stdout: `{"issues":[],"errors":[]}`,
		},
		{
			Name:   "error",
			Error:  tflint.NewContextError("Failed to work", errors.New("I don't feel like working")),
			Stdout: `{"issues":[],"errors":[{"message":"I don't feel like working"}]}`,
		},
	}

	for _, tc := range cases {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		formatter := &Formatter{Stdout: stdout, Stderr: stderr}

		formatter.jsonPrint(tc.Issues, tc.Error)

		if stdout.String() != tc.Stdout {
			t.Fatalf("Failed %s test: expected=%s, stdout=%s", tc.Name, tc.Stdout, stdout.String())
		}
	}
}
