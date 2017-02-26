package detector

import (
	"fmt"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/token"
	"github.com/wata727/tflint/issue"
)

type AwsELBInvalidInstanceDetector struct {
	*Detector
	IssueType string
	Target    string
	DeepCheck bool
	instances map[string]bool
}

func (d *Detector) CreateAwsELBInvalidInstanceDetector() *AwsELBInvalidInstanceDetector {
	return &AwsELBInvalidInstanceDetector{
		Detector:  d,
		IssueType: issue.ERROR,
		Target:    "aws_elb",
		DeepCheck: true,
		instances: map[string]bool{},
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

func (d *AwsELBInvalidInstanceDetector) Detect(file string, item *ast.ObjectItem, issues *[]*issue.Issue) {
	var varToken token.Token
	var instanceTokens []token.Token
	var err error
	if varToken, err = hclLiteralToken(item, "instances"); err == nil {
		instanceTokens, err = d.evalToStringTokens(varToken)
		if err != nil {
			d.Logger.Error(err)
			return
		}
	} else {
		d.Logger.Error(err)
		instanceTokens, err = hclLiteralListToken(item, "instances")
		if err != nil {
			d.Logger.Error(err)
			return
		}
	}

	for _, instanceToken := range instanceTokens {
		instance, err := d.evalToString(instanceToken.Text)
		if err != nil {
			d.Logger.Error(err)
			continue
		}

		if !d.instances[instance] {
			issue := &issue.Issue{
				Type:    d.IssueType,
				Message: fmt.Sprintf("\"%s\" is invalid instance.", instance),
				Line:    instanceToken.Pos.Line,
				File:    file,
			}
			*issues = append(*issues, issue)
		}
	}
}
