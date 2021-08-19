package rules

import (
	"fmt"
	"log"

	"github.com/terraform-linters/tflint/rules/terraformrules"
	"github.com/terraform-linters/tflint/tflint"
)

// Rule is an implementation that receives a Runner and inspects for resources and modules.
type Rule interface {
	Name() string
	Enabled() bool
	Check(runner *tflint.Runner) error
}

// DefaultRules is rules by default
var DefaultRules = []Rule{
	terraformrules.NewTerraformDeprecatedIndexRule(),
	terraformrules.NewTerraformDeprecatedInterpolationRule(),
	terraformrules.NewTerraformDocumentedOutputsRule(),
	terraformrules.NewTerraformDocumentedVariablesRule(),
	terraformrules.NewTerraformModulePinnedSourceRule(),
	terraformrules.NewTerraformModuleVersionRule(),
	terraformrules.NewTerraformNamingConventionRule(),
	terraformrules.NewTerraformStandardModuleStructureRule(),
	terraformrules.NewTerraformTypedVariablesRule(),
	terraformrules.NewTerraformRequiredVersionRule(),
	terraformrules.NewTerraformRequiredProvidersRule(),
	terraformrules.NewTerraformWorkspaceRemoteRule(),
	terraformrules.NewTerraformUnusedDeclarationsRule(),
	terraformrules.NewTerraformUnusedRequiredProvidersRule(),
	terraformrules.NewTerraformCommentSyntaxRule(),
}

// CheckRuleNames returns map of rules indexed by name
func CheckRuleNames(ruleNames []string) error {
	log.Print("[INFO] Checking rules")

	rulesMap := map[string]Rule{}
	for _, rule := range DefaultRules {
		rulesMap[rule.Name()] = rule
	}

	totalEnabled := 0
	for _, rule := range rulesMap {
		if rule.Enabled() {
			totalEnabled++
		}
	}
	log.Printf("[INFO]   %d (%d) rules total", len(rulesMap), totalEnabled)
	for _, rule := range ruleNames {
		if _, ok := rulesMap[rule]; !ok {
			return fmt.Errorf("Rule not found: %s", rule)
		}
	}
	return nil
}

// NewRules returns rules according to configuration
func NewRules(c *tflint.Config) []Rule {
	log.Print("[INFO] Prepare rules")

	ret := []Rule{}

	if c.DisabledByDefault {
		log.Printf("[DEBUG] Only mode is enabled. Ignoring default rules")
	}

	for _, rule := range DefaultRules {
		enabled := rule.Enabled()
		if r := c.Rules[rule.Name()]; r != nil {
			if r.Enabled {
				log.Printf("[DEBUG] `%s` is enabled", rule.Name())
			} else {
				log.Printf("[DEBUG] `%s` is disabled", rule.Name())
			}
			enabled = r.Enabled
		} else if c.DisabledByDefault {
			enabled = false
		}

		if enabled {
			ret = append(ret, rule)
		}
	}
	if c.DisabledByDefault && len(ret) == 0 {
		log.Printf("[WARN] Only mode is enabled and no rules were provided")
	}
	log.Printf("[INFO]   %d default rules enabled", len(ret))
	return ret
}
