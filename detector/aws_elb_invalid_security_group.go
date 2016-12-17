package detector

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/wata727/tflint/issue"
)

type AwsELBInvalidSecurityGroupDetector struct {
	*Detector
}

func (d *Detector) CreateAwsELBInvalidSecurityGroupDetector() *AwsELBInvalidSecurityGroupDetector {
	return &AwsELBInvalidSecurityGroupDetector{d}
}

func (d *AwsELBInvalidSecurityGroupDetector) Detect(issues *[]*issue.Issue) {
	if !d.isDeepCheck("resource", "aws_elb") {
		return
	}

	validSecurityGroups := map[string]bool{}
	if d.ResponseCache.DescribeSecurityGroupsOutput == nil {
		resp, err := d.AwsClient.Ec2.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{})
		if err != nil {
			d.Logger.Error(err)
			d.Error = true
		}
		d.ResponseCache.DescribeSecurityGroupsOutput = resp
	}
	for _, securityGroup := range d.ResponseCache.DescribeSecurityGroupsOutput.SecurityGroups {
		validSecurityGroups[*securityGroup.GroupId] = true
	}

	for filename, list := range d.ListMap {
		for _, item := range list.Filter("resource", "aws_elb").Items {
			securityGroupTokens, err := hclLiteralListToken(item, "security_groups")
			if err != nil {
				d.Logger.Error(err)
				continue
			}

			for _, securityGroupToken := range securityGroupTokens {
				securityGroup, err := d.evalToString(securityGroupToken.Text)
				if err != nil {
					d.Logger.Error(err)
					continue
				}

				if !validSecurityGroups[securityGroup] {
					issue := &issue.Issue{
						Type:    "ERROR",
						Message: fmt.Sprintf("\"%s\" is invalid security group.", securityGroup),
						Line:    securityGroupToken.Pos.Line,
						File:    filename,
					}
					*issues = append(*issues, issue)
				}
			}
		}
	}
}
