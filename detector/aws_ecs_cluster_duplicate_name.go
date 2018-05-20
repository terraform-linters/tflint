package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsECSClusterDuplicateNameDetector struct {
	*Detector
	clusters map[string]bool
}

func (d *Detector) CreateAwsECSClusterDuplicateNameDetector() *AwsECSClusterDuplicateNameDetector {
	nd := &AwsECSClusterDuplicateNameDetector{
		Detector: d,
		clusters: map[string]bool{},
	}
	nd.Name = "aws_ecs_cluster_duplicate_name"
	nd.IssueType = issue.ERROR
	nd.TargetType = "resource"
	nd.Target = "aws_ecs_cluster"
	nd.DeepCheck = true
	nd.Enabled = true
	return nd
}

func (d *AwsECSClusterDuplicateNameDetector) PreProcess() {
	resp, err := d.AwsClient.DescribeClusters()
	if err != nil {
		d.Logger.Error(err)
		d.Error = true
		return
	}

	for _, cluster := range resp.Clusters {
		d.clusters[*cluster.ClusterName] = true
	}
}

func (d *AwsECSClusterDuplicateNameDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	nameToken, ok := resource.GetToken("name")
	if !ok {
		return
	}
	name, err := d.evalToString(nameToken.Text)
	if err != nil {
		d.Logger.Error(err)
		return
	}

	identityCheckFunc := func(attributes map[string]string) bool { return attributes["name"] == name }
	if d.clusters[name] && !d.State.Exists(d.Target, resource.Id, identityCheckFunc) {
		issue := &issue.Issue{
			Detector: d.Name,
			Type:     d.IssueType,
			Message:  fmt.Sprintf("\"%s\" is duplicate name. It must be unique.", name),
			Line:     nameToken.Pos.Line,
			File:     nameToken.Pos.Filename,
		}
		*issues = append(*issues, issue)
	}
}
