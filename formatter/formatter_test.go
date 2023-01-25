package formatter

import (
	sdk "github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint/tflint"
)

type testRule struct{}

func (r *testRule) Name() string {
	return "test_rule"
}

func (r *testRule) Enabled() bool {
	return true
}

func (r *testRule) Severity() tflint.Severity {
	return sdk.ERROR
}

func (r *testRule) Link() string {
	return "https://github.com"
}
