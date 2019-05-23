package awsrules

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

// AwsInstanceInvalidIAMProfileRule checks whether profile actually exists
type AwsInstanceInvalidIAMProfileRule struct {
	resourceType  string
	attributeName string
	profiles      map[string]bool
	dataPrepared  bool
}

// NewAwsInstanceInvalidIAMProfileRule returns new rule with default attributes
func NewAwsInstanceInvalidIAMProfileRule() *AwsInstanceInvalidIAMProfileRule {
	return &AwsInstanceInvalidIAMProfileRule{
		resourceType:  "aws_instance",
		attributeName: "iam_instance_profile",
		profiles:      map[string]bool{},
		dataPrepared:  false,
	}
}

// Name returns the rule name
func (r *AwsInstanceInvalidIAMProfileRule) Name() string {
	return "aws_instance_invalid_iam_profile"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsInstanceInvalidIAMProfileRule) Enabled() bool {
	return true
}

// Type returns the rule severity
func (r *AwsInstanceInvalidIAMProfileRule) Type() string {
	return issue.ERROR
}

// Link returns the rule reference link
func (r *AwsInstanceInvalidIAMProfileRule) Link() string {
	return ""
}

// Check checks whether `iam_instance_profile` are included in the list retrieved by `ListInstanceProfiles`
func (r *AwsInstanceInvalidIAMProfileRule) Check(runner *tflint.Runner) error {
	log.Printf("[INFO] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		if !r.dataPrepared {
			log.Print("[DEBUG] Fetch instance profiles")
			resp, err := runner.AwsClient.IAM.ListInstanceProfiles(&iam.ListInstanceProfilesInput{})
			if err != nil {
				err := &tflint.Error{
					Code:    tflint.ExternalAPIError,
					Level:   tflint.ErrorLevel,
					Message: "An error occurred while describing instance profiles",
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
