package client

import (
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	awsbase "github.com/hashicorp/aws-sdk-go-base"
	homedir "github.com/mitchellh/go-homedir"
)

func Test_Merge(t *testing.T) {
	cases := []struct {
		Name     string
		Self     AwsCredentials
		Other    AwsCredentials
		Expected AwsCredentials
	}{
		{
			Name: "self is empty",
			Self: AwsCredentials{},
			Other: AwsCredentials{
				AccessKey: "AWS_ACCESS_KEY",
				SecretKey: "AWS_SECRET_KEY",
				Profile:   "default",
				CredsFile: "~/.aws/creds",
				Region:    "us-east-1",
			},
			Expected: AwsCredentials{
				AccessKey: "AWS_ACCESS_KEY",
				SecretKey: "AWS_SECRET_KEY",
				Profile:   "default",
				CredsFile: "~/.aws/creds",
				Region:    "us-east-1",
			},
		},
		{
			Name: "other is empty",
			Self: AwsCredentials{
				AccessKey: "AWS_ACCESS_KEY_2",
				SecretKey: "AWS_SECRET_KEY_2",
				Profile:   "staging",
				CredsFile: "~/.aws/creds_stg",
				Region:    "ap-northeast-1",
			},
			Other: AwsCredentials{},
			Expected: AwsCredentials{
				AccessKey: "AWS_ACCESS_KEY_2",
				SecretKey: "AWS_SECRET_KEY_2",
				Profile:   "staging",
				CredsFile: "~/.aws/creds_stg",
				Region:    "ap-northeast-1",
			},
		},
		{
			Name: "merged",
			Self: AwsCredentials{
				AccessKey: "AWS_ACCESS_KEY_2",
				SecretKey: "AWS_SECRET_KEY_2",
				Profile:   "staging",
				CredsFile: "~/.aws/creds_stg",
				Region:    "ap-northeast-1",
			},
			Other: AwsCredentials{
				AccessKey: "AWS_ACCESS_KEY",
				SecretKey: "AWS_SECRET_KEY",
				Profile:   "default",
				CredsFile: "~/.aws/creds",
				Region:    "us-east-1",
			},
			Expected: AwsCredentials{
				AccessKey: "AWS_ACCESS_KEY",
				SecretKey: "AWS_SECRET_KEY",
				Profile:   "default",
				CredsFile: "~/.aws/creds",
				Region:    "us-east-1",
			},
		},
	}

	for _, tc := range cases {
		ret := tc.Self.Merge(tc.Other)
		if !cmp.Equal(tc.Expected, ret) {
			t.Fatalf("Failed `%s` test: Diff=%s", tc.Name, cmp.Diff(tc.Expected, ret))
		}
	}
}

func Test_getBaseConfig(t *testing.T) {
	home, err := homedir.Expand("~/")
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		Name     string
		Creds    AwsCredentials
		Expected *awsbase.Config
	}{
		{
			Name: "static credentials",
			Creds: AwsCredentials{
				AccessKey: "AWS_ACCESS_KEY",
				SecretKey: "AWS_SECRET_KEY",
				Region:    "us-east-1",
			},
			Expected: &awsbase.Config{
				AccessKey: "AWS_ACCESS_KEY",
				SecretKey: "AWS_SECRET_KEY",
				Region:    "us-east-1",
			},
		},
		{
			Name: "shared credentials",
			Creds: AwsCredentials{
				Profile:   "default",
				CredsFile: "~/.aws/creds",
				Region:    "us-east-1",
			},
			Expected: &awsbase.Config{
				Profile:       "default",
				CredsFilename: filepath.Join(home, ".aws", "creds"),
				Region:        "us-east-1",
			},
		},
	}

	for _, tc := range cases {
		base, err := getBaseConfig(tc.Creds)
		if err != nil {
			t.Fatalf("Failed `%s` test: Unexpected error occurred: %s", tc.Name, err)
		}
		if !cmp.Equal(tc.Expected, base) {
			t.Fatalf("Failed `%s` test: Diff=%s", tc.Name, cmp.Diff(tc.Expected, base))
		}
	}
}
