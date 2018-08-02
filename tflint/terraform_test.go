package tflint

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform/terraform"
	"github.com/zclconf/go-cty/cty"
)

func Test_getTFDataDir(t *testing.T) {
	cases := []struct {
		Name     string
		EnvVar   map[string]string
		Expected string
	}{
		{
			Name:     "default",
			Expected: ".terraform",
		},
		{
			Name:     "environment variable",
			EnvVar:   map[string]string{"TF_DATA_DIR": ".tfdata"},
			Expected: ".tfdata",
		},
	}

	for _, tc := range cases {
		for key, value := range tc.EnvVar {
			err := os.Setenv(key, value)
			if err != nil {
				t.Fatal(err)
			}
		}

		ret := getTFDataDir()
		if ret != tc.Expected {
			t.Fatalf("Failed `%s` test: expected value is %s, but get %s", tc.Name, tc.Expected, ret)
		}

		for key := range tc.EnvVar {
			err := os.Unsetenv(key)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}

func Test_getTFModuleDir(t *testing.T) {
	cases := []struct {
		Name     string
		EnvVar   map[string]string
		Expected string
	}{
		{
			Name:     "default",
			Expected: filepath.Join(".terraform", "modules"),
		},
		{
			Name:     "environment variable",
			EnvVar:   map[string]string{"TF_DATA_DIR": ".tfdata"},
			Expected: filepath.Join(".tfdata", "modules"),
		},
	}

	for _, tc := range cases {
		for key, value := range tc.EnvVar {
			err := os.Setenv(key, value)
			if err != nil {
				t.Fatal(err)
			}
		}

		ret := getTFModuleDir()
		if ret != tc.Expected {
			t.Fatalf("Failed `%s` test: expected value is %s, but get %s", tc.Name, tc.Expected, ret)
		}

		for key := range tc.EnvVar {
			err := os.Unsetenv(key)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}

func Test_getTFModuleManifestPath(t *testing.T) {
	cases := []struct {
		Name     string
		EnvVar   map[string]string
		Expected string
	}{
		{
			Name:     "default",
			Expected: filepath.Join(".terraform", "modules", "modules.json"),
		},
		{
			Name:     "environment variable",
			EnvVar:   map[string]string{"TF_DATA_DIR": ".tfdata"},
			Expected: filepath.Join(".tfdata", "modules", "modules.json"),
		},
	}

	for _, tc := range cases {
		for key, value := range tc.EnvVar {
			err := os.Setenv(key, value)
			if err != nil {
				t.Fatal(err)
			}
		}

		ret := getTFModuleManifestPath()
		if ret != tc.Expected {
			t.Fatalf("Failed `%s` test: expected value is %s, but get %s", tc.Name, tc.Expected, ret)
		}

		for key := range tc.EnvVar {
			err := os.Unsetenv(key)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}

func Test_getTFWorkspace(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		Name     string
		Dir      string
		EnvVar   map[string]string
		Expected string
	}{
		{
			Name:     "default",
			Expected: "default",
		},
		{
			Name:     "TF_WORKSPACE",
			EnvVar:   map[string]string{"TF_WORKSPACE": "dev"},
			Expected: "dev",
		},
		{
			Name:     "environment file",
			Dir:      filepath.Join(currentDir, "test-fixtures", "with_environment_file"),
			Expected: "staging",
		},
		{
			Name:     "TF_DATA_DIR",
			Dir:      filepath.Join(currentDir, "test-fixtures", "with_environment_file"),
			EnvVar:   map[string]string{"TF_DATA_DIR": ".terraform_production"},
			Expected: "production",
		},
	}

	for _, tc := range cases {
		if tc.Dir != "" {
			err := os.Chdir(tc.Dir)
			if err != nil {
				t.Fatal(err)
			}
		}

		for key, value := range tc.EnvVar {
			err := os.Setenv(key, value)
			if err != nil {
				t.Fatal(err)
			}
		}

		ret := getTFWorkspace()
		if ret != tc.Expected {
			t.Fatalf("Failed `%s` test: expected value is %s, but get %s", tc.Name, tc.Expected, ret)
		}

		for key := range tc.EnvVar {
			err := os.Unsetenv(key)
			if err != nil {
				t.Fatal(err)
			}
		}

		if tc.Dir != "" {
			err := os.Chdir(currentDir)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}

func Test_getTFEnvVariables(t *testing.T) {
	cases := []struct {
		Name     string
		EnvVar   map[string]string
		Expected terraform.InputValues
	}{
		{
			Name: "environment variable",
			EnvVar: map[string]string{
				"TF_VAR_instance_type": "t2.micro",
				"TF_VAR_count":         "5",
			},
			Expected: terraform.InputValues{
				"instance_type": &terraform.InputValue{
					Value:      cty.StringVal("t2.micro"),
					SourceType: terraform.ValueFromEnvVar,
				},
				"count": &terraform.InputValue{
					Value:      cty.StringVal("5"),
					SourceType: terraform.ValueFromEnvVar,
				},
			},
		},
	}

	for _, tc := range cases {
		for key, value := range tc.EnvVar {
			err := os.Setenv(key, value)
			if err != nil {
				t.Fatal(err)
			}
		}

		ret := getTFEnvVariables()
		if !reflect.DeepEqual(tc.Expected, ret) {
			t.Fatalf("Failed `%s` test:\n Expected: %#v\n Actual: %#v", tc.Name, tc.Expected, ret)
		}

		for key := range tc.EnvVar {
			err := os.Unsetenv(key)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}
