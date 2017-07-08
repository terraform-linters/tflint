package detector

import (
	"reflect"
	"testing"

	"github.com/k0kubun/pp"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/issue"
)

func TestAwsCloudWatchMetricAlarmInvalidUnit(t *testing.T) {
	cases := []struct {
		Name   string
		Src    string
		Issues []*issue.Issue
	}{
		{
			Name: "GB is invalid",
			Src: `
resource "aws_cloudwatch_metric_alarm" "test" {
    metric_name         = "FreeableMemory"
    namespace           = "AWS/RDS"

    period    = "300"
    statistic = "Average"
    threshold = "1"
    unit      = "GB"
}`,
			Issues: []*issue.Issue{
				{
					Detector: "aws_cloudwatch_metric_alarm_invalid_unit",
					Type:     "ERROR",
					Message:  "\"GB\" is invalid unit.",
					Line:     9,
					File:     "test.tf",
					Link:     "https://github.com/wata727/tflint/blob/master/docs/aws_cloudwatch_metric_alarm_invalid_unit.md",
				},
			},
		},
		{
			Name: "Lowercase is invalid",
			Src: `
resource "aws_cloudwatch_metric_alarm" "test" {
    metric_name         = "FreeableMemory"
    namespace           = "AWS/RDS"

    period    = "300"
    statistic = "Average"
    threshold = "1"
    unit      = "gigabytes"
}`,
			Issues: []*issue.Issue{
				{
					Detector: "aws_cloudwatch_metric_alarm_invalid_unit",
					Type:     "ERROR",
					Message:  "\"gigabytes\" is invalid unit.",
					Line:     9,
					File:     "test.tf",
					Link:     "https://github.com/wata727/tflint/blob/master/docs/aws_cloudwatch_metric_alarm_invalid_unit.md",
				},
			},
		},
		{
			Name: "Gigabytes is valid",
			Src: `
resource "aws_cloudwatch_metric_alarm" "test" {
    metric_name         = "FreeableMemory"
    namespace           = "AWS/RDS"

    period    = "300"
    statistic = "Average"
    threshold = "1"
    unit      = "Gigabytes"
}`,
			Issues: []*issue.Issue{},
		},
	}

	for _, tc := range cases {
		var issues = []*issue.Issue{}
		err := TestDetectByCreatorName(
			"CreateAwsCloudWatchMetricAlarmInvalidUnitDetector",
			tc.Src,
			"",
			config.Init(),
			config.Init().NewAwsClient(),
			&issues,
		)
		if err != nil {
			t.Fatalf("\nERROR: %s", err)
		}

		if !reflect.DeepEqual(issues, tc.Issues) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(issues), pp.Sprint(tc.Issues), tc.Name)
		}
	}
}
