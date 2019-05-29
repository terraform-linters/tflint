package awsrules

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/configs/configload"
	"github.com/hashicorp/terraform/terraform"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

func Test_AwsCloudwatchMetricAlarmInvalidUnit(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected issue.Issues
	}{
		{
			Name: "GB is invalid",
			Content: `
resource "aws_cloudwatch_metric_alarm" "test" {
    metric_name         = "FreeableMemory"
    namespace           = "AWS/RDS"

    period    = "300"
    statistic = "Average"
    threshold = "1"
    unit      = "GB"
}`,
			Expected: []*issue.Issue{
				{
					Detector: "aws_cloudwatch_metric_alarm_invalid_unit",
					Type:     "ERROR",
					Message:  "\"GB\" is invalid unit.",
					Line:     9,
					File:     "resource.tf",
				},
			},
		},
		{
			Name: "Lowercase is invalid",
			Content: `
resource "aws_cloudwatch_metric_alarm" "test" {
    metric_name         = "FreeableMemory"
    namespace           = "AWS/RDS"

    period    = "300"
    statistic = "Average"
    threshold = "1"
    unit      = "gigabytes"
}`,
			Expected: []*issue.Issue{
				{
					Detector: "aws_cloudwatch_metric_alarm_invalid_unit",
					Type:     "ERROR",
					Message:  "\"gigabytes\" is invalid unit.",
					Line:     9,
					File:     "resource.tf",
				},
			},
		},
		{
			Name: "Gigabytes is valid",
			Content: `
resource "aws_cloudwatch_metric_alarm" "test" {
    metric_name         = "FreeableMemory"
    namespace           = "AWS/RDS"

    period    = "300"
    statistic = "Average"
    threshold = "1"
    unit      = "Gigabytes"
}`,
			Expected: []*issue.Issue{},
		},
	}

	dir, err := ioutil.TempDir("", "AwsCloudwatchMetricAlarmInvalidUnit")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	for _, tc := range cases {
		loader, err := configload.NewLoader(&configload.Config{})
		if err != nil {
			t.Fatal(err)
		}

		err = ioutil.WriteFile(dir+"/resource.tf", []byte(tc.Content), os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}

		mod, diags := loader.Parser().LoadConfigDir(dir)
		if diags.HasErrors() {
			t.Fatal(diags)
		}
		cfg, tfdiags := configs.BuildConfig(mod, configs.DisabledModuleWalker)
		if tfdiags.HasErrors() {
			t.Fatal(tfdiags)
		}

		runner := tflint.NewRunner(tflint.EmptyConfig(), map[string]tflint.Annotations{}, cfg, map[string]*terraform.InputValue{})
		rule := NewAwsCloudwatchMetricAlarmInvalidUnitRule()

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		if !cmp.Equal(tc.Expected, runner.Issues) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues))
		}
	}
}
