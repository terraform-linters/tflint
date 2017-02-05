package detector

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/golang/mock/gomock"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/mock"
)

func TestDetectAwsElastiCacheInvalidParameterGroup(t *testing.T) {
	cases := []struct {
		Name     string
		Src      string
		Response []*elasticache.CacheParameterGroup
		Issues   []*issue.Issue
	}{
		{
			Name: "parameter_group_name is invalid",
			Src: `
resource "aws_elasticache_cluster" "redis" {
    parameter_group_name = "app-server"
}`,
			Response: []*elasticache.CacheParameterGroup{
				&elasticache.CacheParameterGroup{
					CacheParameterGroupName: aws.String("app-server1"),
				},
				&elasticache.CacheParameterGroup{
					CacheParameterGroupName: aws.String("app-server2"),
				},
			},
			Issues: []*issue.Issue{
				&issue.Issue{
					Type:    "ERROR",
					Message: "\"app-server\" is invalid parameter group name.",
					Line:    3,
					File:    "test.tf",
				},
			},
		},
		{
			Name: "parameter_group_name is valid",
			Src: `
resource "aws_elasticache_cluster" "redis" {
    parameter_group_name = "app-server"
}`,
			Response: []*elasticache.CacheParameterGroup{
				&elasticache.CacheParameterGroup{
					CacheParameterGroupName: aws.String("app-server1"),
				},
				&elasticache.CacheParameterGroup{
					CacheParameterGroupName: aws.String("app-server2"),
				},
				&elasticache.CacheParameterGroup{
					CacheParameterGroupName: aws.String("app-server"),
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
		elasticachemock := mock.NewMockElastiCacheAPI(ctrl)
		elasticachemock.EXPECT().DescribeCacheParameterGroups(&elasticache.DescribeCacheParameterGroupsInput{}).Return(&elasticache.DescribeCacheParameterGroupsOutput{
			CacheParameterGroups: tc.Response,
		}, nil)
		awsClient.Elasticache = elasticachemock

		var issues = []*issue.Issue{}
		TestDetectByCreatorName(
			"CreateAwsElastiCacheClusterInvalidParameterGroupDetector",
			tc.Src,
			"",
			c,
			awsClient,
			&issues,
		)

		if !reflect.DeepEqual(issues, tc.Issues) {
			t.Fatalf("Bad: %s\nExpected: %s\n\ntestcase: %s", issues, tc.Issues, tc.Name)
		}
	}
}
