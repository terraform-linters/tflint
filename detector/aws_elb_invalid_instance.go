package detector

import (
	"fmt"

	"github.com/hashicorp/hcl/hcl/token"
	"github.com/wata727/tflint/issue"
)

type AwsELBInvalidInstanceDetector struct {
	*Detector
}

func (d *Detector) CreateAwsELBInvalidInstanceDetector() *AwsELBInvalidInstanceDetector {
	return &AwsELBInvalidInstanceDetector{d}
}

func (d *AwsELBInvalidInstanceDetector) Detect(issues *[]*issue.Issue) {
	if !d.isDeepCheck("resource", "aws_elb") {
		return
	}

	validInstances := map[string]bool{}
	resp, err := d.AwsClient.DescribeInstances()
	if err != nil {
		d.Logger.Error(err)
		d.Error = true
		return
	}
	for _, reservation := range resp.Reservations {
		for _, instance := range reservation.Instances {
			validInstances[*instance.InstanceId] = true
		}
	}

	for filename, list := range d.ListMap {
		for _, item := range list.Filter("resource", "aws_elb").Items {
			var varToken token.Token
			var instanceTokens []token.Token
			var err error
			if varToken, err = hclLiteralToken(item, "instances"); err == nil {
				instanceTokens, err = d.evalToStringTokens(varToken)
				if err != nil {
					d.Logger.Error(err)
					continue
				}
			} else {
				d.Logger.Error(err)
				instanceTokens, err = hclLiteralListToken(item, "instances")
				if err != nil {
					d.Logger.Error(err)
					continue
				}
			}

			for _, instanceToken := range instanceTokens {
				instance, err := d.evalToString(instanceToken.Text)
				if err != nil {
					d.Logger.Error(err)
					continue
				}

				if !validInstances[instance] {
					issue := &issue.Issue{
						Type:    "ERROR",
						Message: fmt.Sprintf("\"%s\" is invalid instance.", instance),
						Line:    instanceToken.Pos.Line,
						File:    filename,
					}
					*issues = append(*issues, issue)
				}
			}
		}
	}
}
