package detector

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/parser"
	"github.com/wata727/tflint/awsmock"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/evaluator"
	"github.com/wata727/tflint/issue"
)

func TestDetectAwsInstanceInvalidAMI(t *testing.T) {
	cases := []struct {
		Name     string
		Src      string
		Response []*ec2.Image
		Issues   []*issue.Issue
	}{
		{
			Name: "AMI is invalid",
			Src: `
resource "aws_instance" "web" {
    ami = "ami-1234abcd"
}`,
			Response: []*ec2.Image{
				&ec2.Image{
					ImageId: aws.String("ami-0c11b26d"),
				},
				&ec2.Image{
					ImageId: aws.String("ami-9ad76sd1"),
				},
			},
			Issues: []*issue.Issue{
				&issue.Issue{
					Type:    "ERROR",
					Message: "\"ami-1234abcd\" is invalid AMI.",
					Line:    3,
					File:    "test.tf",
				},
			},
		},
		{
			Name: "AMI is valid",
			Src: `
resource "aws_instance" "web" {
    ami = "ami-0c11b26d"
}`,
			Response: []*ec2.Image{
				&ec2.Image{
					ImageId: aws.String("ami-0c11b26d"),
				},
				&ec2.Image{
					ImageId: aws.String("ami-9ad76sd1"),
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
		d := &AwsInstanceInvalidAMIDetector{
			&Detector{
				ListMap:    listMap,
				EvalConfig: evalConfig,
				Config:     c,
				AwsClient:  c.NewAwsClient(),
			},
		}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		iammock := awsmock.NewMockEC2API(ctrl)
		iammock.EXPECT().DescribeImages(&ec2.DescribeImagesInput{}).Return(&ec2.DescribeImagesOutput{
			Images: tc.Response,
		}, nil)
		d.AwsClient.Ec2 = iammock

		var issues = []*issue.Issue{}
		d.Detect(&issues)

		if !reflect.DeepEqual(issues, tc.Issues) {
			t.Fatalf("Bad: %s\nExpected: %s\n\ntestcase: %s", issues, tc.Issues, tc.Name)
		}
	}
}
