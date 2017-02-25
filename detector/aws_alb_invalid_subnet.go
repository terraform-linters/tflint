package detector

import (
	"fmt"

	"github.com/hashicorp/hcl/hcl/token"
	"github.com/wata727/tflint/issue"
)

type AwsALBInvalidSubnetDetector struct {
	*Detector
	IssueType string
	Target    string
	DeepCheck bool
	subnets   map[string]bool
}

func (d *Detector) CreateAwsALBInvalidSubnetDetector() *AwsALBInvalidSubnetDetector {
	return &AwsALBInvalidSubnetDetector{
		Detector:  d,
		IssueType: issue.ERROR,
		Target:    "aws_alb",
		DeepCheck: true,
		subnets:   map[string]bool{},
	}
}

func (d *AwsALBInvalidSubnetDetector) PreProcess() {
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

func (d *AwsALBInvalidSubnetDetector) Detect(issues *[]*issue.Issue) {
	for filename, list := range d.ListMap {
		for _, item := range list.Filter("resource", "aws_alb").Items {
			var varToken token.Token
			var subnetTokens []token.Token
			var err error
			if varToken, err = hclLiteralToken(item, "subnets"); err == nil {
				subnetTokens, err = d.evalToStringTokens(varToken)
				if err != nil {
					d.Logger.Error(err)
					continue
				}
			} else {
				d.Logger.Error(err)
				subnetTokens, err = hclLiteralListToken(item, "subnets")
				if err != nil {
					d.Logger.Error(err)
					continue
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
						Type:    "ERROR",
						Message: fmt.Sprintf("\"%s\" is invalid subnet ID.", subnet),
						Line:    subnetToken.Pos.Line,
						File:    filename,
					}
					*issues = append(*issues, issue)
				}
			}
		}
	}
}
