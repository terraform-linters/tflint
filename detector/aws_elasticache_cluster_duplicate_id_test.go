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

func TestDetectAwsElastiCacheClusterDuplicateID(t *testing.T) {
	cases := []struct {
		Name     string
		Src      string
		State    string
		Response []*elasticache.CacheCluster
		Issues   []*issue.Issue
	}{
		{
			Name: "cluster_id is duplicate",
			Src: `
resource "aws_elasticache_cluster" "test" {
    cluster_id = "cluster-example"
}`,
			Response: []*elasticache.CacheCluster{
				&elasticache.CacheCluster{
					CacheClusterId: aws.String("cluster-example"),
				},
				&elasticache.CacheCluster{
					CacheClusterId: aws.String("test-cluster"),
				},
			},
			Issues: []*issue.Issue{
				&issue.Issue{
					Type:    "ERROR",
					Message: "\"cluster-example\" is duplicate Cluster ID. It must be unique.",
					Line:    3,
					File:    "test.tf",
				},
			},
		},
		{
			Name: "cluster_id is unique",
			Src: `
resource "aws_elasticache_cluster" "test" {
    cluster_id = "cluster-example"
}`,
			Response: []*elasticache.CacheCluster{
				&elasticache.CacheCluster{
					CacheClusterId: aws.String("example-cluster"),
				},
				&elasticache.CacheCluster{
					CacheClusterId: aws.String("test-cluster"),
				},
			},
			Issues: []*issue.Issue{},
		},
		{
			Name: "cluster_id is duplicate, but exists in state",
			Src: `
resource "aws_elasticache_cluster" "test" {
    cluster_id = "cluster-example"
}`,
			State: `
{
    "modules": [
        {
            "resources": {
                "aws_elasticache_cluster.test": {
                    "type": "aws_elasticache_cluster",
                    "depends_on": [],
                    "primary": {
                        "id": "cluster-example",
                        "attributes": {
                            "availability_zone": "us-east-1a",
                            "cache_nodes.#": "1",
                            "cache_nodes.0.address": "cluster-example.hogehoge.0001.useast.cache.amazonaws.com",
                            "cache_nodes.0.availability_zone": "us-east-1a",
                            "cache_nodes.0.id": "0001",
                            "cache_nodes.0.port": "11211",
                            "cluster_address": "cluster-example.hogehoge.cfg.useast.cache.amazonaws.com",
                            "cluster_id": "cluster-example",
                            "configuration_endpoint": "cluster-example.hogehoge.cfg.useast.cache.amazonaws.com:11211",
                            "engine": "memcached",
                            "engine_version": "1.4.33",
                            "id": "cluster-example",
                            "maintenance_window": "sun:15:00-sun:16:00",
                            "node_type": "cache.t2.micro",
                            "num_cache_nodes": "1",
                            "parameter_group_name": "default.memcached1.4",
                            "port": "11211",
                            "security_group_ids.#": "0",
                            "security_group_names.#": "0",
                            "snapshot_retention_limit": "0",
                            "snapshot_window": "",
                            "subnet_group_name": "default",
                            "tags.%": "0"
                        }
                    },
                    "provider": ""
                }
            }
        }
    ]
}
`,
			Response: []*elasticache.CacheCluster{
				&elasticache.CacheCluster{
					CacheClusterId: aws.String("cluster-example"),
				},
				&elasticache.CacheCluster{
					CacheClusterId: aws.String("test-cluster"),
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
		elasticachemock.EXPECT().DescribeCacheClusters(&elasticache.DescribeCacheClustersInput{}).Return(&elasticache.DescribeCacheClustersOutput{
			CacheClusters: tc.Response,
		}, nil)
		awsClient.Elasticache = elasticachemock

		var issues = []*issue.Issue{}
		TestDetectByCreatorName(
			"CreateAwsElastiCacheClusterDuplicateIDDetector",
			tc.Src,
			tc.State,
			c,
			awsClient,
			&issues,
		)

		if !reflect.DeepEqual(issues, tc.Issues) {
			t.Fatalf("Bad: %s\nExpected: %s\n\ntestcase: %s", issues, tc.Issues, tc.Name)
		}
	}
}
