package detector

import (
	"fmt"

	"github.com/hashicorp/hcl/hcl/ast"
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
	resp, err := d.AwsClient.DescribeSecurityGroups()
	if err != nil {
		d.Logger.Error(err)
		d.Error = true
		return
	}
	for _, securityGroup := range resp.SecurityGroups {
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
			vpc, err = d.fetchVpcId(item)
			if err != nil {
				d.Logger.Error(err)
				continue
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

func (d *AwsSecurityGroupDuplicateDetector) fetchVpcId(item *ast.ObjectItem) (string, error) {
	var vpc string
	vpcToken, err := hclLiteralToken(item, "vpc_id")
	if err != nil {
		d.Logger.Error(err)
		// "vpc_id" is optional. If omitted, use default vpc_id.
		resp, err := d.AwsClient.DescribeVpcs()
		if err != nil {
			return "", err
		}
		for _, vpcResource := range resp.Vpcs {
			if *vpcResource.IsDefault {
				vpc = *vpcResource.VpcId
				break
			}
		}
	} else {
		vpc, err = d.evalToString(vpcToken.Text)
		if err != nil {
			return "", err
		}
	}

	return vpc, nil
}
