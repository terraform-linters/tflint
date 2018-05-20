package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsDBInstancePreviousTypeDetector struct {
	*Detector
	previousInstanceTypes map[string]bool
}

func (d *Detector) CreateAwsDBInstancePreviousTypeDetector() *AwsDBInstancePreviousTypeDetector {
	nd := &AwsDBInstancePreviousTypeDetector{
		Detector:              d,
		previousInstanceTypes: map[string]bool{},
	}
	nd.Name = "aws_db_instance_previous_type"
	nd.IssueType = issue.WARNING
	nd.TargetType = "resource"
	nd.Target = "aws_db_instance"
	nd.DeepCheck = false
	nd.Link = "https://github.com/wata727/tflint/blob/master/docs/aws_db_instance_previous_type.md"
	nd.Enabled = true
	return nd
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

func (d *AwsDBInstancePreviousTypeDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	instanceTypeToken, ok := resource.GetToken("instance_class")
	if !ok {
		return
	}
	instanceType, err := d.evalToString(instanceTypeToken.Text)
	if err != nil {
		d.Logger.Error(err)
		return
	}

	if d.previousInstanceTypes[instanceType] {
		issue := &issue.Issue{
			Detector: d.Name,
			Type:     d.IssueType,
			Message:  fmt.Sprintf("\"%s\" is previous generation instance type.", instanceType),
			Line:     instanceTypeToken.Pos.Line,
			File:     instanceTypeToken.Pos.Filename,
			Link:     d.Link,
		}
		*issues = append(*issues, issue)
	}
}
