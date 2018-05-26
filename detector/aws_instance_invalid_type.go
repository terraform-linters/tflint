package detector

import (
	"fmt"

	instances "github.com/cristim/ec2-instances-info"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsInstanceInvalidTypeDetector struct {
	*Detector
	instanceTypes map[string]bool
}

func (d *Detector) CreateAwsInstanceInvalidTypeDetector() *AwsInstanceInvalidTypeDetector {
	nd := &AwsInstanceInvalidTypeDetector{
		Detector:      d,
		instanceTypes: map[string]bool{},
	}
	nd.Name = "aws_instance_invalid_type"
	nd.IssueType = issue.ERROR
	nd.TargetType = "resource"
	nd.Target = "aws_instance"
	nd.DeepCheck = false
	nd.Link = "https://github.com/wata727/tflint/blob/master/docs/aws_instance_invalid_type.md"
	nd.Enabled = true
	return nd
}

func (d *AwsInstanceInvalidTypeDetector) PreProcess() {
	data, err := instances.Data()
	if err != nil {
		d.Logger.Error(err)
		d.Error = true
		return
	}

	for _, i := range *data {
		d.instanceTypes[i.InstanceType] = true
	}
}

func (d *AwsInstanceInvalidTypeDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	instanceTypeToken, ok := resource.GetToken("instance_type")
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
