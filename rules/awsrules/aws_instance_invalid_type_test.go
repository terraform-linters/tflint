package awsrules

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/configs/configload"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

func Test_AwsInstanceInvalidType(t *testing.T) {
	cases := []struct {
		Name     string
		Dir      string
		Expected issue.Issues
	}{
		{
			Name: "basic",
			Dir:  "basic",
			Expected: []*issue.Issue{
				{
					Detector: "aws_instance_invalid_type",
					Type:     issue.ERROR,
					Message:  "\"t1.2xlarge\" is invalid instance type.",
					Line:     2,
					File:     "instances.tf",
					Link:     "https://github.com/wata727/tflint/blob/master/docs/aws_instance_invalid_type.md",
				},
			},
		},
	}

	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	loader, err := configload.NewLoader(&configload.Config{})
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range cases {
		mod, diags := loader.Parser().LoadConfigDir(dir + "/test-fixtures/aws_instance_invalid_type/" + tc.Dir)
		if diags.HasErrors() {
			panic(diags)
		}
		cfg, tfdiags := configs.BuildConfig(mod, configs.DisabledModuleWalker)
		if tfdiags.HasErrors() {
			panic(tfdiags)
		}

		runner := tflint.NewRunner(cfg)
		rule := &AwsInstanceInvalidTypeRule{}
		rule.PreProcess()
		rule.Check(runner)

		if !cmp.Equal(tc.Expected, runner.Issues) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues))
		}
	}
}
