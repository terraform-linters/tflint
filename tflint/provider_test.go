package tflint

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform/configs/configschema"
	"github.com/terraform-linters/tflint/client"
	"github.com/zclconf/go-cty/cty"
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

func Test_GetBlock(t *testing.T) {
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
			Schema *configschema.Block
			Value  map[string]string
			Exists bool
			Err    error
		}{
			{
				Key: "assume_role",
				Schema: &configschema.Block{
					Attributes: map[string]*configschema.Attribute{
						"role_arn":     {Type: cty.String},
						"session_name": {Type: cty.String},
						"external_id":  {Type: cty.String},
						"policy":       {Type: cty.String},
					},
				},
				Value: map[string]string{
					"role_arn":     "arn:aws:iam::ACCOUNT_ID:role/ROLE_NAME",
					"session_name": "SESSION_NAME",
					"external_id":  "EXTERNAL_ID",
					"policy":       "POLICY_NAME",
				},
				Exists: true,
				Err:    nil,
			},
			{
				Key:    "undefined",
				Value:  map[string]string{},
				Exists: false,
				Err:    nil,
			},
		}

		for _, tc := range cases {
			val, exists, err := providerConfig.GetBlock(tc.Key, tc.Schema)
			if err != tc.Err {
				t.Fatalf("Expected `%s` as the error, but got `%s`", tc.Err, err)
			}
			if exists != tc.Exists {
				t.Fatalf("Expected `%t` as the exists, but got `%t`", tc.Exists, exists)
			}
			if !cmp.Equal(tc.Value, val) {
				t.Fatalf("Expected value is not matched:\n %s\n", cmp.Diff(tc.Value, val))
			}
		}
	})
}
