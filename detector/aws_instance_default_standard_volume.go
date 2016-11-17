package detector

import (
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/wata727/tflint/issue"
)

func (d *Detector) DetectAwsInstanceDefaultStandardVolume(issues *[]*issue.Issue) {
	for filename, list := range d.ListMap {
		for _, item := range list.Filter("resource", "aws_instance").Items {
			d.detectForBlockDevices(issues, item, filename, "root_block_device")
			d.detectForBlockDevices(issues, item, filename, "ebs_block_device")
		}
	}
}

func (d *Detector) detectForBlockDevices(issues *[]*issue.Issue, item *ast.ObjectItem, filename string, device string) {
	if !IsKeyNotFound(item, device) {
		deviceItems, err := hclObjectItems(item, device)
		if err != nil {
			d.Logger.Error(err)
			return
		}

		for _, deviceItem := range deviceItems {
			if IsKeyNotFound(deviceItem, "volume_type") {
				issue := &issue.Issue{
					Type:    "NOTICE",
					Message: "\"volume_type\" is not specified. Default standard volume type is not recommended. You should use \"gp2\"",
					Line:    deviceItem.Assign.Line,
					File:    filename,
				}
				*issues = append(*issues, issue)
			}
		}
	}
}
