package detector

import (
	"fmt"

	"github.com/hashicorp/hcl/hcl/token"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsELBInvalidInstanceDetector struct {
	*Detector
	IssueType  string
	TargetType string
	Target     string
	DeepCheck  bool
	instances  map[string]bool
}

func (d *Detector) CreateAwsELBInvalidInstanceDetector() *AwsELBInvalidInstanceDetector {
	return &AwsELBInvalidInstanceDetector{
		Detector:   d,
		IssueType:  issue.ERROR,
		TargetType: "resource",
		Target:     "aws_elb",
		DeepCheck:  true,
		instances:  map[string]bool{},
	}
}

func (d *AwsELBInvalidInstanceDetector) PreProcess() {
	resp, err := d.AwsClient.DescribeInstances()
	if err != nil {
		d.Logger.Error(err)
		d.Error = true
		return
	}

	for _, reservation := range resp.Reservations {
		for _, instance := range reservation.Instances {
			d.instances[*instance.InstanceId] = true
		}
	}
}

func (d *AwsELBInvalidInstanceDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	var varToken token.Token
	var instanceTokens []token.Token
	var ok bool
	if varToken, ok = resource.GetToken("instances"); ok {
		var err error
		instanceTokens, err = d.evalToStringTokens(varToken)
		if err != nil {
			d.Logger.Error(err)
			return
		}
	} else {
		instanceTokens, ok = resource.GetListToken("instances")
		if !ok {
			return
		}
	}

	for _, instanceToken := range instanceTokens {
		instance, err := d.evalToString(instanceToken.Text)
		if err != nil {
			d.Logger.Error(err)
			continue
		}

		// If `instances` is interpolated by list variable, file name is empty.
		if instanceToken.Pos.Filename == "" {
			instanceToken.Pos.Filename = varToken.Pos.Filename
		}
		if !d.instances[instance] {
			issue := &issue.Issue{
				Type:    d.IssueType,
				Message: fmt.Sprintf("\"%s\" is invalid instance.", instance),
				Line:    instanceToken.Pos.Line,
				File:    instanceToken.Pos.Filename,
			}
			*issues = append(*issues, issue)
		}
	}
}
