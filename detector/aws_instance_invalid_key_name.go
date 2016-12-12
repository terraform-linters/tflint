package detector

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/wata727/tflint/issue"
)

type AwsInstanceInvalidKeyNameDetector struct {
	*Detector
}

func (d *Detector) CreateAwsInstanceInvalidKeyNameDetector() *AwsInstanceInvalidKeyNameDetector {
	return &AwsInstanceInvalidKeyNameDetector{d}
}

func (d *AwsInstanceInvalidKeyNameDetector) Detect(issues *[]*issue.Issue) {
	if !d.Config.DeepCheck {
		d.Logger.Info("skip this rule.")
		return
	}

	validKeyNames := map[string]bool{}
	if d.ResponseCache.DescribeKeyPairsOutput == nil {
		resp, err := d.AwsClient.Ec2.DescribeKeyPairs(&ec2.DescribeKeyPairsInput{})
		if err != nil {
			d.Logger.Error(err)
			d.Error = true
		}
		d.ResponseCache.DescribeKeyPairsOutput = resp
	}
	for _, keyPair := range d.ResponseCache.DescribeKeyPairsOutput.KeyPairs {
		validKeyNames[*keyPair.KeyName] = true
	}

	for filename, list := range d.ListMap {
		for _, item := range list.Filter("resource", "aws_instance").Items {
			keyNameToken, err := hclLiteralToken(item, "key_name")
			if err != nil {
				d.Logger.Error(err)
				continue
			}
			keyName, err := d.evalToString(keyNameToken.Text)
			if err != nil {
				d.Logger.Error(err)
				continue
			}

			if !validKeyNames[keyName] {
				issue := &issue.Issue{
					Type:    "ERROR",
					Message: fmt.Sprintf("\"%s\" is invalid key name.", keyName),
					Line:    keyNameToken.Pos.Line,
					File:    filename,
				}
				*issues = append(*issues, issue)
			}
		}
	}
}
