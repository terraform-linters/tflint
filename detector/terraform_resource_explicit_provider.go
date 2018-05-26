package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type TerraformResourceExplicitProviderDetector struct {
	*Detector
}

func (d *Detector) CreateTerraformResourceExplicitProviderDetector() *TerraformResourceExplicitProviderDetector {
	nd := &TerraformResourceExplicitProviderDetector{Detector: d}
	nd.Name = "terraform_resource_explicit_provider"
	nd.IssueType = issue.WARNING
	nd.TargetType = "resource"
	nd.Target = ""
	nd.DeepCheck = false
	nd.Link = "https://github.com/wata727/tflint/blob/master/docs/terraform_resource_explicit_provider.md"
	nd.Enabled = false
	return nd
}

func (d *TerraformResourceExplicitProviderDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	_, ok := resource.GetToken("provider")
	if !ok {
		issue := &issue.Issue{
			Detector: d.Name,
			Type:     d.IssueType,
			Message:  fmt.Sprintf("Resource \"%s\" provider is implicit", resource.Id),
			Line:     resource.Pos.Line,
			File:     resource.Pos.Filename,
			Link:     d.Link,
		}
		*issues = append(*issues, issue)
	}
}
