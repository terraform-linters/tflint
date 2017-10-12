package schema

import (
	"testing"

	"os"
	"path/filepath"
	"reflect"

	"github.com/hashicorp/hcl/hcl/token"
	"github.com/k0kubun/pp"
)

func TestLoad(t *testing.T) {
	cases := []struct {
		Name   string
		Input  string
		Result []*Template
		Error  bool
	}{
		{
			Name: "init module with v0.10.5",
			Input: `
module "ec2_instance" {
	source = "./tf_aws_ec2_instance"
}`,
			Result: []*Template{
				{
					File: "./tf_aws_ec2_instance/test.tf",
					Resources: []*Resource{
						{
							Source: &Source{
								File: "./tf_aws_ec2_instance/test.tf",
								Pos: token.Pos{
									Filename: "./tf_aws_ec2_instance/test.tf",
									Offset:   30,
									Line:     1,
									Column:   31,
								},
								Attrs: map[string]*Attribute{
									"instance_type": {
										Poses: []token.Pos{
											{
												Filename: "./tf_aws_ec2_instance/test.tf",
												Offset:   83,
												Line:     3,
												Column:   19,
											},
										},
										Vals: []interface{}{
											token.Token{
												Type: 9,
												Pos: token.Pos{
													Filename: "./tf_aws_ec2_instance/test.tf",
													Offset:   83,
													Line:     3,
													Column:   19,
												},
												Text: "\"t1.2xlarge\"",
												JSON: false,
											},
										},
									},
									"ami": {
										Poses: []token.Pos{
											{
												Filename: "./tf_aws_ec2_instance/test.tf",
												Offset:   50,
												Line:     2,
												Column:   19,
											},
										},
										Vals: []interface{}{
											token.Token{
												Type: 9,
												Pos: token.Pos{
													Filename: "./tf_aws_ec2_instance/test.tf",
													Offset:   50,
													Line:     2,
													Column:   19,
												},
												Text: "\"ami-12345678\"",
												JSON: false,
											},
										},
									},
								},
							},
							Type: "aws_instance",
							Id:   "web",
						},
					},
				},
			},
			Error: false,
		},
		{
			Name: "init module with v0.10.6",
			Input: `
module "ec2_instance" {
	source = "./tf_aws_ec2_instance_v2"
}`,
			Result: []*Template{
				{
					File: "./tf_aws_ec2_instance_v2/test.tf",
					Resources: []*Resource{
						{
							Source: &Source{
								File: "./tf_aws_ec2_instance_v2/test.tf",
								Pos: token.Pos{
									Filename: "./tf_aws_ec2_instance_v2/test.tf",
									Offset:   30,
									Line:     1,
									Column:   31,
								},
								Attrs: map[string]*Attribute{
									"instance_type": {
										Poses: []token.Pos{
											{
												Filename: "./tf_aws_ec2_instance_v2/test.tf",
												Offset:   83,
												Line:     3,
												Column:   19,
											},
										},
										Vals: []interface{}{
											token.Token{
												Type: 9,
												Pos: token.Pos{
													Filename: "./tf_aws_ec2_instance_v2/test.tf",
													Offset:   83,
													Line:     3,
													Column:   19,
												},
												Text: "\"t2.micro\"",
												JSON: false,
											},
										},
									},
									"ami": {
										Poses: []token.Pos{
											{
												Filename: "./tf_aws_ec2_instance_v2/test.tf",
												Offset:   50,
												Line:     2,
												Column:   19,
											},
										},
										Vals: []interface{}{
											token.Token{
												Type: 9,
												Pos: token.Pos{
													Filename: "./tf_aws_ec2_instance_v2/test.tf",
													Offset:   50,
													Line:     2,
													Column:   19,
												},
												Text: "\"ami-abcd1234\"",
												JSON: false,
											},
										},
									},
								},
							},
							Type: "aws_instance",
							Id:   "web",
						},
					},
				},
			},
			Error: false,
		},
		{
			Name: "init module with v0.10.7",
			Input: `
module "ec2_instance" {
	source = "./tf_aws_ec2_instance_v3"
}`,
			Result: []*Template{
				{
					File: "./tf_aws_ec2_instance_v3/test.tf",
					Resources: []*Resource{
						{
							Source: &Source{
								File: "./tf_aws_ec2_instance_v3/test.tf",
								Pos: token.Pos{
									Filename: "./tf_aws_ec2_instance_v3/test.tf",
									Offset:   30,
									Line:     1,
									Column:   31,
								},
								Attrs: map[string]*Attribute{
									"instance_type": {
										Poses: []token.Pos{
											{
												Filename: "./tf_aws_ec2_instance_v3/test.tf",
												Offset:   83,
												Line:     3,
												Column:   19,
											},
										},
										Vals: []interface{}{
											token.Token{
												Type: 9,
												Pos: token.Pos{
													Filename: "./tf_aws_ec2_instance_v3/test.tf",
													Offset:   83,
													Line:     3,
													Column:   19,
												},
												Text: "\"m3.large\"",
												JSON: false,
											},
										},
									},
									"ami": {
										Poses: []token.Pos{
											{
												Filename: "./tf_aws_ec2_instance_v3/test.tf",
												Offset:   50,
												Line:     2,
												Column:   19,
											},
										},
										Vals: []interface{}{
											token.Token{
												Type: 9,
												Pos: token.Pos{
													Filename: "./tf_aws_ec2_instance_v3/test.tf",
													Offset:   50,
													Line:     2,
													Column:   19,
												},
												Text: "\"ami-9876abcd\"",
												JSON: false,
											},
										},
									},
								},
							},
							Type: "aws_instance",
							Id:   "web",
						},
					},
				},
			},
			Error: false,
		},
		{
			Name: "module not found",
			Input: `
module "ec2_instances" {
	source = "unresolvable_module_source"
}`,
			Result: nil,
			Error:  true,
		},
	}

	for _, tc := range cases {
		prev, _ := filepath.Abs(".")
		dir, _ := os.Getwd()
		defer os.Chdir(prev)
		testDir := dir + "/test-fixtures"
		os.Chdir(testDir)
		files := map[string][]byte{"test.tf": []byte(tc.Input)}
		schema, _ := Make(files)

		module := schema[0].Modules[0]
		err := module.Load()

		if tc.Error && err == nil {
			t.Fatalf("\nshould be happen error.\n\ntestcase: %s", tc.Name)
			continue
		}
		if !tc.Error && err != nil {
			t.Fatalf("\nshould not be happen error.\nError: %s\n\ntestcase: %s", err, tc.Name)
			continue
		}

		if !reflect.DeepEqual(module.Templates, tc.Result) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(module.Templates), pp.Sprint(tc.Result), tc.Name)
		}
	}
}
