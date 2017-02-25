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

func (d *AwsInstanceDefaultStandardVolumeDetector) Detect(issues *[]*issue.Issue) {
	for filename, list := range d.ListMap {
		for _, item := range list.Filter("resource", "aws_instance").Items {
			d.detectForBlockDevices(issues, item, filename, "root_block_device")
			d.detectForBlockDevices(issues, item, filename, "ebs_block_device")
		}
	}
}

func (d *AwsInstanceDefaultStandardVolumeDetector) detectForBlockDevices(issues *[]*issue.Issue, item *ast.ObjectItem, filename string, device string) {
	if !IsKeyNotFound(item, device) {
		deviceItems, err := hclObjectItems(item, device)
		if err != nil {
			d.Logger.Error(err)
			return
		}

		for _, deviceItem := range deviceItems {
			if IsKeyNotFound(deviceItem, "volume_type") {
				issue := &issue.Issue{
					Type:    "WARNING",
					Message: "\"volume_type\" is not specified. Default standard volume type is not recommended. You can use \"gp2\", \"io1\", etc instead.",
					Line:    deviceItem.Assign.Line,
					File:    filename,
				}
				*issues = append(*issues, issue)
			}
		}
	}
}
