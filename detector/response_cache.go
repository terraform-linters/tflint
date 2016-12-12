package detector

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
)

type ResponseCache struct {
	DescribeImagesOutput       *ec2.DescribeImagesOutput
	DescribeKeyPairsOutput     *ec2.DescribeKeyPairsOutput
	DescribeSubnetsOutput      *ec2.DescribeSubnetsOutput
	ListInstanceProfilesOutput *iam.ListInstanceProfilesOutput
}
