package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
)

type AwsInstanceInvalidKeyNameDetector struct {
	*Detector
	IssueType string
	Target    string
	DeepCheck bool
	keypairs  map[string]bool
}

func (d *Detector) CreateAwsInstanceInvalidKeyNameDetector() *AwsInstanceInvalidKeyNameDetector {
	return &AwsInstanceInvalidKeyNameDetector{
		Detector:  d,
		IssueType: issue.ERROR,
		Target:    "aws_instance",
		DeepCheck: true,
		keypairs:  map[string]bool{},
	}
}

func (d *AwsInstanceInvalidKeyNameDetector) PreProcess() {
	resp, err := d.AwsClient.DescribeKeyPairs()
	if err != nil {
		d.Logger.Error(err)
		d.Error = true
		return
	}

	for _, keyPair := range resp.KeyPairs {
		d.keypairs[*keyPair.KeyName] = true
	}
}

func (d *AwsInstanceInvalidKeyNameDetector) Detect(issues *[]*issue.Issue) {
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

			if !d.keypairs[keyName] {
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
