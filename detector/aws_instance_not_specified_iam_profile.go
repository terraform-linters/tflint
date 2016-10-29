package detector

import "github.com/wata727/tflint/issue"

func (d *Detector) DetectAwsInstanceNotSpecifiedIamProfile(issues *[]*issue.Issue) {
	for filename, list := range d.ListMap {
		for _, item := range list.Filter("resource", "aws_instance").Items {
			_, err := hclLiteralToken(item, "iam_instance_profile")
			if err != nil && err.Error() == "key not found" {
				issue := &issue.Issue{
					Type:    "NOTICE",
					Message: "\"iam_instance_profile\" is not specified. You cannot edit this value later.",
					Line:    item.Pos().Line,
					File:    filename,
				}
				*issues = append(*issues, issue)
			}
		}
	}
}
