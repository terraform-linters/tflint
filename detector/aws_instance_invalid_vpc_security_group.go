package detector

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/wata727/tflint/issue"
)

type AwsInstanceInvalidVPCSecurityGroupDetector struct {
	*Detector
}

func (d *Detector) CreateAwsInstanceInvalidVPCSecurityGroupDetector() *AwsInstanceInvalidVPCSecurityGroupDetector {
	return &AwsInstanceInvalidVPCSecurityGroupDetector{d}
}

func (d *AwsInstanceInvalidVPCSecurityGroupDetector) Detect(issues *[]*issue.Issue) {
	if !d.Config.DeepCheck {
		d.Logger.Info("skip this rule.")
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
		for _, item := range list.Filter("resource", "aws_instance").Items {
			securityGroupTokens, err := hclLiteralListToken(item, "vpc_security_group_ids")
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
