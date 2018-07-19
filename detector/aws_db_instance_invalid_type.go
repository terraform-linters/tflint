package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsDBInstanceInvalidTypeDetector struct {
	*Detector
	instanceTypes map[string]bool
}

func (d *Detector) CreateAwsDBInstanceInvalidTypeDetector() *AwsDBInstanceInvalidTypeDetector {
	nd := &AwsDBInstanceInvalidTypeDetector{
		Detector:      d,
		instanceTypes: map[string]bool{},
	}
	nd.Name = "aws_db_instance_invalid_type"
	nd.IssueType = issue.ERROR
	nd.TargetType = "resource"
	nd.Target = "aws_db_instance"
	nd.DeepCheck = false
	nd.Link = "https://github.com/wata727/tflint/blob/master/docs/aws_db_instance_invalid_type.md"
	nd.Enabled = true
	return nd
}

func (d *AwsDBInstanceInvalidTypeDetector) PreProcess() {
	d.instanceTypes = map[string]bool{
		"db.t2.micro":     true,
		"db.t2.small":     true,
		"db.t2.medium":    true,
		"db.t2.large":     true,
		"db.t2.xlarge":    true,
		"db.t2.2xlarge":   true,
		"db.m4.large":     true,
		"db.m4.xlarge":    true,
		"db.m4.2xlarge":   true,
		"db.m4.4xlarge":   true,
		"db.m4.10xlarge":  true,
		"db.m4.16xlarge":  true,
		"db.m3.medium":    true,
		"db.m3.large":     true,
		"db.m3.xlarge":    true,
		"db.m3.2xlarge":   true,
		"db.r4.large":     true,
		"db.r4.xlarge":    true,
		"db.r4.2xlarge":   true,
		"db.r4.4xlarge":   true,
		"db.r4.8xlarge":   true,
		"db.r4.16xlarge":  true,
		"db.r3.large":     true,
		"db.r3.xlarge":    true,
		"db.r3.2xlarge":   true,
		"db.r3.4xlarge":   true,
		"db.r3.8xlarge":   true,
		"db.t1.micro":     true,
		"db.m1.small":     true,
		"db.m1.medium":    true,
		"db.m1.large":     true,
		"db.m1.xlarge":    true,
		"db.m2.xlarge":    true,
		"db.m2.2xlarge":   true,
		"db.m2.4xlarge":   true,
		"db.cr1.8xlarge":  true,
		"db.x1.16xlarge":  true,
		"db.x1.32xlarge":  true,
		"db.x1e.xlarge":   true,
		"db.x1e.2xlarge":  true,
		"db.x1e.4xlarge":  true,
		"db.x1e.8xlarge":  true,
		"db.x1e.16xlarge": true,
		"db.x1e.32xlarge": true,
	}
}

func (d *AwsDBInstanceInvalidTypeDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	instanceTypeToken, ok := resource.GetToken("instance_class")
	if !ok {
		return
	}
	instanceType, err := d.evalToString(instanceTypeToken.Text)
	if err != nil {
		d.Logger.Error(err)
		return
	}

	if !d.instanceTypes[instanceType] {
		issue := &issue.Issue{
			Detector: d.Name,
			Type:     d.IssueType,
			Message:  fmt.Sprintf("\"%s\" is invalid instance type.", instanceType),
			Line:     instanceTypeToken.Pos.Line,
			File:     instanceTypeToken.Pos.Filename,
			Link:     d.Link,
		}
		*issues = append(*issues, issue)
	}
}
