package detector

import (
	"fmt"

	"github.com/hashicorp/hcl/hcl/token"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsELBInvalidSecurityGroupDetector struct {
	*Detector
	securityGroups map[string]bool
}

func (d *Detector) CreateAwsELBInvalidSecurityGroupDetector() *AwsELBInvalidSecurityGroupDetector {
	nd := &AwsELBInvalidSecurityGroupDetector{
		Detector:       d,
		securityGroups: map[string]bool{},
	}
	nd.Name = "aws_elb_invalid_security_group"
	nd.IssueType = issue.ERROR
	nd.TargetType = "resource"
	nd.Target = "aws_elb"
	nd.DeepCheck = true
	return nd
}

func (d *AwsELBInvalidSecurityGroupDetector) PreProcess() {
	resp, err := d.AwsClient.DescribeSecurityGroups()
	if err != nil {
		d.Logger.Error(err)
		d.Error = true
		return
	}

	for _, securityGroup := range resp.SecurityGroups {
		d.securityGroups[*securityGroup.GroupId] = true
	}
}

func (d *AwsELBInvalidSecurityGroupDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	var varToken token.Token
	var securityGroupTokens []token.Token
	var ok bool
	if varToken, ok = resource.GetToken("security_groups"); ok {
		var err error
		securityGroupTokens, err = d.evalToStringTokens(varToken)
		if err != nil {
			d.Logger.Error(err)
			return
		}
	} else {
		securityGroupTokens, ok = resource.GetListToken("security_groups")
		if !ok {
			return
		}
	}

	for _, securityGroupToken := range securityGroupTokens {
		securityGroup, err := d.evalToString(securityGroupToken.Text)
		if err != nil {
			d.Logger.Error(err)
			continue
		}

		// If `security_groups` is interpolated by list variable, Filename is empty
		if securityGroupToken.Pos.Filename == "" {
			securityGroupToken.Pos.Filename = varToken.Pos.Filename
		}
		if !d.securityGroups[securityGroup] {
			issue := &issue.Issue{
				Detector: d.Name,
				Type:     d.IssueType,
				Message:  fmt.Sprintf("\"%s\" is invalid security group.", securityGroup),
				Line:     securityGroupToken.Pos.Line,
				File:     securityGroupToken.Pos.Filename,
			}
			*issues = append(*issues, issue)
		}
	}
}
