package awsrules

import (
	"fmt"
	"path/filepath"

	instances "github.com/cristim/ec2-instances-info"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

// AwsInstanceInvalidTypeRule checks whether "aws_instance" has invalid intance type.
type AwsInstanceInvalidTypeRule struct {
	instanceTypes map[string]bool
}

// PreProcess makes valid instance type list.
func (r *AwsInstanceInvalidTypeRule) PreProcess() {
	r.instanceTypes = map[string]bool{}

	data, err := instances.Data()
	if err != nil {
		// TODO
		panic(err)
	}

	for _, i := range *data {
		r.instanceTypes[i.InstanceType] = true
	}
}

// Check checks "aws_instance" according to the rule.
func (r *AwsInstanceInvalidTypeRule) Check(runner *tflint.Runner) {
	for _, resource := range runner.TFConfig.Module.ManagedResources {
		if resource.Type != "aws_instance" {
			continue
		}

		body, _, diags := resource.Config.PartialContent(&hcl.BodySchema{
			Attributes: []hcl.AttributeSchema{
				{
					Name: "instance_type",
				},
			},
		})
		if diags.HasErrors() {
			// TODO
			panic(diags)
		}

		if attribute, ok := body.Attributes["instance_type"]; ok {
			var instanceType string
			err := runner.EvaluateExpr(attribute.Expr, &instanceType)
			if err != nil {
				// TODO
				panic(err)
			}

			if !r.instanceTypes[instanceType] {
				runner.Issues = append(runner.Issues, &issue.Issue{
					Detector: "aws_instance_invalid_type",
					Type:     issue.ERROR,
					Message:  fmt.Sprintf("\"%s\" is invalid instance type.", instanceType),
					Line:     attribute.Range.Start.Line,
					File:     filepath.Base(attribute.Range.Filename),
					Link:     "https://github.com/wata727/tflint/blob/master/docs/aws_instance_invalid_type.md",
				})
			}
		}
	}
}
