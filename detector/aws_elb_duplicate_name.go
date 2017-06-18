package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
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

func (d *AwsELBDuplicateNameDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	nameToken, ok := resource.GetToken("name")
	if !ok {
		return
	}
	name, err := d.evalToString(nameToken.Text)
	if err != nil {
		d.Logger.Error(err)
		return
	}

	if d.loadBalancers[name] && !d.State.Exists(d.Target, resource.Id) {
		issue := &issue.Issue{
			Type:    d.IssueType,
			Message: fmt.Sprintf("\"%s\" is duplicate name. It must be unique.", name),
			Line:    nameToken.Pos.Line,
			File:    nameToken.Pos.Filename,
		}
		*issues = append(*issues, issue)
	}
}
