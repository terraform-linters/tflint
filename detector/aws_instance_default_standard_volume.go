package detector

import (
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsInstanceDefaultStandardVolumeDetector struct {
	*Detector
}

func (d *Detector) CreateAwsInstanceDefaultStandardVolumeDetector() *AwsInstanceDefaultStandardVolumeDetector {
	nd := &AwsInstanceDefaultStandardVolumeDetector{Detector: d}
	nd.Name = "aws_instance_default_standard_volume"
	nd.IssueType = issue.WARNING
	nd.TargetType = "resource"
	nd.Target = "aws_instance"
	nd.DeepCheck = false
	nd.Link = "https://github.com/wata727/tflint/blob/master/docs/aws_instance_default_standard_volume.md"
	return nd
}

func (d *AwsInstanceDefaultStandardVolumeDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	var devices []string = []string{"root_block_device", "ebs_block_device"}

	for _, device := range devices {
		if deviceTokens, ok := resource.GetAllMapTokens(device); ok {
			for i, deviceToken := range deviceTokens {
				if deviceToken["volume_type"].Text == "" {
					issue := &issue.Issue{
						Detector: d.Name,
						Type:     d.IssueType,
						Message:  "\"volume_type\" is not specified. Default standard volume type is not recommended. You can use \"gp2\", \"io1\", etc instead.",
						Line:     resource.Attrs[device].Poses[i].Line,
						File:     resource.Attrs[device].Poses[i].Filename,
						Link:     d.Link,
					}
					*issues = append(*issues, issue)
				}
			}
		}
	}
}
