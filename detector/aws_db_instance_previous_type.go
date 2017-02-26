package detector

import (
	"fmt"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/wata727/tflint/issue"
)

type AwsDBInstancePreviousTypeDetector struct {
	*Detector
	IssueType             string
	Target                string
	DeepCheck             bool
	previousInstanceTypes map[string]bool
}

func (d *Detector) CreateAwsDBInstancePreviousTypeDetector() *AwsDBInstancePreviousTypeDetector {
	return &AwsDBInstancePreviousTypeDetector{
		Detector:              d,
		IssueType:             issue.WARNING,
		Target:                "aws_db_instance",
		DeepCheck:             false,
		previousInstanceTypes: map[string]bool{},
	}
}

func (d *AwsDBInstancePreviousTypeDetector) PreProcess() {
	d.previousInstanceTypes = map[string]bool{
		"db.t1.micro":    true,
		"db.m1.small":    true,
		"db.m1.medium":   true,
		"db.m1.large":    true,
		"db.m1.xlarge":   true,
		"db.m2.xlarge":   true,
		"db.m2.2xlarge":  true,
		"db.m2.4xlarge":  true,
		"db.cr1.8xlarge": true,
	}
}

func (d *AwsDBInstancePreviousTypeDetector) Detect(file string, item *ast.ObjectItem, issues *[]*issue.Issue) {
	instanceTypeToken, err := hclLiteralToken(item, "instance_class")
	if err != nil {
		d.Logger.Error(err)
		return
	}
	instanceType, err := d.evalToString(instanceTypeToken.Text)
	if err != nil {
		d.Logger.Error(err)
		return
	}

	if d.previousInstanceTypes[instanceType] {
		issue := &issue.Issue{
			Type:    d.IssueType,
			Message: fmt.Sprintf("\"%s\" is previous generation instance type.", instanceType),
			Line:    instanceTypeToken.Pos.Line,
			File:    file,
		}
		*issues = append(*issues, issue)
	}
}
