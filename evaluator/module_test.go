package evaluator

import (
	"testing"

	"os"
	"path/filepath"
	"reflect"

	hclast "github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/parser"
	"github.com/hashicorp/hil"
	hilast "github.com/hashicorp/hil/ast"
	"github.com/k0kubun/pp"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/schema"
)

func TestInitModule(t *testing.T) {
	cases := []struct {
		Name   string
		Input  string
		Result hil.EvalConfig
		Error  bool
	}{
		{
			Name: "init module",
			Input: `
module "ec2_instance" {
	source = "./tf_aws_ec2_instance"
	ami = "ami-12345"
	num = "1"
}`,
			Result: hil.EvalConfig{
				GlobalScope: &hilast.BasicScope{
					VarMap: map[string]hilast.Variable{
						"var.ami": {
							Type:  hilast.TypeString,
							Value: "ami-12345",
						},
						"var.num": {
							Type:  hilast.TypeString,
							Value: "1",
						},
					},
				},
			},
			Error: false,
		},
		{
			Name: "init module with string variable",
			Input: `
variable "ami" {
    default = "ami-12345"
}

module "ec2_instance" {
    source = "./tf_aws_ec2_instance"
    ami = "${var.ami}"
}`,
			Result: hil.EvalConfig{
				GlobalScope: &hilast.BasicScope{
					VarMap: map[string]hilast.Variable{
						"var.ami": {
							Type:  hilast.TypeString,
							Value: "ami-12345",
						},
					},
				},
			},
			Error: false,
		},
		{
			Name: "init module with list variable",
			Input: `
variable "amis" {
    default = ["ami-12345", "ami-54321"]
}

module "ec2_instance" {
    source = "./tf_aws_ec2_instance"
    ami = "${var.amis}"
}`,
			Result: hil.EvalConfig{
				GlobalScope: &hilast.BasicScope{
					VarMap: map[string]hilast.Variable{
						"var.ami": {
							Type: hilast.TypeList,
							Value: []hilast.Variable{
								{
									Type:  hilast.TypeString,
									Value: "ami-12345",
								},
								{
									Type:  hilast.TypeString,
									Value: "ami-54321",
								},
							},
						},
					},
				},
			},
			Error: false,
		},
		{
			Name: "init module with list in map",
			Input: `
variable "amis" {
    default = {
        test1 = ["ami-12345", "ami-54321"]
        test2 = ["ami-123ab", "ami-abc12"]
    }
}

module "ec2_instance" {
    source = "./tf_aws_ec2_instance"
    ami = "${var.amis["test1"]}"
}`,
			Result: hil.EvalConfig{
				GlobalScope: &hilast.BasicScope{
					VarMap: map[string]hilast.Variable{
						"var.ami": {
							Type: hilast.TypeList,
							Value: []hilast.Variable{
								{
									Type:  hilast.TypeString,
									Value: "ami-12345",
								},
								{
									Type:  hilast.TypeString,
									Value: "ami-54321",
								},
							},
						},
					},
				},
			},
			Error: false,
		},
		{
			Name: "init module with map variable",
			Input: `
variable "ami_info" {
    default = {
		name = "awesome image"
		value = "ami-12345"
    }
}

module "ec2_instance" {
    source = "./tf_aws_ec2_instance"
    ami = "${var.ami_info}"
}`,
			Result: hil.EvalConfig{
				GlobalScope: &hilast.BasicScope{
					VarMap: map[string]hilast.Variable{
						"var.ami": {
							Type: hilast.TypeMap,
							Value: map[string]hilast.Variable{
								"name": {
									Type:  hilast.TypeString,
									Value: "awesome image",
								},
								"value": {
									Type:  hilast.TypeString,
									Value: "ami-12345",
								},
							},
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
    ami = "ami-12345"
    num = "1"
}`,
			Result: hil.EvalConfig{},
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
		templates := make(map[string]*hclast.File)
		templates["test.tf"], _ = parser.Parse([]byte(tc.Input))
		schema, _ := schema.Make(files)

		evaluator, _ := NewEvaluator(templates, schema, []*hclast.File{}, config.Init())
		module := schema[0].Modules[0]
		err := evaluator.initModule(module, config.Init())

		if tc.Error && err == nil {
			t.Fatalf("\nshould be happen error.\n\ntestcase: %s", tc.Name)
			continue
		}
		if !tc.Error && err != nil {
			t.Fatalf("\nshould not be happen error.\nError: %s\n\ntestcase: %s", err, tc.Name)
			continue
		}

		if !reflect.DeepEqual(module.EvalConfig, tc.Result) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(module.EvalConfig), pp.Sprint(tc.Result), tc.Name)
		}
	}
}
