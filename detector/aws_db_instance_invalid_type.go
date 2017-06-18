package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsDBInstanceInvalidTypeDetector struct {
	*Detector
	IssueType     string
	Target        string
	DeepCheck     bool
	instanceTypes map[string]bool
}

func (d *Detector) CreateAwsDBInstanceInvalidTypeDetector() *AwsDBInstanceInvalidTypeDetector {
	return &AwsDBInstanceInvalidTypeDetector{
		Detector:      d,
		IssueType:     issue.ERROR,
		Target:        "aws_db_instance",
		DeepCheck:     false,
		instanceTypes: map[string]bool{},
	}
}

func (d *AwsDBInstanceInvalidTypeDetector) PreProcess() {
	d.instanceTypes = map[string]bool{
		"db.t2.micro":    true,
		"db.t2.small":    true,
		"db.t2.medium":   true,
		"db.t2.large":    true,
		"db.m4.large":    true,
		"db.m4.xlarge":   true,
		"db.m4.2xlarge":  true,
		"db.m4.4xlarge":  true,
		"db.m4.10xlarge": true,
		"db.m3.medium":   true,
		"db.m3.large":    true,
		"db.m3.xlarge":   true,
		"db.m3.2xlarge":  true,
		"db.r3.large":    true,
		"db.r3.xlarge":   true,
		"db.r3.2xlarge":  true,
		"db.r3.4xlarge":  true,
		"db.r3.8xlarge":  true,
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
			Type:    d.IssueType,
			Message: fmt.Sprintf("\"%s\" is invalid instance type.", instanceType),
			Line:    instanceTypeToken.Pos.Line,
			File:    instanceTypeToken.Pos.Filename,
		}
		*issues = append(*issues, issue)
	}
}
