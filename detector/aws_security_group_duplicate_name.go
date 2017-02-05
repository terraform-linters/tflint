package detector

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/wata727/tflint/issue"
)

type AwsSecurityGroupDuplicateDetector struct {
	*Detector
}

func (d *Detector) CreateAwsSecurityGroupDuplicateDetector() *AwsSecurityGroupDuplicateDetector {
	return &AwsSecurityGroupDuplicateDetector{d}
}

func (d *AwsSecurityGroupDuplicateDetector) Detect(issues *[]*issue.Issue) {
	if !d.isDeepCheck("resource", "aws_security_group") {
		return
	}

	existSecuriyGroupNames := map[string]bool{}
	if d.ResponseCache.DescribeSecurityGroupsOutput == nil {
		resp, err := d.AwsClient.Ec2.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{})
		if err != nil {
			d.Logger.Error(err)
			d.Error = true
		}
		d.ResponseCache.DescribeSecurityGroupsOutput = resp
	}
	for _, securityGroup := range d.ResponseCache.DescribeSecurityGroupsOutput.SecurityGroups {
		existSecuriyGroupNames[*securityGroup.VpcId+"."+*securityGroup.GroupName] = true
	}

	for filename, list := range d.ListMap {
		for _, item := range list.Filter("resource", "aws_security_group").Items {
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
			var vpc string
			vpcToken, err := hclLiteralToken(item, "vpc_id")
			if err != nil {
				d.Logger.Error(err)
				// "vpc_id" is optional. If omitted, use default vpc_id.
				if d.ResponseCache.DescribeVpcsOutput == nil {
					resp, err := d.AwsClient.Ec2.DescribeVpcs(&ec2.DescribeVpcsInput{})
					if err != nil {
						d.Logger.Error(err)
						d.Error = true
					}
					d.ResponseCache.DescribeVpcsOutput = resp
				}
				for _, vpcResource := range d.ResponseCache.DescribeVpcsOutput.Vpcs {
					if *vpcResource.IsDefault {
						vpc = *vpcResource.VpcId
						break
					}
				}
			} else {
				vpc, err = d.evalToString(vpcToken.Text)
				if err != nil {
					d.Logger.Error(err)
					continue
				}
			}

			if existSecuriyGroupNames[vpc+"."+name] && !d.State.Exists("aws_security_group", hclObjectKeyText(item)) {
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
