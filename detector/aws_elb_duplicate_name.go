package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
)

type AwsELBDuplicateNameDetector struct {
	*Detector
	IssueType     string
	Target        string
	DeepCheck     bool
	loadBalancers map[string]bool
}

func (d *Detector) CreateAwsELBDuplicateNameDetector() *AwsELBDuplicateNameDetector {
	return &AwsELBDuplicateNameDetector{
		Detector:      d,
		IssueType:     issue.ERROR,
		Target:        "aws_elb",
		DeepCheck:     true,
		loadBalancers: map[string]bool{},
	}
}

func (d *AwsELBDuplicateNameDetector) PreProcess() {
	resp, err := d.AwsClient.DescribeClassicLoadBalancers()
	if err != nil {
		d.Logger.Error(err)
		d.Error = true
		return
	}

	for _, loadBalancer := range resp.LoadBalancerDescriptions {
		d.loadBalancers[*loadBalancer.LoadBalancerName] = true
	}
}

func (d *AwsELBDuplicateNameDetector) Detect(issues *[]*issue.Issue) {
	for filename, list := range d.ListMap {
		for _, item := range list.Filter("resource", "aws_elb").Items {
			nameToken, err := hclLiteralToken(item, "name")
			if err != nil {
				d.Logger.Error(err)
				continue
			}
			name, err := d.evalToString(nameToken.Text)
			if err != nil {
				d.Logger.Error(err)
				continue
			}

			if d.loadBalancers[name] && !d.State.Exists("aws_elb", hclObjectKeyText(item)) {
				issue := &issue.Issue{
					Type:    "ERROR",
					Message: fmt.Sprintf("\"%s\" is duplicate name. It must be unique.", name),
					Line:    nameToken.Pos.Line,
					File:    filename,
				}
				*issues = append(*issues, issue)
			}
		}
	}
}
