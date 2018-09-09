package rules

import (
	"log"

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
	awsrules.NewAwsDBInstanceReadablePasswordRule(),
	awsrules.NewAwsInstanceInvalidTypeRule(),
}

var deepCheckRules = []Rule{
	awsrules.NewAwsInstanceInvalidAMIRule(),
}

// NewRules returns rules according to configuration
func NewRules(c *tflint.Config) []Rule {
	log.Print("[INFO] Prepare rules")

	ret := []Rule{}
	allRules := []Rule{}

	if c.DeepCheck {
		log.Printf("[DEBUG] Deep check mode is enabled. Add deep check rules")
		allRules = append(DefaultRules, deepCheckRules...)
	} else {
		allRules = DefaultRules
	}

	for _, rule := range allRules {
		if r := c.Rules[rule.Name()]; r != nil {
			if r.Enabled {
				log.Printf("[DEBUG] `%s` is enabled", rule.Name())
				ret = append(ret, rule)
			} else {
				log.Printf("[DEBUG] `%s` is disabled", rule.Name())
			}
		} else {
			if !c.IgnoreRule[rule.Name()] && rule.Enabled() {
				log.Printf("[DEBUG] `%s` is enabled", rule.Name())
				ret = append(ret, rule)
			} else {
				log.Printf("[DEBUG] `%s` is disabled", rule.Name())
			}
		}
	}

	return ret
}
