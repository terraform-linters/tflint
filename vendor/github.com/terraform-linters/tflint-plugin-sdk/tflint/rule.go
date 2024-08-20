package tflint

// DefaultRule implements optional fields in the rule interface.
// You can create a rule by embedding this rule.
type DefaultRule struct{}

// Link allows you to add a reference link to the rule.
// The default is empty.
func (r *DefaultRule) Link() string {
	return ""
}

// Metadata allows you to set any metadata to the rule.
// This value is never referenced by the SDK and can be used for your custom ruleset.
func (r *DefaultRule) Metadata() interface{} {
	return nil
}

func (r *DefaultRule) mustEmbedDefaultRule() {}

var _ Rule = &embedDefaultRule{}

type embedDefaultRule struct {
	DefaultRule
}

func (r *embedDefaultRule) Name() string              { return "" }
func (r *embedDefaultRule) Enabled() bool             { return true }
func (r *embedDefaultRule) Severity() Severity        { return ERROR }
func (r *embedDefaultRule) Check(runner Runner) error { return nil }
