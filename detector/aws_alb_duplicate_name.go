package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
)

type AwsALBDuplicateNameDetector struct {
	*Detector
	IssueType     string
	Target        string
	DeepCheck     bool
	loadBalancers map[string]bool
}

func (d *Detector) CreateAwsALBDuplicateNameDetector() *AwsALBDuplicateNameDetector {
	return &AwsALBDuplicateNameDetector{
		Detector:      d,
		IssueType:     issue.ERROR,
		Target:        "aws_alb",
		DeepCheck:     true,
		loadBalancers: map[string]bool{},
	}
}

func (d *AwsALBDuplicateNameDetector) PreProcess() {
	resp, err := d.AwsClient.DescribeLoadBalancers()
	if err != nil {
		d.Logger.Error(err)
		d.Error = true
		return
	}

	for _, loadBalancer := range resp.LoadBalancers {
		d.loadBalancers[*loadBalancer.LoadBalancerName] = true
	}
}

func (d *AwsALBDuplicateNameDetector) Detect(issues *[]*issue.Issue) {
	for filename, list := range d.ListMap {
		for _, item := range list.Filter("resource", "aws_alb").Items {
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

			if d.loadBalancers[name] && !d.State.Exists("aws_alb", hclObjectKeyText(item)) {
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
