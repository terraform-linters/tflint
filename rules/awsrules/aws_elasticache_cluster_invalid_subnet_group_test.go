package awsrules

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/configs/configload"
	"github.com/hashicorp/terraform/terraform"
	"github.com/wata727/tflint/client"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

func Test_AwsElastiCacheClusterInvalidSubnetGroup(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Response []*elasticache.CacheSubnetGroup
		Expected issue.Issues
	}{
		{
			Name: "parameter_group_name is invalid",
			Content: `
resource "aws_elasticache_cluster" "redis" {
    subnet_group_name = "app-server"
}`,
			Response: []*elasticache.CacheSubnetGroup{
				{
					CacheSubnetGroupName: aws.String("app-server1"),
				},
				{
					CacheSubnetGroupName: aws.String("app-server2"),
				},
			},
			Expected: []*issue.Issue{
				{
					Detector: "aws_elasticache_cluster_invalid_subnet_group",
					Type:     "ERROR",
					Message:  "\"app-server\" is invalid subnet group name.",
					Line:     3,
					File:     "resource.tf",
				},
			},
		},
		{
			Name: "parameter_group_name is valid",
			Content: `
resource "aws_elasticache_cluster" "redis" {
    subnet_group_name = "app-server"
}`,
			Response: []*elasticache.CacheSubnetGroup{
				{
					CacheSubnetGroupName: aws.String("app-server1"),
				},
				{
					CacheSubnetGroupName: aws.String("app-server2"),
				},
				{
					CacheSubnetGroupName: aws.String("app-server"),
				},
			},
			Expected: []*issue.Issue{},
		},
	}

	dir, err := ioutil.TempDir("", "AwsElastiCacheClusterInvalidSubnetGroup")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(currentDir)

	err = os.Chdir(dir)
	if err != nil {
		t.Fatal(err)
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for _, tc := range cases {
		loader, err := configload.NewLoader(&configload.Config{})
		if err != nil {
			t.Fatal(err)
		}

		err = ioutil.WriteFile(dir+"/resource.tf", []byte(tc.Content), os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}

		mod, diags := loader.Parser().LoadConfigDir(".")
		if diags.HasErrors() {
			t.Fatal(diags)
		}
		cfg, tfdiags := configs.BuildConfig(mod, configs.DisabledModuleWalker)
		if tfdiags.HasErrors() {
			t.Fatal(tfdiags)
		}

		runner, err := tflint.NewRunner(tflint.EmptyConfig(), map[string]tflint.Annotations{}, cfg, map[string]*terraform.InputValue{})
		if err != nil {
			t.Fatal(err)
		}
		rule := NewAwsElastiCacheClusterInvalidSubnetGroupRule()

		mock := client.NewMockElastiCacheAPI(ctrl)
		mock.EXPECT().DescribeCacheSubnetGroups(&elasticache.DescribeCacheSubnetGroupsInput{}).Return(&elasticache.DescribeCacheSubnetGroupsOutput{
			CacheSubnetGroups: tc.Response,
		}, nil)
		runner.AwsClient.ElastiCache = mock

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		if !cmp.Equal(tc.Expected, runner.Issues) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues))
		}
	}
}
