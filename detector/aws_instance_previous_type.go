package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
)

func (d *Detector) DetectAwsInstancePreviousType(issues *[]*issue.Issue) {
	var previousInstanceType = map[string]bool{
		"t1.micro":    true,
		"m1.small":    true,
		"m1.medium":   true,
		"m1.large":    true,
		"m1.xlarge":   true,
		"c1.medium":   true,
		"c1.xlarge":   true,
		"cc2.8xlarge": true,
		"cg1.4xlarge": true,
		"m2.xlarge":   true,
		"m2.2xlarge":  true,
		"m2.4xlarge":  true,
		"cr1.8xlarge": true,
		"hi1.4xlarge": true,
		"hs1.8xlarge": true,
	}

	for filename, list := range d.ListMap {
		for _, item := range list.Filter("resource", "aws_instance").Items {
			instanceTypeToken, err := hclLiteralToken(item, "instance_type")
			if err != nil {
				continue
			}
			instanceType, err := d.evalToString(instanceTypeToken.Text)
			if err != nil {
				continue
			}

			if previousInstanceType[instanceType] {
				issue := &issue.Issue{
					Type:    "NOTICE",
					Message: fmt.Sprintf("\"%s\" is previous generation instance type.", instanceType),
					Line:    instanceTypeToken.Pos.Line,
					File:    filename,
				}
				*issues = append(*issues, issue)
			}
		}
	}
}
