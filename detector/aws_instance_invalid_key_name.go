package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsInstanceInvalidKeyNameDetector struct {
	*Detector
	keypairs map[string]bool
}

func (d *Detector) CreateAwsInstanceInvalidKeyNameDetector() *AwsInstanceInvalidKeyNameDetector {
	nd := &AwsInstanceInvalidKeyNameDetector{
		Detector: d,
		keypairs: map[string]bool{},
	}
	nd.Name = "aws_instance_invalid_key_name"
	nd.IssueType = issue.ERROR
	nd.TargetType = "resource"
	nd.Target = "aws_instance"
	nd.DeepCheck = true
	return nd
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

func (d *AwsInstanceInvalidKeyNameDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	keyNameToken, ok := resource.GetToken("key_name")
	if !ok {
		return
	}
	keyName, err := d.evalToString(keyNameToken.Text)
	if err != nil {
		d.Logger.Error(err)
		return
	}

	if !d.keypairs[keyName] {
		issue := &issue.Issue{
			Type:    d.IssueType,
			Message: fmt.Sprintf("\"%s\" is invalid key name.", keyName),
			Line:    keyNameToken.Pos.Line,
			File:    keyNameToken.Pos.Filename,
		}
		*issues = append(*issues, issue)
	}
}
