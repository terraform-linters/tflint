package detector

import (
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsInstanceDefaultStandardVolumeDetector struct {
	*Detector
	IssueType string
	Target    string
	DeepCheck bool
}

func (d *Detector) CreateAwsInstanceDefaultStandardVolumeDetector() *AwsInstanceDefaultStandardVolumeDetector {
	return &AwsInstanceDefaultStandardVolumeDetector{
		Detector:  d,
		IssueType: issue.WARNING,
		Target:    "aws_instance",
		DeepCheck: false,
	}
}

func (d *AwsInstanceDefaultStandardVolumeDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	var devices []string = []string{"root_block_device", "ebs_block_device"}

	for _, device := range devices {
		if deviceTokens, ok := resource.GetAllMapTokens(device); ok {
			for i, deviceToken := range deviceTokens {
				if deviceToken["volume_type"].Text == "" {
					issue := &issue.Issue{
						Type:    d.IssueType,
						Message: "\"volume_type\" is not specified. Default standard volume type is not recommended. You can use \"gp2\", \"io1\", etc instead.",
						Line:    resource.Attrs[device].Poses[i].Line,
						File:    resource.Attrs[device].Poses[i].Filename,
					}
					*issues = append(*issues, issue)
				}
			}
		}
	}
}
