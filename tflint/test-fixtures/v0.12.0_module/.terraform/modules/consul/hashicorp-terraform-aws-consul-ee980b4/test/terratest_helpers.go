package test

import (
	"github.com/gruntwork-io/terratest/packer"
	"github.com/gruntwork-io/terratest"
	"log"
	"testing"
)

const CONSUL_AMI_TEMPLATE_VAR_REGION = "aws_region"

// Deploy the given terraform code
func deploy(t *testing.T, terratestOptions *terratest.TerratestOptions) {
	_, err := terratest.Apply(terratestOptions)
	if err != nil {
		t.Fatalf("Failed to apply templates: %s", err.Error())
	}
}

// Use Packer to build the AMI in the given packer template, with the given build name, and return the AMI's ID
func buildAmi(t *testing.T, packerTemplatePath string, packerBuildName string, resourceCollection *terratest.RandomResourceCollection, logger *log.Logger) string {
	options := packer.PackerOptions{
		Template: packerTemplatePath,
		Only: packerBuildName,
		Vars: map[string]string{
			CONSUL_AMI_TEMPLATE_VAR_REGION: resourceCollection.AwsRegion,
		},
	}

	amiId, err := packer.BuildAmi(options, logger)
	if err != nil {
		t.Fatalf("Failed to build AMI for Packer template %s: %s", packerTemplatePath, err.Error())
	}
	if amiId == "" {
		t.Fatalf("Got blank AMI ID after building Packer template %s", packerTemplatePath)
	}

	return amiId
}

// Create the basic RandomResourceCollection for testing the consul-cluster example
func createBaseRandomResourceCollection(t *testing.T) *terratest.RandomResourceCollection {
	resourceCollectionOptions := terratest.NewRandomResourceCollectionOptions()

	randomResourceCollection, err := terratest.CreateRandomResourceCollection(resourceCollectionOptions)
	if err != nil {
		t.Fatalf("Failed to create Random Resource Collection: %s", err.Error())
	}

	return randomResourceCollection
}

// Create the basic TerratestOptions for testing the consul-cluster example
func createBaseTerratestOptions(t *testing.T, testName string, templatePath string, resourceCollection *terratest.RandomResourceCollection) *terratest.TerratestOptions {
	terratestOptions := terratest.NewTerratestOptions()

	terratestOptions.UniqueId = resourceCollection.UniqueId
	terratestOptions.TemplatePath = templatePath
	terratestOptions.TestName = testName

	return terratestOptions
}

