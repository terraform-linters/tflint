package awsrules

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/tflint"
)

// AwsLaunchConfigurationInvalidIAMProfileRule checks whether profile actually exists
type AwsLaunchConfigurationInvalidIAMProfileRule struct {
	resourceType  string
	attributeName string
	profiles      map[string]bool
	dataPrepared  bool
}

// NewAwsLaunchConfigurationInvalidIAMProfileRule returns new rule with default attributes
func NewAwsLaunchConfigurationInvalidIAMProfileRule() *AwsLaunchConfigurationInvalidIAMProfileRule {
	return &AwsLaunchConfigurationInvalidIAMProfileRule{
		resourceType:  "aws_launch_configuration",
		attributeName: "iam_instance_profile",
		profiles:      map[string]bool{},
		dataPrepared:  false,
	}
}

// Name returns the rule name
func (r *AwsLaunchConfigurationInvalidIAMProfileRule) Name() string {
	return "aws_launch_configuration_invalid_iam_profile"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsLaunchConfigurationInvalidIAMProfileRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsLaunchConfigurationInvalidIAMProfileRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsLaunchConfigurationInvalidIAMProfileRule) Link() string {
	return ""
}

// Check checks whether `iam_instance_profile` are included in the list retrieved by `ListInstanceProfiles`
func (r *AwsLaunchConfigurationInvalidIAMProfileRule) Check(runner *tflint.Runner) error {
	log.Printf("[INFO] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		if !r.dataPrepared {
			log.Print("[DEBUG] Fetch instance profiles")
			resp, err := runner.AwsClient.IAM.ListInstanceProfiles(&iam.ListInstanceProfilesInput{})
			if err != nil {
				err := &tflint.Error{
					Code:    tflint.ExternalAPIError,
					Level:   tflint.ErrorLevel,
					Message: "An error occurred while describing launch configuration profiles",
					Cause:   err,
				}
				log.Printf("[ERROR] %s", err)
				return err
			}
			for _, iamProfile := range resp.InstanceProfiles {
				r.profiles[*iamProfile.InstanceProfileName] = true
			}
			r.dataPrepared = true
		}

		var profile string
		err := runner.EvaluateExpr(attribute.Expr, &profile)

		return runner.EnsureNoError(err, func() error {
			if !r.profiles[profile] {
				runner.EmitIssue(
					r,
					fmt.Sprintf("\"%s\" is invalid IAM profile name.", profile),
					attribute.Expr.Range(),
				)
			}
			return nil
		})
	})
}
