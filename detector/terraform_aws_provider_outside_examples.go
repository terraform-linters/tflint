package detector

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type TerraformAWSProviderOutsideExamplesDetector struct {
	*Detector
}

func (d *Detector) CreateTerraformAWSProviderOutsideExamplesDetector() *TerraformAWSProviderOutsideExamplesDetector {
	nd := &TerraformAWSProviderOutsideExamplesDetector{Detector: d}
	nd.Name = "terraform_aws_provider_outside_examples"
	nd.IssueType = issue.ERROR
	nd.TargetType = "provider"
	nd.DeepCheck = false
	nd.Link = "https://github.com/wata727/tflint/blob/master/docs/terraform_aws_provider_outside_examples.md"
	nd.Enabled = true
	return nd
}

func (d *TerraformAWSProviderOutsideExamplesDetector) Detect(provider *schema.Provider, issues *[]*issue.Issue) {
	if (provider.Type != "aws") {
		return nil;
	}

	filename := strings.ToLower(provider.Source.Filename)
	if (!strings.Contains(filename, "examples")) {
		issue := &issue.Issue{
			Detector: d.Name,
			Type:     d.IssueType,
			Message:  fmt.Sprintf("AWS Provider in non-example directory of remote module: \"%s\" Will probably result in resources being deployed inappropriately", provider.Source.Filename),
			Line:     sourceToken.Pos.Line,
			File:     sourceToken.Pos.Filename,
			Link:     d.Link,
		}
		*issues = append(*issues, issue)
	}
}

