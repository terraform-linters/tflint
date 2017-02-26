package detector

import (
	"fmt"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/token"
	"github.com/wata727/tflint/issue"
)

type AwsELBInvalidSubnetDetector struct {
	*Detector
	IssueType string
	Target    string
	DeepCheck bool
	subnets   map[string]bool
}

func (d *Detector) CreateAwsELBInvalidSubnetDetector() *AwsELBInvalidSubnetDetector {
	return &AwsELBInvalidSubnetDetector{
		Detector:  d,
		IssueType: issue.ERROR,
		Target:    "aws_elb",
		DeepCheck: true,
		subnets:   map[string]bool{},
	}
}

func (d *AwsELBInvalidSubnetDetector) PreProcess() {
	resp, err := d.AwsClient.DescribeSubnets()
	if err != nil {
		d.Logger.Error(err)
		d.Error = true
		return
	}

	for _, subnet := range resp.Subnets {
		d.subnets[*subnet.SubnetId] = true
	}
}

func (d *AwsELBInvalidSubnetDetector) Detect(file string, item *ast.ObjectItem, issues *[]*issue.Issue) {
	var varToken token.Token
	var subnetTokens []token.Token
	var err error
	if varToken, err = hclLiteralToken(item, "subnets"); err == nil {
		subnetTokens, err = d.evalToStringTokens(varToken)
		if err != nil {
			d.Logger.Error(err)
			return
		}
	} else {
		d.Logger.Error(err)
		subnetTokens, err = hclLiteralListToken(item, "subnets")
		if err != nil {
			d.Logger.Error(err)
			return
		}
	}

	for _, subnetToken := range subnetTokens {
		subnet, err := d.evalToString(subnetToken.Text)
		if err != nil {
			d.Logger.Error(err)
			continue
		}

		if !d.subnets[subnet] {
			issue := &issue.Issue{
				Type:    d.IssueType,
				Message: fmt.Sprintf("\"%s\" is invalid subnet ID.", subnet),
				Line:    subnetToken.Pos.Line,
				File:    file,
			}
			*issues = append(*issues, issue)
		}
	}
}
