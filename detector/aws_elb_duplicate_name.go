package detector

import (
	"fmt"

	"github.com/hashicorp/hcl/hcl/ast"
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

func (d *AwsELBDuplicateNameDetector) Detect(file string, item *ast.ObjectItem, issues *[]*issue.Issue) {
	nameToken, err := hclLiteralToken(item, "name")
	if err != nil {
		d.Logger.Error(err)
		return
	}
	name, err := d.evalToString(nameToken.Text)
	if err != nil {
		d.Logger.Error(err)
		return
	}

	if d.loadBalancers[name] && !d.State.Exists(d.Target, hclObjectKeyText(item)) {
		issue := &issue.Issue{
			Type:    d.IssueType,
			Message: fmt.Sprintf("\"%s\" is duplicate name. It must be unique.", name),
			Line:    nameToken.Pos.Line,
			File:    file,
		}
		*issues = append(*issues, issue)
	}
}
