package rules

import (
	"testing"

	"github.com/terraform-linters/tflint-plugin-sdk/helper"
	"github.com/terraform-linters/tflint-ruleset-terraform/terraform"
)

func testRunner(t *testing.T, files map[string]string) *terraform.Runner {
	return terraform.NewRunner(helper.TestRunner(t, files))
}
