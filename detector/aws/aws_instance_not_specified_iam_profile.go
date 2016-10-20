package aws

import (
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/wata727/tflint/issue"
)

func DetectAwsInstanceNotSpecifiedIamProfile(list *ast.ObjectList, file string) []*issue.Issue {
	var issues = []*issue.Issue{}

	for _, item := range list.Filter("resource", "aws_instance").Items {
		instanceIAMProfile := item.Val.(*ast.ObjectType).List.Filter("iam_instance_profile")

		if len(instanceIAMProfile.Items) == 0 {
			issue := &issue.Issue{
				Type:    "NOTICE",
				Message: "\"iam_instance_profile\" is not specified. You cannot edit this value later.",
				Line:    item.Pos().Line,
				File:    file,
			}
			issues = append(issues, issue)
		}
	}

	return issues
}
