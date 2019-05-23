package test

import (
	"testing"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// Get the IP address from a randomly chosen EC2 Instance in an Auto Scaling Group of the given name in the given
// region
func getIpAddressOfAsgInstance(t *testing.T, asgName string, awsRegion string) string {
	instanceId := getIdOfAsgInstance(t, asgName, awsRegion)
	return getPublicIpOfInstance(t, instanceId, awsRegion)
}

// Get the ID of a a randomly chosen EC2 Instance in an Auto Scaling Group of the given name in the given region
func getIdOfAsgInstance(t *testing.T, asgName string, awsRegion string) string {
	autoscalingClient := createAutoscalingClient(t, awsRegion)

	input := autoscaling.DescribeAutoScalingGroupsInput{AutoScalingGroupNames: []*string{aws.String(asgName)}}
	output, err := autoscalingClient.DescribeAutoScalingGroups(&input)
	if err != nil {
		t.Fatalf("Failed to call DescribeAutoScalingGroupsInput API due to error: %v", err)
	}

	for _, asg := range output.AutoScalingGroups {
		for _, instance := range asg.Instances {
			return *instance.InstanceId
		}
	}

	t.Fatalf("Could not find any Instance Ids for ASG %s: %v", asgName, output)
	return ""
}

// Get the public IP address of the given EC2 Instnace in the given region
func getPublicIpOfInstance(t *testing.T, instanceId string, awsRegion string) string {
	ec2Client := createEc2Client(t, awsRegion)

	input := ec2.DescribeInstancesInput{InstanceIds: []*string{aws.String(instanceId)}}
	output, err := ec2Client.DescribeInstances(&input)
	if err != nil {
		t.Fatalf("Failed to fetch information about EC2 Instance %s due to error: %v", instanceId, err)
	}

	for _, reserveration := range output.Reservations {
		for _, instance := range reserveration.Instances {
			return *instance.PublicIpAddress
		}
	}

	t.Fatalf("Failed to find public IP address for EC2 Instance %s: %v", instanceId, output)
	return ""
}

// Create a client that can be used to make EC2 API calls
func createEc2Client(t *testing.T, awsRegion string) *ec2.EC2 {
	awsConfig := createAwsConfig(t, awsRegion)
	return ec2.New(session.New(), awsConfig)
}

// Create a client that can be used to make Auto Scaling API calls
func createAutoscalingClient(t *testing.T, awsRegion string) *autoscaling.AutoScaling {
	awsConfig := createAwsConfig(t, awsRegion)
	return autoscaling.New(session.New(), awsConfig)
}

// Create an AWS config. This method will check for credentials and fail the test if it can't find them.
func createAwsConfig(t *testing.T, awsRegion string) *aws.Config {
	config := defaults.Get().Config.WithRegion(awsRegion)

	_, err := config.Credentials.Get()
	if err != nil {
		t.Fatalf("Error finding AWS credentials (did you set the AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables?). Underlying error: %v", err)
	}

	return config
}
