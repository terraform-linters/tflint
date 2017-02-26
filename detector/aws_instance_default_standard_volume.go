package detector

import (
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/wata727/tflint/issue"
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

func (d *AwsInstanceDefaultStandardVolumeDetector) Detect(file string, item *ast.ObjectItem, issues *[]*issue.Issue) {
	d.detectForBlockDevices(issues, item, file, "root_block_device")
	d.detectForBlockDevices(issues, item, file, "ebs_block_device")
}

func (d *AwsInstanceDefaultStandardVolumeDetector) detectForBlockDevices(issues *[]*issue.Issue, item *ast.ObjectItem, file string, device string) {
	if !IsKeyNotFound(item, device) {
		deviceItems, err := hclObjectItems(item, device)
		if err != nil {
			d.Logger.Error(err)
			return
		}

		for _, deviceItem := range deviceItems {
			if IsKeyNotFound(deviceItem, "volume_type") {
				issue := &issue.Issue{
					Type:    d.IssueType,
					Message: "\"volume_type\" is not specified. Default standard volume type is not recommended. You can use \"gp2\", \"io1\", etc instead.",
					Line:    deviceItem.Assign.Line,
					File:    file,
				}
				*issues = append(*issues, issue)
			}
		}
	}
}
