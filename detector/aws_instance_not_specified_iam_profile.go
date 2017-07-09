package detector

import (
	"github.com/hashicorp/go-version"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsInstanceNotSpecifiedIAMProfileDetector struct {
	*Detector
}

func (d *Detector) CreateAwsInstanceNotSpecifiedIAMProfileDetector() *AwsInstanceNotSpecifiedIAMProfileDetector {
	nd := &AwsInstanceNotSpecifiedIAMProfileDetector{Detector: d}
	nd.Name = "aws_instance_not_specified_iam_profile"
	nd.IssueType = issue.NOTICE
	nd.TargetType = "resource"
	nd.Target = "aws_instance"
	nd.DeepCheck = false
	nd.Link = "https://github.com/wata727/tflint/blob/master/docs/aws_instance_not_specified_iam_profile.md"
	return nd
}

func (d *AwsInstanceNotSpecifiedIAMProfileDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	v1, err := version.NewVersion(d.Config.TerraformVersion)
	// If `terraform_version` is not set, always detect.
	if err != nil {
		v1, _ = version.NewVersion("0.8.0")
	}
	v2, _ := version.NewVersion("0.8.8")

	if _, ok := resource.GetToken("iam_instance_profile"); !ok && v1.LessThan(v2) {
		issue := &issue.Issue{
			Detector: d.Name,
			Type:     d.IssueType,
			Message:  "\"iam_instance_profile\" is not specified. If you want to change it, you need to recreate the instance.",
			Line:     resource.Pos.Line,
			File:     resource.Pos.Filename,
			Link:     d.Link,
		}
		*issues = append(*issues, issue)
	}
}
