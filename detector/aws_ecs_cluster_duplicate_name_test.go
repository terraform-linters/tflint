package detector

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/golang/mock/gomock"
	"github.com/k0kubun/pp"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/mock"
)

func TestDetectAwsECSDuplicateName(t *testing.T) {
	cases := []struct {
		Name     string
		Src      string
		State    string
		Response []*ecs.Cluster
		Issues   []*issue.Issue
	}{
		{
			Name: "Cluster name is duplicate",
			Src: `
resource "aws_ecs_cluster" "foo" {
  name = "white-hart"
}`,
			Response: []*ecs.Cluster{
				{
					ClusterName: aws.String("white-hart"),
				},
				{
					ClusterName: aws.String("black-hart"),
				},
			},
			Issues: []*issue.Issue{
				{
					Detector: "aws_ecs_cluster_duplicate_name",
					Type:     "ERROR",
					Message:  "\"white-hart\" is duplicate name. It must be unique.",
					Line:     3,
					File:     "test.tf",
				},
			},
		},
		{
			Name: "Cluster name is unique",
			Src: `
resource "aws_ecs_cluster" "foo" {
  name = "white-hart"
}`,
			Response: []*ecs.Cluster{
				{
					ClusterName: aws.String("brown-hart"),
				},
				{
					ClusterName: aws.String("black-hart"),
				},
			},
			Issues: []*issue.Issue{},
		},
		{
			Name: "Cluster is duplicate, but exists in state",
			Src: `
resource "aws_ecs_cluster" "foo" {
  name = "white-hart"
}`,
			State: `
{
    "modules": [
        {
            "resources": {
                "aws_ecs_cluster.foo": {
                    "type": "aws_ecs_cluster",
                    "depends_on": [],
                    "primary": {
                        "id": "arn:aws:ecs:us-east-1:hogehoge:cluster/white-hart",
                        "attributes": {
                            "id": "arn:aws:ecs:us-east-1:hogehoge:cluster/white-hart",
                            "name": "white-hart"
                        },
                        "meta": {},
                        "tainted": false
                    },
                    "deposed": [],
                    "provider": ""
                }
            }
        }
    ]
}
`,
			Response: []*ecs.Cluster{
				{
					ClusterName: aws.String("white-hart"),
				},
				{
					ClusterName: aws.String("black-hart"),
				},
			},
			Issues: []*issue.Issue{},
		},
	}

	for _, tc := range cases {
		c := config.Init()
		c.DeepCheck = true

		awsClient := c.NewAwsClient()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ecsmock := mock.NewMockECSAPI(ctrl)
		ecsmock.EXPECT().ListClusters(&ecs.ListClustersInput{}).Return(&ecs.ListClustersOutput{}, nil)
		ecsmock.EXPECT().DescribeClusters(&ecs.DescribeClustersInput{}).Return(&ecs.DescribeClustersOutput{
			Clusters: tc.Response,
		}, nil)
		awsClient.Ecs = ecsmock

		var issues = []*issue.Issue{}
		err := TestDetectByCreatorName(
			"CreateAwsECSClusterDuplicateNameDetector",
			tc.Src,
			tc.State,
			c,
			awsClient,
			&issues,
		)
		if err != nil {
			t.Fatalf("\nERROR: %s", err)
		}

		if !reflect.DeepEqual(issues, tc.Issues) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(issues), pp.Sprint(tc.Issues), tc.Name)
		}
	}
}
