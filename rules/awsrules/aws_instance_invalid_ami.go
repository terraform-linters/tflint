package awsrules

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
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

// Check checks whether "aws_instance" has invalid AMI ID
func (r *AwsInstanceInvalidAMIRule) Check(runner *tflint.Runner) error {
	resources := runner.LookupResourcesByType(r.resourceType)
	if len(resources) == 0 {
		return nil
	}

	resp, err := runner.AwsClient.EC2.DescribeImages(&ec2.DescribeImagesInput{})
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
	for _, image := range resp.Images {
		r.amiIDs[*image.ImageId] = true
	}

	for _, resource := range resources {
		body, _, diags := resource.Config.PartialContent(&hcl.BodySchema{
			Attributes: []hcl.AttributeSchema{
				{
					Name: r.attributeName,
				},
			},
		})
		if diags.HasErrors() {
			panic(diags)
		}

		if attribute, ok := body.Attributes[r.attributeName]; ok {
			var ami string
			err := runner.EvaluateExpr(attribute.Expr, &ami)
			if appErr, ok := err.(*tflint.Error); ok {
				switch appErr.Level {
				case tflint.WarningLevel:
					continue
				case tflint.ErrorLevel:
					return appErr
				default:
					panic(appErr)
				}
			}

			if !r.amiIDs[ami] {
				runner.Issues = append(runner.Issues, &issue.Issue{
					Detector: r.Name(),
					Type:     issue.ERROR,
					Message:  fmt.Sprintf("\"%s\" is invalid AMI ID.", ami),
					Line:     attribute.Range.Start.Line,
					File:     runner.GetFileName(attribute.Range.Filename),
				})
			}
		}
	}

	return nil
}
