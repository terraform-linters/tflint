package client

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/wata727/tflint/config"
)

func Test_newAwsSession(t *testing.T) {
	type Result struct {
		Credentials *credentials.Credentials
		Region      *string
	}
	path, _ := homedir.Expand("~/.aws/credentials")

	cases := []struct {
		Name     string
		Config   *config.Config
		Expected Result
	}{
		{
			Name: "static credentials",
			Config: &config.Config{
				AwsCredentials: map[string]string{
					"access_key": "AWS_ACCESS_KEY",
					"secret_key": "AWS_SECRET_KEY",
					"region":     "us-east-1",
				},
			},
			Expected: Result{
				Credentials: credentials.NewStaticCredentials("AWS_ACCESS_KEY", "AWS_SECRET_KEY", ""),
				Region:      aws.String("us-east-1"),
			},
		},
		{
			Name: "shared credentials",
			Config: &config.Config{
				AwsCredentials: map[string]string{
					"profile": "account1",
					"region":  "us-east-1",
				},
			},
			Expected: Result{
				Credentials: credentials.NewSharedCredentials(path, "account1"),
				Region:      aws.String("us-east-1"),
			},
		},
	}

	for _, tc := range cases {
		s := newAwsSession(tc.Config)
		if !reflect.DeepEqual(tc.Expected.Credentials, s.Config.Credentials) {
			t.Fatalf("Failed `%s` test: expected credentials are `%#v`, but get `%#v`", tc.Name, tc.Expected.Credentials, s.Config.Credentials)
		}
		if !reflect.DeepEqual(tc.Expected.Region, s.Config.Region) {
			t.Fatalf("Failed `%s` test: expected region are `%#v`, but get `%#v`", tc.Name, tc.Expected.Region, s.Config.Region)
		}
	}
}
