package test

import (
	"github.com/gruntwork-io/terratest"
	"testing"
	"os"
	terralog "github.com/gruntwork-io/terratest/log"
	"log"
	"github.com/gruntwork-io/terratest/util"
	"time"
	"fmt"
	"github.com/hashicorp/consul/api"
	"path/filepath"
	"errors"
)

const REPO_ROOT = "../"
const CONSUL_CLUSTER_EXAMPLE_REL_PATH = "examples/consul-cluster"
const CONSUL_CLUSTER_EXAMPLE_VAR_AMI_ID = "ami_id"
const CONSUL_CLUSTER_EXAMPLE_VAR_AWS_REGION = "aws_region"
const CONSUL_CLUSTER_EXAMPLE_VAR_CLUSTER_NAME = "cluster_name"
const CONSUL_CLUSTER_EXAMPLE_VAR_NUM_SERVERS = "num_servers"
const CONSUL_CLUSTER_EXAMPLE_VAR_NUM_CLIENTS = "num_clients"

const CONSUL_CLUSTER_EXAMPLE_DEFAULT_NUM_SERVERS = 3
const CONSUL_CLUSTER_EXAMPLE_DEFAULT_NUM_CLIENTS = 6

const CONSUL_CLUSTER_EXAMPLE_OUTPUT_SERVER_ASG_NAME = "asg_name_servers"
const CONSUL_CLUSTER_EXAMPLE_OUTPUT_CLIENT_ASG_NAME = "asg_name_clients"

const CONSUL_AMI_EXAMPLE_PATH = "../examples/consul-ami/consul.json"

// Test the consul-cluster example by:
//
// 1. Copying the code in this repo to a temp folder so tests on the Terraform code can run in parallel without the
//    state files overwriting each other.
// 2. Building the AMI in the consul-ami example with the given build name
// 3. Deploying that AMI using the consul-cluster Terraform code
// 4. Checking that the Consul cluster comes up within a reasonable time period and can respond to requests
func runConsulClusterTest(t *testing.T, testName string, packerBuildName string) {
	rootTempPath := copyRepoToTempFolder(t, REPO_ROOT)
	defer os.RemoveAll(rootTempPath)

	resourceCollection := createBaseRandomResourceCollection(t)
	terratestOptions := createBaseTerratestOptions(t, testName, filepath.Join(rootTempPath, CONSUL_CLUSTER_EXAMPLE_REL_PATH), resourceCollection)
	defer terratest.Destroy(terratestOptions, resourceCollection)

	logger := terralog.NewLogger(testName)
	amiId := buildAmi(t, CONSUL_AMI_EXAMPLE_PATH, packerBuildName, resourceCollection, logger)

	terratestOptions.Vars = map[string]interface{} {
		CONSUL_CLUSTER_EXAMPLE_VAR_AWS_REGION: resourceCollection.AwsRegion,
		CONSUL_CLUSTER_EXAMPLE_VAR_CLUSTER_NAME: testName + resourceCollection.UniqueId,
		CONSUL_CLUSTER_EXAMPLE_VAR_NUM_SERVERS: CONSUL_CLUSTER_EXAMPLE_DEFAULT_NUM_SERVERS,
		CONSUL_CLUSTER_EXAMPLE_VAR_NUM_CLIENTS: CONSUL_CLUSTER_EXAMPLE_DEFAULT_NUM_CLIENTS,
		CONSUL_CLUSTER_EXAMPLE_VAR_AMI_ID: amiId,
	}

	deploy(t, terratestOptions)

	// Check the Consul servers
	checkConsulClusterIsWorking(t, CONSUL_CLUSTER_EXAMPLE_OUTPUT_SERVER_ASG_NAME, terratestOptions, resourceCollection, logger)

	// Check the Consul clients
	checkConsulClusterIsWorking(t, CONSUL_CLUSTER_EXAMPLE_OUTPUT_CLIENT_ASG_NAME, terratestOptions, resourceCollection, logger)
}

// Check that the Consul cluster comes up within a reasonable time period and can respond to requests
func checkConsulClusterIsWorking(t *testing.T, asgNameOutputVar string, terratestOptions *terratest.TerratestOptions, resourceCollection *terratest.RandomResourceCollection, logger *log.Logger) {
	asgName, err := terratest.Output(terratestOptions, asgNameOutputVar)
	if err != nil {
		t.Fatalf("Could not read output %s due to error: %v", asgNameOutputVar, err)
	}

	nodeIpAddress := getIpAddressOfAsgInstance(t, asgName, resourceCollection.AwsRegion)
	testConsulCluster(t, nodeIpAddress, logger)
}

// Use a Consul client to connect to the given node and use it to verify that:
//
// 1. The Consul cluster has deployed
// 2. The cluster has the expected number of members
// 3. The cluster has elected a leader
func testConsulCluster(t *testing.T, nodeIpAddress string, logger *log.Logger) {
	consulClient := createConsulClient(t, nodeIpAddress)
	maxRetries := 60
	sleepBetweenRetries := 10 * time.Second
	expectedMembers := CONSUL_CLUSTER_EXAMPLE_DEFAULT_NUM_CLIENTS + CONSUL_CLUSTER_EXAMPLE_DEFAULT_NUM_SERVERS

	leader, err := util.DoWithRetry("Check Consul members", maxRetries, sleepBetweenRetries, logger, func() (string, error) {
		members, err := consulClient.Agent().Members(false)
		if err != nil {
			return "", err
		}

		if len(members) != expectedMembers {
			return "", fmt.Errorf("Expected the cluster to have %d members, but found %d", expectedMembers, len(members))
		}

		leader, err := consulClient.Status().Leader()
		if err != nil {
			return "", err
		}

		if leader == "" {
			return "", errors.New("Consul cluster returned an empty leader response, so a leader must not have been elected yet.")
		}

		return leader, nil
	})

	if err != nil {
		t.Fatalf("Could not verify Consul node at %s was working: %v", nodeIpAddress, err)
	}

	logger.Printf("Consul cluster is properly deployed and has elected leader %s", leader)
}

// Create a Consul client
func createConsulClient(t *testing.T, ipAddress string) *api.Client {
	config := api.DefaultConfig()
	config.Address = fmt.Sprintf("%s:8500", ipAddress)
	config.HttpClient.Timeout = 5 * time.Second

	client, err := api.NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create Consul client due to error: %v", err)
	}

	return client
}
