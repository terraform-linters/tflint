package detector

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/wata727/tflint/issue"
)

type AwsALBDuplicateNameDetector struct {
	*Detector
}

func (d *Detector) CreateAwsALBDuplicateNameDetector() *AwsALBDuplicateNameDetector {
	return &AwsALBDuplicateNameDetector{d}
}

func (d *AwsALBDuplicateNameDetector) Detect(issues *[]*issue.Issue) {
	if !d.isDeepCheck("resource", "aws_alb") {
		return
	}

	existLoadBalancerNames := map[string]bool{}
	if d.ResponseCache.DescribeLoadBalancersOutput == nil {
		resp, err := d.AwsClient.Elbv2.DescribeLoadBalancers(&elbv2.DescribeLoadBalancersInput{})
		if err != nil {
			d.Logger.Error(err)
			d.Error = true
		}
		d.ResponseCache.DescribeLoadBalancersOutput = resp
	}
	for _, loadBalancer := range d.ResponseCache.DescribeLoadBalancersOutput.LoadBalancers {
		existLoadBalancerNames[*loadBalancer.LoadBalancerName] = true
	}

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

			if existLoadBalancerNames[name] && !d.State.Exists("aws_alb", hclObjectKeyText(item)) {
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
