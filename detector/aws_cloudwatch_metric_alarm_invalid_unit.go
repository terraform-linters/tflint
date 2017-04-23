package detector

import (
	"fmt"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/wata727/tflint/issue"
)

type AwsCloudWatchMetricAlarmInvalidUnitDetector struct {
	*Detector
	IssueType  string
	Target     string
	DeepCheck  bool
	validUnits map[string]bool
}

func (d *Detector) CreateAwsCloudWatchMetricAlarmInvalidUnitDetector() *AwsCloudWatchMetricAlarmInvalidUnitDetector {
	return &AwsCloudWatchMetricAlarmInvalidUnitDetector{
		Detector:   d,
		IssueType:  issue.ERROR,
		Target:     "aws_cloudwatch_metric_alarm",
		DeepCheck:  false,
		validUnits: map[string]bool{},
	}
}

func (d *AwsCloudWatchMetricAlarmInvalidUnitDetector) PreProcess() {
	// Ref: http://docs.aws.amazon.com/cli/latest/reference/cloudwatch/put-metric-alarm.html
	d.validUnits = map[string]bool{
		"Seconds":          true,
		"Microseconds":     true,
		"Milliseconds":     true,
		"Bytes":            true,
		"Kilobytes":        true,
		"Megabytes":        true,
		"Gigabytes":        true,
		"Terabytes":        true,
		"Bits":             true,
		"Kilobits":         true,
		"Megabits":         true,
		"Gigabits":         true,
		"Terabits":         true,
		"Percent":          true,
		"Count":            true,
		"Bytes/Second":     true,
		"Kilobytes/Second": true,
		"Megabytes/Second": true,
		"Gigabytes/Second": true,
		"Terabytes/Second": true,
		"Bits/Second":      true,
		"Kilobits/Second":  true,
		"Megabits/Second":  true,
		"Gigabits/Second":  true,
		"Terabits/Second":  true,
		"Count/Second":     true,
		"None":             true,
	}
}

func (d *AwsCloudWatchMetricAlarmInvalidUnitDetector) Detect(file string, item *ast.ObjectItem, issues *[]*issue.Issue) {
	unitToken, err := hclLiteralToken(item, "unit")
	if err != nil {
		d.Logger.Error(err)
		return
	}
	unit, err := d.evalToString(unitToken.Text)
	if err != nil {
		d.Logger.Error(err)
		return
	}

	if !d.validUnits[unit] {
		issue := &issue.Issue{
			Type:    d.IssueType,
			Message: fmt.Sprintf("\"%s\" is invalid unit.", unit),
			Line:    unitToken.Pos.Line,
			File:    file,
		}
		*issues = append(*issues, issue)
	}
}
