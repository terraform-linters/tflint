package detector

import (
	"fmt"

	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/schema"
)

type AwsRouteInvalidNetworkInterfaceDetector struct {
	*Detector
	IssueType         string
	Target            string
	DeepCheck         bool
	networkInterfaces map[string]bool
}

func (d *Detector) CreateAwsRouteInvalidNetworkInterfaceDetector() *AwsRouteInvalidNetworkInterfaceDetector {
	return &AwsRouteInvalidNetworkInterfaceDetector{
		Detector:          d,
		IssueType:         issue.ERROR,
		Target:            "aws_route",
		DeepCheck:         true,
		networkInterfaces: map[string]bool{},
	}
}

func (d *AwsRouteInvalidNetworkInterfaceDetector) PreProcess() {
	resp, err := d.AwsClient.DescribeNetworkInterfaces()
	if err != nil {
		d.Logger.Error(err)
		d.Error = true
		return
	}

	for _, networkInterface := range resp.NetworkInterfaces {
		d.networkInterfaces[*networkInterface.NetworkInterfaceId] = true
	}
}

func (d *AwsRouteInvalidNetworkInterfaceDetector) Detect(resource *schema.Resource, issues *[]*issue.Issue) {
	networkInterfaceToken, ok := resource.GetToken("network_interface_id")
	if !ok {
		return
	}
	networkInterface, err := d.evalToString(networkInterfaceToken.Text)
	if err != nil {
		d.Logger.Error(err)
		return
	}

	if !d.networkInterfaces[networkInterface] {
		issue := &issue.Issue{
			Type:    d.IssueType,
			Message: fmt.Sprintf("\"%s\" is invalid network interface ID.", networkInterface),
			Line:    networkInterfaceToken.Pos.Line,
			File:    networkInterfaceToken.Pos.Filename,
		}
		*issues = append(*issues, issue)
	}
}
