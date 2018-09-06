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
	dataPrepared  bool
}

// NewAwsInstanceInvalidAMIRule returns new rule with default attributes
func NewAwsInstanceInvalidAMIRule() *AwsInstanceInvalidAMIRule {
	return &AwsInstanceInvalidAMIRule{
		resourceType:  "aws_instance",
		attributeName: "ami",
		amiIDs:        map[string]bool{},
		dataPrepared:  false,
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
	log.Printf("[INFO] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		if !r.dataPrepared {
			log.Print("[DEBUG] Fetch AMI images")
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
			r.dataPrepared = true
		}

		var ami string
		err := runner.EvaluateExpr(attribute.Expr, &ami)

		return runner.EnsureNoError(err, func() error {
			if !r.amiIDs[ami] {
				runner.Issues = append(runner.Issues, &issue.Issue{
					Detector: r.Name(),
					Type:     issue.ERROR,
					Message:  fmt.Sprintf("\"%s\" is invalid AMI ID.", ami),
					Line:     attribute.Range.Start.Line,
					File:     runner.GetFileName(attribute.Range.Filename),
				})
			}
			return nil
		})
	})
}
