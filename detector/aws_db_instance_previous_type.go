package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
)

type AwsDBInstancePreviousTypeDetector struct {
	*Detector
}

func (d *Detector) CreateAwsDBInstancePreviousTypeDetector() *AwsDBInstancePreviousTypeDetector {
	return &AwsDBInstancePreviousTypeDetector{d}
}

func (d *AwsDBInstancePreviousTypeDetector) Detect(issues *[]*issue.Issue) {
	var previousInstanceType = map[string]bool{
		"db.t1.micro":    true,
		"db.m1.small":    true,
		"db.m1.medium":   true,
		"db.m1.large":    true,
		"db.m1.xlarge":   true,
		"db.m2.xlarge":   true,
		"db.m2.2xlarge":  true,
		"db.m2.4xlarge":  true,
		"db.cr1.8xlarge": true,
	}

	for filename, list := range d.ListMap {
		for _, item := range list.Filter("resource", "aws_db_instance").Items {
			instanceTypeToken, err := hclLiteralToken(item, "instance_class")
			if err != nil {
				d.Logger.Error(err)
				continue
			}
			instanceType, err := d.evalToString(instanceTypeToken.Text)
			if err != nil {
				d.Logger.Error(err)
				continue
			}

			if previousInstanceType[instanceType] {
				issue := &issue.Issue{
					Type:    "WARNING",
					Message: fmt.Sprintf("\"%s\" is previous generation instance type.", instanceType),
					Line:    instanceTypeToken.Pos.Line,
					File:    filename,
				}
				*issues = append(*issues, issue)
			}
		}
	}
}
