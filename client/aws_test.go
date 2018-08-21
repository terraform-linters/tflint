package client

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	homedir "github.com/mitchellh/go-homedir"
)

func Test_newAwsSession(t *testing.T) {
	type Result struct {
		Credentials *credentials.Credentials
		Region      *string
	}
	path, _ := homedir.Expand("~/.aws/credentials")

	cases := []struct {
		Name     string
		Creds    AwsCredentials
		Expected Result
	}{
		{
			Name: "static credentials",
			Creds: AwsCredentials{
				AccessKey: "AWS_ACCESS_KEY",
				SecretKey: "AWS_SECRET_KEY",
				Region:    "us-east-1",
			},
			Expected: Result{
				Credentials: credentials.NewStaticCredentials("AWS_ACCESS_KEY", "AWS_SECRET_KEY", ""),
				Region:      aws.String("us-east-1"),
			},
		},
		{
			Name: "shared credentials",
			Creds: AwsCredentials{
				Profile: "production",
				Region:  "us-east-1",
			},
			Expected: Result{
				Credentials: credentials.NewSharedCredentials(path, "production"),
				Region:      aws.String("us-east-1"),
			},
		},
	}

	for _, tc := range cases {
		s := newAwsSession(tc.Creds)
		if !reflect.DeepEqual(tc.Expected.Credentials, s.Config.Credentials) {
			t.Fatalf("Failed `%s` test: expected credentials are `%#v`, but get `%#v`", tc.Name, tc.Expected.Credentials, s.Config.Credentials)
		}
		if !reflect.DeepEqual(tc.Expected.Region, s.Config.Region) {
			t.Fatalf("Failed `%s` test: expected region are `%#v`, but get `%#v`", tc.Name, tc.Expected.Region, s.Config.Region)
		}
	}
}
