package awsrules

import (
	"log"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/lang"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

// AwsDBInstanceReadablePasswordRule checks whether "aws_db_instance" has readable password field
type AwsDBInstanceReadablePasswordRule struct {
	resourceType  string
	attributeName string
}

// NewAwsDBInstanceReadablePasswordRule returns new rule with default attributes
func NewAwsDBInstanceReadablePasswordRule() *AwsDBInstanceReadablePasswordRule {
	return &AwsDBInstanceReadablePasswordRule{
		resourceType:  "aws_db_instance",
		attributeName: "password",
	}
}

// Name returns the rule name
func (r *AwsDBInstanceReadablePasswordRule) Name() string {
	return "aws_db_instance_readable_password"
}

// Enabled returns whether the rule is enabled by default
func (r *AwsDBInstanceReadablePasswordRule) Enabled() bool {
	return true
}

// Check checks password
func (r *AwsDBInstanceReadablePasswordRule) Check(runner *tflint.Runner) error {
	log.Printf("[INFO] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	return runner.WalkResourceAttributes(r.resourceType, r.attributeName, func(attribute *hcl.Attribute) error {
		refs, diags := lang.ReferencesInExpr(attribute.Expr)
		if diags.HasErrors() {
			// Maybe this is bug
			panic(diags.Err())
		}

		varSubjects := []addrs.InputVariable{}
		readableSubjects := []addrs.InputVariable{}
		for _, ref := range refs {
			if sub, ok := ref.Subject.(addrs.InputVariable); ok {
				varSubjects = append(varSubjects, sub)

				variable := runner.TFConfig.Module.Variables[sub.Name]
				if variable == nil {
					continue
				}
				if !variable.Default.IsNull() {
					readableSubjects = append(readableSubjects, sub)
				}
			}
		}

		if len(varSubjects) == 0 || len(varSubjects) == len(readableSubjects) {
			runner.Issues = append(runner.Issues, &issue.Issue{
				Detector: r.Name(),
				Type:     issue.WARNING,
				Message:  "Password for the master DB user is readable. Recommend using environment variables or variable files.",
				Line:     attribute.Range.Start.Line,
				File:     runner.GetFileName(attribute.Range.Filename),
				Link:     "https://github.com/wata727/tflint/blob/master/docs/aws_db_instance_readable_password.md",
			})
		}

		return nil
	})
}
