package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsSecurityGroupDuplicateDetector struct {
	*Detector
	IssueType     string
	TargetType    string
	Target        string
	DeepCheck     bool
	securiyGroups map[string]bool
	defaultVpc    string
}

func (d *Detector) CreateAwsSecurityGroupDuplicateDetector() *AwsSecurityGroupDuplicateDetector {
	return &AwsSecurityGroupDuplicateDetector{
		Detector:      d,
		IssueType:     issue.ERROR,
		TargetType:    "resource",
		Target:        "aws_security_group",
		DeepCheck:     true,
		securiyGroups: map[string]bool{},
		defaultVpc:    "",
	}
}

func (d *AwsSecurityGroupDuplicateDetector) PreProcess() {
	securityGroupsResp, err := d.AwsClient.DescribeSecurityGroups()
	if err != nil {
		d.Logger.Error(err)
		d.Error = true
		return
	}
	attrsResp, err := d.AwsClient.DescribeAccountAttributes()
	if err != nil {
		d.Logger.Error(err)
		d.Error = true
		return
	}

	for _, securityGroup := range securityGroupsResp.SecurityGroups {
		var vpcId string
		// If vpcId is nil, it is on EC2-Classic.
		if securityGroup.VpcId == nil {
			vpcId = "none"
		} else {
			vpcId = *securityGroup.VpcId
		}
		d.securiyGroups[vpcId+"."+*securityGroup.GroupName] = true
	}
	for _, attr := range attrsResp.AccountAttributes {
		if *attr.AttributeName == "default-vpc" {
			d.defaultVpc = *attr.AttributeValues[0].AttributeValue
			break
		}
	}
}

func (d *AwsSecurityGroupDuplicateDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	nameToken, ok := resource.GetToken("name")
	if !ok {
		return
	}
	name, err := d.evalToString(nameToken.Text)
	if err != nil {
		d.Logger.Error(err)
		return
	}
	var vpc string
	vpc, err = d.fetchVpcId(resource)
	if err != nil {
		d.Logger.Error(err)
		return
	}

	if d.securiyGroups[vpc+"."+name] && !d.State.Exists(d.Target, resource.Id) {
		issue := &issue.Issue{
			Type:    d.IssueType,
			Message: fmt.Sprintf("\"%s\" is duplicate name. It must be unique.", name),
			Line:    nameToken.Pos.Line,
			File:    nameToken.Pos.Filename,
		}
		*issues = append(*issues, issue)
	}
}

func (d *AwsSecurityGroupDuplicateDetector) fetchVpcId(resource *schema.Resource) (string, error) {
	var vpc string
	vpcToken, ok := resource.GetToken("vpc_id")
	if !ok {
		// "vpc_id" is optional. If omitted, use default vpc_id.
		vpc = d.defaultVpc
	} else {
		var err error
		vpc, err = d.evalToString(vpcToken.Text)
		if err != nil {
			return "", err
		}
	}

	return vpc, nil
}
