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

// DefaultRules is rules by default
var DefaultRules = append(manualRules, modelRules...)

var manualRules = []Rule{
	awsrules.NewAwsDBInstanceDefaultParameterGroupRule(),
	awsrules.NewAwsDBInstanceInvalidTypeRule(),
	awsrules.NewAwsDBInstancePreviousTypeRule(),
	awsrules.NewAwsDBInstanceReadablePasswordRule(),
	awsrules.NewAwsElastiCacheClusterDefaultParameterGroupRule(),
	awsrules.NewAwsElastiCacheClusterInvalidTypeRule(),
	awsrules.NewAwsElastiCacheClusterPreviousTypeRule(),
	awsrules.NewAwsInstanceDefaultStandardVolumeRule(),
	awsrules.NewAwsInstancePreviousTypeRule(),
	awsrules.NewAwsLaunchConfigurationInvalidTypeRule(),
	awsrules.NewAwsRouteNotSpecifiedTargetRule(),
	awsrules.NewAwsRouteSpecifiedMultipleTargetsRule(),
	terraformrules.NewTerraformModulePinnedSourceRule(),
}

var deepCheckRules = []Rule{
	awsrules.NewAwsALBInvalidSecurityGroupRule(),
	awsrules.NewAwsALBInvalidSubnetRule(),
	awsrules.NewAwsDBInstanceInvalidDBSubnetGroupRule(),
	awsrules.NewAwsDBInstanceInvalidOptionGroupRule(),
	awsrules.NewAwsDBInstanceInvalidParameterGroupRule(),
	awsrules.NewAwsDBInstanceInvalidVPCSecurityGroupRule(),
	awsrules.NewAwsElastiCacheClusterInvalidParameterGroupRule(),
	awsrules.NewAwsElastiCacheClusterInvalidSecurityGroupRule(),
	awsrules.NewAwsElastiCacheClusterInvalidSubnetGroupRule(),
	awsrules.NewAwsELBInvalidInstanceRule(),
	awsrules.NewAwsELBInvalidSecurityGroupRule(),
	awsrules.NewAwsELBInvalidSubnetRule(),
	awsrules.NewAwsInstanceInvalidAMIRule(),
	awsrules.NewAwsInstanceInvalidIAMProfileRule(),
	awsrules.NewAwsInstanceInvalidKeyNameRule(),
	awsrules.NewAwsInstanceInvalidSubnetRule(),
	awsrules.NewAwsInstanceInvalidVPCSecurityGroupRule(),
	awsrules.NewAwsLaunchConfigurationInvalidImageIDRule(),
	awsrules.NewAwsLaunchConfigurationInvalidIAMProfileRule(),
	awsrules.NewAwsRouteInvalidEgressOnlyGatewayRule(),
	awsrules.NewAwsRouteInvalidGatewayRule(),
	awsrules.NewAwsRouteInvalidInstanceRule(),
	awsrules.NewAwsRouteInvalidNatGatewayRule(),
	awsrules.NewAwsRouteInvalidNetworkInterfaceRule(),
	awsrules.NewAwsRouteInvalidRouteTableRule(),
	awsrules.NewAwsRouteInvalidVPCPeeringConnectionRule(),
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
