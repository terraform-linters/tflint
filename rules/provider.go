package rules

import (
	"log"

	"github.com/wata727/tflint/rules/awsrules"
	"github.com/wata727/tflint/rules/terraformrules"
	"github.com/wata727/tflint/tflint"
)

// Rule is an implementation that receives a Runner and inspects for resources and modules.
type Rule interface {
	Name() string
	Enabled() bool
	Check(runner *tflint.Runner) error
}

// defaultRules is rules by default
var defaultRules = append(manualDefaultRules, modelRules...)
var deepCheckRules = append(manualDeepCheckRules, apiRules...)

var manualDefaultRules = []Rule{
	awsrules.NewAwsDBInstanceDefaultParameterGroupRule(),
	awsrules.NewAwsDBInstanceInvalidTypeRule(),
	awsrules.NewAwsDBInstancePreviousTypeRule(),
	awsrules.NewAwsElastiCacheClusterDefaultParameterGroupRule(),
	awsrules.NewAwsElastiCacheClusterInvalidTypeRule(),
	awsrules.NewAwsElastiCacheClusterPreviousTypeRule(),
	awsrules.NewAwsInstancePreviousTypeRule(),
	awsrules.NewAwsRouteNotSpecifiedTargetRule(),
	awsrules.NewAwsRouteSpecifiedMultipleTargetsRule(),
	awsrules.NewAwsS3BucketInvalidACLRule(),
	awsrules.NewAwsS3BucketInvalidRegionRule(),
	awsrules.NewAwsSpotFleetRequestInvalidExcessCapacityTerminationPolicyRule(),
	terraformrules.NewTerraformDashInResourceNameRule(),
	terraformrules.NewTerraformDocumentedOutputsRule(),
	terraformrules.NewTerraformDocumentedVariablesRule(),
	terraformrules.NewTerraformModulePinnedSourceRule(),
}

var manualDeepCheckRules = []Rule{
	awsrules.NewAwsInstanceInvalidAMIRule(),
	awsrules.NewAwsLaunchConfigurationInvalidImageIDRule(),
}

// AllRules returns map of rules indexed by name
func AllRulesMap() map[string]Rule {
	ret := map[string]Rule{}
	for _, rule := range append(defaultRules, deepCheckRules...) {
		ret[rule.Name()] = rule
	}
	return ret
}

// NewRules returns rules according to configuration
func NewRules(c *tflint.Config) []Rule {
	log.Print("[INFO] Prepare rules")

	rulesMap := AllRulesMap()
	totalEnabled := 0
	for _, rule := range rulesMap {
		if rule.Enabled() {
			totalEnabled += 1
		}
	}
	log.Printf("[INFO]   %d (%d) rules total", len(rulesMap), totalEnabled)
	for rulename, _ := range c.Rules {
		if _, ok := rulesMap[rulename]; !ok {
			log.Printf("[ERROR] Invalid rule:  %s", rulename)
		}
	}

	ret := []Rule{}
	allRules := []Rule{}

	if c.DeepCheck {
		log.Printf("[DEBUG] Deep check mode is enabled. Add deep check rules")
		allRules = append(defaultRules, deepCheckRules...)
	} else {
		allRules = defaultRules
	}

	for _, rule := range allRules {
		enabled := rule.Enabled()
		if r := c.Rules[rule.Name()]; r != nil {
			if r.Enabled {
				log.Printf("[DEBUG] `%s` is enabled", rule.Name())
			} else {
				log.Printf("[DEBUG] `%s` is disabled", rule.Name())
			}
			enabled = r.Enabled
		}

		if enabled {
			ret = append(ret, rule)
		}
	}
	log.Printf("[INFO]   %d rules enabled", len(ret))
	return ret
}
