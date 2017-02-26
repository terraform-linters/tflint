package detector

import (
	"fmt"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/token"
	"github.com/wata727/tflint/issue"
)

type AwsInstanceInvalidVPCSecurityGroupDetector struct {
	*Detector
	IssueType      string
	Target         string
	DeepCheck      bool
	securityGroups map[string]bool
}

func (d *Detector) CreateAwsInstanceInvalidVPCSecurityGroupDetector() *AwsInstanceInvalidVPCSecurityGroupDetector {
	return &AwsInstanceInvalidVPCSecurityGroupDetector{
		Detector:       d,
		IssueType:      issue.ERROR,
		Target:         "aws_instance",
		DeepCheck:      true,
		securityGroups: map[string]bool{},
	}
}

func (d *AwsInstanceInvalidVPCSecurityGroupDetector) PreProcess() {
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

func (d *AwsInstanceInvalidVPCSecurityGroupDetector) Detect(file string, item *ast.ObjectItem, issues *[]*issue.Issue) {
	var varToken token.Token
	var securityGroupTokens []token.Token
	var err error
	if varToken, err = hclLiteralToken(item, "vpc_security_group_ids"); err == nil {
		securityGroupTokens, err = d.evalToStringTokens(varToken)
		if err != nil {
			d.Logger.Error(err)
			return
		}
	} else {
		d.Logger.Error(err)
		securityGroupTokens, err = hclLiteralListToken(item, "vpc_security_group_ids")
		if err != nil {
			d.Logger.Error(err)
			return
		}
	}

	for _, securityGroupToken := range securityGroupTokens {
		securityGroup, err := d.evalToString(securityGroupToken.Text)
		if err != nil {
			d.Logger.Error(err)
			continue
		}

		if !d.securityGroups[securityGroup] {
			issue := &issue.Issue{
				Type:    d.IssueType,
				Message: fmt.Sprintf("\"%s\" is invalid security group.", securityGroup),
				Line:    securityGroupToken.Pos.Line,
				File:    file,
			}
			*issues = append(*issues, issue)
		}
	}
}
