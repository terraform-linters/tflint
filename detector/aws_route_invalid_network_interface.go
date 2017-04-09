package detector

import (
	"fmt"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/wata727/tflint/issue"
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

func (d *AwsRouteInvalidNetworkInterfaceDetector) Detect(file string, item *ast.ObjectItem, issues *[]*issue.Issue) {
	networkInterfaceToken, err := hclLiteralToken(item, "network_interface_id")
	if err != nil {
		d.Logger.Error(err)
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
			File:    file,
		}
		*issues = append(*issues, issue)
	}
}
