package rules

import (
	"github.com/wata727/tflint/rules/awsrules"
	"github.com/wata727/tflint/tflint"
)

// Rule is an implementation that receives a Runner and inspects for resources and modules.
type Rule interface {
	Name() string
	Enabled() bool
	Check(runner *tflint.Runner) error
}

// DefaultRules is rules by default
var DefaultRules = []Rule{
	awsrules.NewAwsInstanceInvalidTypeRule(),
}

var deepCheckRules = []Rule{
	awsrules.NewAwsInstanceInvalidAMIRule(),
}

// NewRules returns rules according to configuration
func NewRules(c *tflint.Config) []Rule {
	ret := []Rule{}
	allRules := []Rule{}

	if c.DeepCheck {
		allRules = append(DefaultRules, deepCheckRules...)
	} else {
		allRules = DefaultRules
	}

	for _, rule := range allRules {
		if r := c.Rules[rule.Name()]; r != nil {
			if r.Enabled {
				ret = append(ret, rule)
			}
		} else {
			if !c.IgnoreRule[rule.Name()] && rule.Enabled() {
				ret = append(ret, rule)
			}
		}
	}

	return ret
}
