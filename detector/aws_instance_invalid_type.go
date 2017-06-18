package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsInstanceInvalidTypeDetector struct {
	*Detector
	IssueType     string
	Target        string
	DeepCheck     bool
	instanceTypes map[string]bool
}

func (d *Detector) CreateAwsInstanceInvalidTypeDetector() *AwsInstanceInvalidTypeDetector {
	return &AwsInstanceInvalidTypeDetector{
		Detector:      d,
		IssueType:     issue.ERROR,
		Target:        "aws_instance",
		DeepCheck:     false,
		instanceTypes: map[string]bool{},
	}
}

func (d *AwsInstanceInvalidTypeDetector) PreProcess() {
	d.instanceTypes = map[string]bool{
		"t2.nano":     true,
		"t2.micro":    true,
		"t2.small":    true,
		"t2.medium":   true,
		"t2.large":    true,
		"t2.xlarge":   true,
		"t2.2xlarge":  true,
		"m4.large":    true,
		"m4.xlarge":   true,
		"m4.2xlarge":  true,
		"m4.4xlarge":  true,
		"m4.10xlarge": true,
		"m4.16xlarge": true,
		"m3.medium":   true,
		"m3.large":    true,
		"m3.xlarge":   true,
		"m3.2xlarge":  true,
		"c4.large":    true,
		"c4.2xlarge":  true,
		"c4.4xlarge":  true,
		"c4.8xlarge":  true,
		"c3.large":    true,
		"c3.xlarge":   true,
		"c3.2xlarge":  true,
		"c3.4xlarge":  true,
		"c3.8xlarge":  true,
		"x1.16xlarge": true,
		"x1.32xlarge": true,
		"r4.large":    true,
		"r4.xlarge":   true,
		"r4.2xlarge":  true,
		"r4.4xlarge":  true,
		"r4.8xlarge":  true,
		"r4.16xlarge": true,
		"r3.large":    true,
		"r3.xlarge":   true,
		"r3.2xlarge":  true,
		"r3.4xlarge":  true,
		"r3.8xlarge":  true,
		"p2.xlarge":   true,
		"p2.8xlarge":  true,
		"p2.16xlarge": true,
		"g2.2xlarge":  true,
		"g2.8xlarge":  true,
		"i2.xlarge":   true,
		"i2.2xlarge":  true,
		"i2.4xlarge":  true,
		"i2.8xlarge":  true,
		"d2.xlarge":   true,
		"d2.2xlarge":  true,
		"d2.4xlarge":  true,
		"d2.8xlarge":  true,
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
		"i3.large":    true,
		"i3.xlarge":   true,
		"i3.2xlarge":  true,
		"i3.4xlarge":  true,
		"i3.8xlarge":  true,
		"i3.16xlarge": true,
		"f1.2xlarge":  true,
		"f1.16xlarge": true,
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
			Type:    d.IssueType,
			Message: fmt.Sprintf("\"%s\" is invalid instance type.", instanceType),
			Line:    instanceTypeToken.Pos.Line,
			File:    instanceTypeToken.Pos.Filename,
		}
		*issues = append(*issues, issue)
	}
}
