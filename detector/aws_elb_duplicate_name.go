package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsELBDuplicateNameDetector struct {
	*Detector
	loadBalancers map[string]bool
}

func (d *Detector) CreateAwsELBDuplicateNameDetector() *AwsELBDuplicateNameDetector {
	nd := &AwsELBDuplicateNameDetector{
		Detector:      d,
		loadBalancers: map[string]bool{},
	}
	nd.Name = "aws_elb_duplicate_name"
	nd.IssueType = issue.ERROR
	nd.TargetType = "resource"
	nd.Target = "aws_elb"
	nd.DeepCheck = true
	return nd
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
