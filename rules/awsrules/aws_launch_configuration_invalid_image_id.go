package awsrules

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

// AwsLaunchConfigurationInvalidImageIdRule checks whether "aws_instance" has invalid AMI ID
type AwsLaunchConfigurationInvalidImageIdRule struct {
	resourceType  string
	attributeName string
	amiIDs        map[string]bool
}

// NewAwsLaunchConfigurationInvalidImageIdRule returns new rule with default attributes
func NewAwsLaunchConfigurationInvalidImageIdRule() *AwsLaunchConfigurationInvalidImageIdRule {
	return &AwsLaunchConfigurationInvalidImageIdRule{
		resourceType:  "aws_launch_configuration",
		attributeName: "image_id",
		amiIDs:        map[string]bool{},
	}
}

// Name returns the rule name
func (r *AwsLaunchConfigurationInvalidImageIdRule) Name() string {
	return "aws_launch_configuration_invalid_image_id"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsLaunchConfigurationInvalidImageIdRule) Enabled() bool {
	return true
}

// Type returns the rule severity
func (r *AwsLaunchConfigurationInvalidImageIdRule) Type() string {
	return issue.ERROR
}

// Link returns the rule reference link
func (r *AwsLaunchConfigurationInvalidImageIdRule) Link() string {
	return ""
}

// Check checks whether "aws_instance" has invalid AMI ID
func (r *AwsLaunchConfigurationInvalidImageIdRule) Check(runner *tflint.Runner) error {
	log.Printf("[INFO] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		var ami string
		err := runner.EvaluateExpr(attribute.Expr, &ami)

		return runner.EnsureNoError(err, func() error {
			if !r.amiIDs[ami] {
				log.Printf("[DEBUG] Fetch AMI images: %s", ami)
				resp, err := runner.AwsClient.EC2.DescribeImages(&ec2.DescribeImagesInput{
					ImageIds: aws.StringSlice([]string{ami}),
				})
				if err != nil {
					err := &tflint.Error{
						Code:    tflint.ExternalAPIError,
						Level:   tflint.ErrorLevel,
						Message: "An error occurred while describing images",
						Cause:   err,
					}
					log.Printf("[ERROR] %s", err)
					return err
				}

				if len(resp.Images) != 0 {
					for _, image := range resp.Images {
						r.amiIDs[*image.ImageId] = true
					}
				} else {
					runner.EmitIssue(
						r,
						fmt.Sprintf("\"%s\" is invalid image id.", ami),
						attribute.Expr.Range(),
					)
				}
			}
			return nil
		})
	})
}
