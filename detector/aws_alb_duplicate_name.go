package detector

import (
	"fmt"

	"github.com/hashicorp/hcl/hcl/ast"
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

func (d *AwsALBDuplicateNameDetector) Detect(file string, item *ast.ObjectItem, issues *[]*issue.Issue) {
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
