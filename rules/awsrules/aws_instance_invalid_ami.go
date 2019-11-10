package awsrules

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

// AwsInstanceInvalidAMIRule checks whether "aws_instance" has invalid AMI ID
type AwsInstanceInvalidAMIRule struct {
	resourceType  string
	attributeName string
	amiIDs        map[string]bool
}

// NewAwsInstanceInvalidAMIRule returns new rule with default attributes
func NewAwsInstanceInvalidAMIRule() *AwsInstanceInvalidAMIRule {
	return &AwsInstanceInvalidAMIRule{
		resourceType:  "aws_instance",
		attributeName: "ami",
		amiIDs:        map[string]bool{},
	}
}

// Name returns the rule name
func (r *AwsInstanceInvalidAMIRule) Name() string {
	return "aws_instance_invalid_ami"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsInstanceInvalidAMIRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsInstanceInvalidAMIRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsInstanceInvalidAMIRule) Link() string {
	return ""
}

// Check checks whether "aws_instance" has invalid AMI ID
func (r *AwsInstanceInvalidAMIRule) Check(runner *tflint.Runner) error {
	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

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
					if aerr, ok := err.(awserr.Error); ok {
						switch aerr.Code() {
						case "InvalidAMIID.Malformed":
							fallthrough
						case "InvalidAMIID.NotFound":
							fallthrough
						case "InvalidAMIID.Unavailable":
							runner.EmitIssue(
								r,
								fmt.Sprintf("\"%s\" is invalid AMI ID.", ami),
								attribute.Expr.Range(),
							)
							return nil
						}
					}
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
						fmt.Sprintf("\"%s\" is invalid AMI ID.", ami),
						attribute.Expr.Range(),
					)
				}
			}
			return nil
		})
	})
}
