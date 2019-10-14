package formatter

import "github.com/wata727/tflint/tflint"

type testRule struct{}

func (r *testRule) Name() string {
	return "test_rule"
}
func (r *testRule) Enabled() bool {
	return true
}
func (r *testRule) Severity() string {
	return tflint.ERROR
}
func (r *testRule) Link() string {
	return "https://github.com"
}
func (r *testRule) Check(runner *tflint.Runner) error {
	return nil
}
