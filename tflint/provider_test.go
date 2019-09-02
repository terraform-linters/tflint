package tflint

import (
	"testing"

	"github.com/wata727/tflint/client"
)

func Test_Get(t *testing.T) {
	withinFixtureDir(t, "provider_config", func() {
		runner := testRunnerWithOsFs(t, EmptyConfig())
		providerConfig, err := NewProviderConfig(
			runner.TFConfig.Module.ProviderConfigs["aws"],
			runner,
			client.AwsProviderBlockSchema,
		)
		if err != nil {
			t.Fatalf("Unexpected error occrred: %s", err)
		}

		cases := []struct {
			Key    string
			Value  string
			Exists bool
			Err    error
		}{
			{
				Key:    "access_key",
				Value:  "AWS_ACCESS_KEY",
				Exists: true,
				Err:    nil,
			},
			{
				Key:    "secret_key",
				Value:  "",
				Exists: true,
				Err:    nil,
			},
			{
				Key:    "region",
				Value:  "us-east-1",
				Exists: true,
				Err:    nil,
			},
			{
				Key:    "profile",
				Value:  "",
				Exists: true,
				Err:    nil,
			},
			{
				Key:    "shared_credentials_file",
				Value:  "",
				Exists: true,
				Err:    nil,
			},
			{
				Key:    "undefined",
				Value:  "",
				Exists: false,
				Err:    nil,
			},
		}

		for _, tc := range cases {
			val, exists, err := providerConfig.Get(tc.Key)
			if val != tc.Value {
				t.Fatalf("Expected `%s` as the key value of `%s`, but got `%s`", tc.Value, tc.Key, val)
			}
			if exists != tc.Exists {
				t.Fatalf("Expected `%t` as the exists, but got `%t`", tc.Exists, exists)
			}
			if err != tc.Err {
				t.Fatalf("Expected `%s` as the error, but got `%s`", tc.Err, err)
			}
		}
	})
}

func Test_Get_withEmptyProvider(t *testing.T) {
	withinFixtureDir(t, "provider_config", func() {
		runner := testRunnerWithOsFs(t, EmptyConfig())
		providerConfig, err := NewProviderConfig(
			nil,
			runner,
			client.AwsProviderBlockSchema,
		)
		if err != nil {
			t.Fatalf("Unexpected error occrred: %s", err)
		}

		val, exists, err := providerConfig.Get("key")
		if val != "" {
			t.Fatalf("Expected empty string, but got `%s`", val)
		}
		if exists {
			t.Fatal("Expected not exists, but exists")
		}
		if err != nil {
			t.Fatalf("Expected to return nil, but got `%s`", err)
		}
	})
}
