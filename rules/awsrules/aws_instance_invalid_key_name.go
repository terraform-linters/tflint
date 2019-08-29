package awsrules

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/tflint"
)

// AwsInstanceInvalidKeyNameRule checks whether key pair actually exists
type AwsInstanceInvalidKeyNameRule struct {
	resourceType  string
	attributeName string
	keypairs      map[string]bool
	dataPrepared  bool
}

// NewAwsInstanceInvalidKeyNameRule returns new rule with default attributes
func NewAwsInstanceInvalidKeyNameRule() *AwsInstanceInvalidKeyNameRule {
	return &AwsInstanceInvalidKeyNameRule{
		resourceType:  "aws_instance",
		attributeName: "key_name",
		keypairs:      map[string]bool{},
		dataPrepared:  false,
	}
}

// Name returns the rule name
func (r *AwsInstanceInvalidKeyNameRule) Name() string {
	return "aws_instance_invalid_key_name"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsInstanceInvalidKeyNameRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *AwsInstanceInvalidKeyNameRule) Severity() string {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *AwsInstanceInvalidKeyNameRule) Link() string {
	return ""
}

// Check checks whether `key_name` are included in the list retrieved by `DescribeKeyPairs`
func (r *AwsInstanceInvalidKeyNameRule) Check(runner *tflint.Runner) error {
	log.Printf("[INFO] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		if !r.dataPrepared {
			log.Print("[DEBUG] Fetch key pairs")
			resp, err := runner.AwsClient.EC2.DescribeKeyPairs(&ec2.DescribeKeyPairsInput{})
			if err != nil {
				err := &tflint.Error{
					Code:    tflint.ExternalAPIError,
					Level:   tflint.ErrorLevel,
					Message: "An error occurred while describing key pairs",
					Cause:   err,
				}
				log.Printf("[ERROR] %s", err)
				return err
			}
			for _, keyPair := range resp.KeyPairs {
				r.keypairs[*keyPair.KeyName] = true
			}
			r.dataPrepared = true
		}

		var key string
		err := runner.EvaluateExpr(attribute.Expr, &key)

		return runner.EnsureNoError(err, func() error {
			if !r.keypairs[key] {
				runner.EmitIssue(
					r,
					fmt.Sprintf("\"%s\" is invalid key name.", key),
					attribute.Expr.Range(),
				)
			}
			return nil
		})
	})
}
