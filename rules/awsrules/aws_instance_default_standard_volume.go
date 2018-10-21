package awsrules

import (
	"log"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

// AwsInstanceDefaultStandardVolumeRule checks whether the volume type is unspecified
type AwsInstanceDefaultStandardVolumeRule struct {
	resourceType string
}

// NewAwsInstanceDefaultStandardVolumeRule returns new rule with default attributes
func NewAwsInstanceDefaultStandardVolumeRule() *AwsInstanceDefaultStandardVolumeRule {
	return &AwsInstanceDefaultStandardVolumeRule{
		resourceType: "aws_instance",
	}
}

// Name returns the rule name
func (r *AwsInstanceDefaultStandardVolumeRule) Name() string {
	return "aws_instance_default_standard_volume"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsInstanceDefaultStandardVolumeRule) Enabled() bool {
	return true
}

// Type returns the rule severity
func (r *AwsInstanceDefaultStandardVolumeRule) Type() string {
	return issue.WARNING
}

// Link returns the rule reference link
func (r *AwsInstanceDefaultStandardVolumeRule) Link() string {
	return "https://github.com/wata727/tflint/blob/master/docs/aws_instance_default_standard_volume.md"
}

// Check checks whether `volume_type` is defined for `root_block_device` or `ebs_block_device`
func (r *AwsInstanceDefaultStandardVolumeRule) Check(runner *tflint.Runner) error {
	log.Printf("[INFO] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	if err := runner.WalkResourceBlocks(r.resourceType, "root_block_device", func(block *hcl.Block) error {
		return r.blockWalker(runner, block)
	}); err != nil {
		return err
	}

	if err := runner.WalkResourceBlocks(r.resourceType, "ebs_block_device", func(block *hcl.Block) error {
		return r.blockWalker(runner, block)
	}); err != nil {
		return err
	}

	// Since Terraform v0.12, block device definitions must be defined as block, but it walks attributes for the backward compatibility.
	if err := runner.WalkResourceAttributes(r.resourceType, "root_block_device", func(attribute *hcl.Attribute) error {
		return r.attributeWalker(runner, attribute)
	}); err != nil {
		return err
	}

	if err := runner.WalkResourceAttributes(r.resourceType, "ebs_block_device", func(attribute *hcl.Attribute) error {
		return r.attributeWalker(runner, attribute)
	}); err != nil {
		return err
	}

	return nil
}

func (r *AwsInstanceDefaultStandardVolumeRule) message() string {
	return "\"volume_type\" is not specified. Default standard volume type is not recommended. You can use \"gp2\", \"io1\", etc instead."
}

func (r *AwsInstanceDefaultStandardVolumeRule) blockWalker(runner *tflint.Runner, block *hcl.Block) error {
	body, _, diags := block.Body.PartialContent(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name: "volume_type",
			},
		},
	})
	if diags.HasErrors() {
		return diags
	}

	if _, ok := body.Attributes["volume_type"]; !ok {
		runner.EmitIssue(r, r.message(), block.TypeRange)
	}
	return nil
}

func (r *AwsInstanceDefaultStandardVolumeRule) attributeWalker(runner *tflint.Runner, attribute *hcl.Attribute) error {
	var val map[string]string
	err := runner.EvaluateExpr(attribute.Expr, &val)

	return runner.EnsureNoError(err, func() error {
		if _, ok := val["volume_type"]; !ok {
			runner.EmitIssue(r, r.message(), attribute.Range)
		}
		return nil
	})
}
