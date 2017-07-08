package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsInstancePreviousTypeDetector struct {
	*Detector
	previousInstanceTypes map[string]bool
}

func (d *Detector) CreateAwsInstancePreviousTypeDetector() *AwsInstancePreviousTypeDetector {
	nd := &AwsInstancePreviousTypeDetector{
		Detector:              d,
		previousInstanceTypes: map[string]bool{},
	}
	nd.Name = "aws_instance_previous_type"
	nd.IssueType = issue.WARNING
	nd.TargetType = "resource"
	nd.Target = "aws_instance"
	nd.DeepCheck = false
	nd.Link = "https://github.com/wata727/tflint/blob/master/docs/aws_instance_previous_type.md"
	return nd
}

func (d *AwsInstancePreviousTypeDetector) PreProcess() {
	d.previousInstanceTypes = map[string]bool{
		"t1.micro":    true,
		"m1.small":    true,
		"m1.medium":   true,
		"m1.large":    true,
		"m1.xlarge":   true,
		"c1.medium":   true,
		"c1.xlarge":   true,
		"cc2.8xlarge": true,
		"cg1.4xlarge": true,
		"m2.xlarge":   true,
		"m2.2xlarge":  true,
		"m2.4xlarge":  true,
		"cr1.8xlarge": true,
		"hi1.4xlarge": true,
		"hs1.8xlarge": true,
	}
}

func (d *AwsInstancePreviousTypeDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	instanceTypeToken, ok := resource.GetToken("instance_type")
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
			Type:    d.IssueType,
			Message: fmt.Sprintf("\"%s\" is previous generation instance type.", instanceType),
			Line:    instanceTypeToken.Pos.Line,
			File:    instanceTypeToken.Pos.Filename,
		}
		*issues = append(*issues, issue)
	}
}
