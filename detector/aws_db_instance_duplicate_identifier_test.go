package detector

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/golang/mock/gomock"
	"github.com/k0kubun/pp"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/mock"
)

func TestDetectAwsDBInstanceDuplicateIdentifier(t *testing.T) {
	cases := []struct {
		Name     string
		Src      string
		State    string
		Response []*rds.DBInstance
		Issues   []*issue.Issue
	}{
		{
			Name: "identifier is duplicate",
			Src: `
resource "aws_db_instance" "test" {
    identifier = "my-db"
}`,
			Response: []*rds.DBInstance{
				&rds.DBInstance{
					DBInstanceIdentifier: aws.String("my-db"),
				},
				&rds.DBInstance{
					DBInstanceIdentifier: aws.String("your-db"),
				},
			},
			Issues: []*issue.Issue{
				&issue.Issue{
					Type:    "ERROR",
					Message: "\"my-db\" is duplicate identifier. It must be unique.",
					Line:    3,
					File:    "test.tf",
				},
			},
		},
		{
			Name: "identifier is unique",
			Src: `
resource "aws_db_instance" "test" {
    name = "my-db"
}`,
			Response: []*rds.DBInstance{
				&rds.DBInstance{
					DBInstanceIdentifier: aws.String("our-db"),
				},
				&rds.DBInstance{
					DBInstanceIdentifier: aws.String("your-db"),
				},
			},
			Issues: []*issue.Issue{},
		},
		{
			Name: "omitted identifier",
			Src: `
resource "aws_db_instance" "test" {
    instance_class = "db.t2.micro"
}`,
			Issues: []*issue.Issue{},
		},
		{
			Name: "identifier is duplicate, but exists in state",
			Src: `
resource "aws_db_instance" "test" {
    identifier = "my-db"
}`,
			State: `
{
    "modules": [
        {
            "resources": {
                "aws_db_instance.test": {
                    "type": "aws_db_instance",
                    "depends_on": [],
                    "primary": {
                        "id": "my-db",
                        "attributes": {
                            "address": "my-db.hogehoge.us-east-1.rds.amazonaws.com",
                            "allocated_storage": "10",
                            "arn": "arn:aws:rds:us-east-1:hogehoge:db:my-db",
                            "auto_minor_version_upgrade": "true",
                            "availability_zone": "us-east-1a",
                            "backup_retention_period": "0",
                            "backup_window": "18:36-19:06",
                            "copy_tags_to_snapshot": "false",
                            "db_subnet_group_name": "dbsubnet",
                            "endpoint": "my-db.hogehoge.us-east-1.rds.amazonaws.com:3306",
                            "engine": "mysql",
                            "engine_version": "5.6.27",
                            "hosted_zone_id": "HAND737292A",
                            "id": "my-db",
                            "identifier": "my-db",
                            "instance_class": "db.t1.micro",
                            "iops": "0",
                            "kms_key_id": "",
                            "license_model": "general-public-license",
                            "maintenance_window": "mon:17:06-mon:17:36",
                            "monitoring_interval": "0",
                            "multi_az": "false",
                            "name": "mydb",
                            "option_group_name": "default:mysql-5-6",
                            "parameter_group_name": "default.mysql5.6",
                            "password": "password",
                            "port": "3306",
                            "publicly_accessible": "false",
                            "replicas.#": "0",
                            "replicate_source_db": "",
                            "security_group_names.#": "0",
                            "skip_final_snapshot": "true",
                            "status": "available",
                            "storage_encrypted": "false",
                            "storage_type": "standard",
                            "tags.%": "0",
                            "timezone": "",
                            "username": "foo",
                            "vpc_security_group_ids.#": "1",
                            "vpc_security_group_ids.3963419045": "sg-1234abcd"
                        }
                    },
                    "provider": ""
                }
            }
        }
    ]
}
`,
			Response: []*rds.DBInstance{
				&rds.DBInstance{
					DBInstanceIdentifier: aws.String("my-db"),
				},
				&rds.DBInstance{
					DBInstanceIdentifier: aws.String("your-db"),
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
		rdsmock := mock.NewMockRDSAPI(ctrl)
		rdsmock.EXPECT().DescribeDBInstances(&rds.DescribeDBInstancesInput{}).Return(&rds.DescribeDBInstancesOutput{
			DBInstances: tc.Response,
		}, nil)
		awsClient.Rds = rdsmock

		var issues = []*issue.Issue{}
		TestDetectByCreatorName(
			"CreateAwsDBInstanceDuplicateIdentifierDetector",
			tc.Src,
			tc.State,
			c,
			awsClient,
			&issues,
		)

		if !reflect.DeepEqual(issues, tc.Issues) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(issues), pp.Sprint(tc.Issues), tc.Name)
		}
	}
}
