package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsALBDuplicateNameDetector struct {
	*Detector
	loadBalancers map[string]bool
}

func (d *Detector) CreateAwsALBDuplicateNameDetector() *AwsALBDuplicateNameDetector {
	nd := &AwsALBDuplicateNameDetector{
		Detector:      d,
		loadBalancers: map[string]bool{},
	}
	nd.Name = "aws_alb_duplicate_name"
	nd.IssueType = issue.ERROR
	nd.TargetType = "resource"
	nd.Target = "aws_alb"
	nd.DeepCheck = true
	return nd
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

func (d *AwsALBDuplicateNameDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
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
