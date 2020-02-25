package awsrules

import (
	"io/ioutil"
	"os"
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_AwsInstanceWithTags(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Config   string
		Expected tflint.Issues
	}{
		{
			Name: "Wanted tags: Bar,Foo, found: bar,foo",
			Content: `
resource "aws_instance" "ec2_instance" {
    instance_type = "t2.micro"
    tags = {
      foo = "bar"
      bar = "baz"
    }
}`,
			Config: `
rule "aws_resource_missing_tags" {
  enabled = true
  tags = ["Foo", "Bar"]
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewAwsResourceMissingTagsRule(),
					Message: "The resource is missing the following tags: \"Bar\", \"Foo\".",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 4, Column: 12},
						End:      hcl.Pos{Line: 7, Column: 6},
					},
				},
			},
		},
		{
			Name: "Tags are correct",
			Content: `
resource "aws_instance" "ec2_instance" {
    instance_type = "t2.micro"
    tags = {
      Foo = "bar"
      Bar = "baz"
    }
}`,
			Config: `
rule "aws_resource_missing_tags" {
  enabled = true
  tags = ["Foo", "Bar"]
}`,
			Expected: tflint.Issues{},
		},
	}

	rule := NewAwsResourceMissingTagsRule()

	for _, tc := range cases {
		runner := tflint.TestRunnerWithConfig(t, map[string]string{"module.tf": tc.Content}, loadConfigfromTempFile(t, tc.Config))

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}

// TODO: Replace with TestRunner
func loadConfigfromTempFile(t *testing.T, content string) *tflint.Config {
	if content == "" {
		return tflint.EmptyConfig()
	}

	tmpfile, err := ioutil.TempFile("", "aws_resource_missing_tags")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	config, err := tflint.LoadConfig(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	return config
}
