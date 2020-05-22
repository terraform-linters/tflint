package terraformrules

import (
	"fmt"
	"log"
	"sort"

	"github.com/terraform-linters/tflint/tflint"
)

// TerraformBlacklistedResourcesRule checks whether variables have a type declared
type TerraformBlacklistedResourcesRule struct{}

type terraformBlacklistedResourcesRuleConfig struct {
	Types map[string]string `hcl:"types"`
}

// TerraformBlacklistedResourcesRule returns a new rule
func NewTerraformBlacklistedResourcesRule() *TerraformBlacklistedResourcesRule {
	return &TerraformBlacklistedResourcesRule{}
}

// Name returns the rule name
func (r *TerraformBlacklistedResourcesRule) Name() string {
	return "terraform_blacklisted_resources"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformBlacklistedResourcesRule) Enabled() bool {
	return false
}

// Severity returns the rule severity
func (r *TerraformBlacklistedResourcesRule) Severity() string {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformBlacklistedResourcesRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

// TODO: comment
func (r *TerraformBlacklistedResourcesRule) Check(runner *tflint.Runner) error {
	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	config := terraformBlacklistedResourcesRuleConfig{}
	config.Types = make(map[string]string)

	if err := runner.DecodeRuleConfig(r.Name(), &config); err != nil {
		return err
	}

	// Since the resources are stored as a map, there is no guarantee of the order in which
	// they will be iterated over. The resources need to be iterated in a consistent manner
	// to ensure that the tests are consistent
	managedResources := runner.TFConfig.Module.ManagedResources
	resourceNames := make([]string, 0, len(managedResources))
	for k := range managedResources {
		resourceNames = append(resourceNames, k)
	}
	sort.Strings(resourceNames)

	for _, resourceName := range resourceNames {
		resource := managedResources[resourceName]
		msg, exists := config.Types[resource.Type]
		if exists {
			runner.EmitIssue(
				r,
				fmt.Sprintf("`%v` resource type is blacklisted\n\n%v", resource.Type, msg),
				resource.TypeRange,
			)
		}
	}

	return nil
}
