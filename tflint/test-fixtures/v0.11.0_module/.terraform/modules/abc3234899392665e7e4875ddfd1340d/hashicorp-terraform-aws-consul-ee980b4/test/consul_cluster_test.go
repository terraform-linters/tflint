package test

import (
	"testing"
)

func TestConsulClusterWithUbuntuAmi(t *testing.T) {
	t.Parallel()
	runConsulClusterTest(t, "TestConsulUbuntu", "ubuntu16-ami")
}

func TestConsulClusterWithAmazonLinuxAmi(t *testing.T) {
	t.Parallel()
	runConsulClusterTest(t, "TestConsulAmznLnx", "amazon-linux-ami")
}

