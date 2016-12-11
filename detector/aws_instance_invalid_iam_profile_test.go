package detector

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/parser"
	"github.com/wata727/tflint/awsmock"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/evaluator"
	"github.com/wata727/tflint/issue"
)

func TestDetectAwsInstanceInvalidIAMProfile(t *testing.T) {
	cases := []struct {
		Name     string
		Src      string
		Response []*iam.InstanceProfile
		Issues   []*issue.Issue
	}{
		{
			Name: "iam_instance_profile is invalid",
			Src: `
resource "aws_instance" "web" {
    iam_instance_profile = "app-server"
}`,
			Response: []*iam.InstanceProfile{
				&iam.InstanceProfile{
					InstanceProfileName: aws.String("app-server1"),
				},
				&iam.InstanceProfile{
					InstanceProfileName: aws.String("app-server2"),
				},
			},
			Issues: []*issue.Issue{
				&issue.Issue{
					Type:    "ERROR",
					Message: "\"app-server\" is invalid IAM profile name.",
					Line:    3,
					File:    "test.tf",
				},
			},
		},
		{
			Name: "iam_instance_profile is valid",
			Src: `
resource "aws_instance" "web" {
    iam_instance_profile = "app-server"
}`,
			Response: []*iam.InstanceProfile{
				&iam.InstanceProfile{
					InstanceProfileName: aws.String("app-server1"),
				},
				&iam.InstanceProfile{
					InstanceProfileName: aws.String("app-server2"),
				},
				&iam.InstanceProfile{
					InstanceProfileName: aws.String("app-server"),
				},
			},
			Issues: []*issue.Issue{},
		},
	}

	for _, tc := range cases {
		listMap := make(map[string]*ast.ObjectList)
		root, _ := parser.Parse([]byte(tc.Src))
		list, _ := root.Node.(*ast.ObjectList)
		listMap["test.tf"] = list

		c := config.Init()
		c.DeepCheck = true
		evalConfig, _ := evaluator.NewEvaluator(listMap, config.Init())
		d := &AwsInstanceInvalidIAMProfileDetector{
			&Detector{
				ListMap:    listMap,
				EvalConfig: evalConfig,
				Config:     c,
				AwsClient:  c.NewAwsClient(),
			},
		}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		iammock := awsmock.NewMockIAMAPI(ctrl)
		iammock.EXPECT().ListInstanceProfiles(&iam.ListInstanceProfilesInput{}).Return(&iam.ListInstanceProfilesOutput{
			InstanceProfiles: tc.Response,
		}, nil)
		d.AwsClient.Iam = iammock

		var issues = []*issue.Issue{}
		d.Detect(&issues)

		if !reflect.DeepEqual(issues, tc.Issues) {
			t.Fatalf("Bad: %s\nExpected: %s\n\ntestcase: %s", issues, tc.Issues, tc.Name)
		}
	}
}
