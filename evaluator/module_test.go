package evaluator

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	hclast "github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/parser"
	"github.com/hashicorp/hil"
	hilast "github.com/hashicorp/hil/ast"
	"github.com/k0kubun/pp"
	"github.com/wata727/tflint/config"
)

func TestDetectModules(t *testing.T) {
	cases := []struct {
		Name   string
		Input  map[string]string
		Result map[string]*hclModule
		Error  bool
	}{
		{
			Name: "detect module",
			Input: map[string]string{
				"module.tf": `
module "ec2_instance" {
    source = "./tf_aws_ec2_instance"
    ami = "ami-12345"
    num = "1"
}`,
			},
			Result: map[string]*hclModule{
				"960d94c2f60d34845dc3051edfad76e1": {
					Name:   "ec2_instance",
					Source: "./tf_aws_ec2_instance",
					File:   "module.tf",
					Config: hil.EvalConfig{
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
					Templates: map[string]*hclast.File{},
				},
			},
			Error: false,
		},
		{
			Name: "detect multi modules",
			Input: map[string]string{
				"module1.tf": `
module "ec2_instance" {
    source = "./tf_aws_ec2_instance"
    ami = "ami-12345"
    num = "1"
}`,
				"module2.tf": `
module "ec2_instance" {
    source = "github.com/wata727/example-module"
    ami = "ami-54321"
}`,
			},
			Result: map[string]*hclModule{
				"960d94c2f60d34845dc3051edfad76e1": {
					Name:   "ec2_instance",
					Source: "./tf_aws_ec2_instance",
					File:   "module1.tf",
					Config: hil.EvalConfig{
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
					Templates: map[string]*hclast.File{},
				},
				"0cf2d4dab02de8de33c7058799b6f81e": {
					Name:   "ec2_instance",
					Source: "github.com/wata727/example-module",
					File:   "module2.tf",
					Config: hil.EvalConfig{
						GlobalScope: &hilast.BasicScope{
							VarMap: map[string]hilast.Variable{
								"var.ami": {
									Type:  hilast.TypeString,
									Value: "ami-54321",
								},
							},
						},
					},
					Templates: map[string]*hclast.File{},
				},
			},
			Error: false,
		},
		{
			Name: "detect module with string variable",
			Input: map[string]string{
				"module.tf": `
variable "ami" {
    default = "ami-12345"
}

module "ec2_instance" {
    source = "./tf_aws_ec2_instance"
    ami = "${var.ami}"
}`,
			},
			Result: map[string]*hclModule{
				"960d94c2f60d34845dc3051edfad76e1": {
					Name:   "ec2_instance",
					Source: "./tf_aws_ec2_instance",
					File:   "module.tf",
					Config: hil.EvalConfig{
						GlobalScope: &hilast.BasicScope{
							VarMap: map[string]hilast.Variable{
								"var.ami": {
									Type:  hilast.TypeString,
									Value: "ami-12345",
								},
							},
						},
					},
					Templates: map[string]*hclast.File{},
				},
			},
			Error: false,
		},
		{
			Name: "detect module with list variable",
			Input: map[string]string{
				"module.tf": `
variable "amis" {
    default = ["ami-12345", "ami-54321"]
}

module "ec2_instance" {
    source = "./tf_aws_ec2_instance"
    ami = "${var.amis}"
}`,
			},
			Result: map[string]*hclModule{
				"960d94c2f60d34845dc3051edfad76e1": {
					Name:   "ec2_instance",
					Source: "./tf_aws_ec2_instance",
					File:   "module.tf",
					Config: hil.EvalConfig{
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
					Templates: map[string]*hclast.File{},
				},
			},
			Error: false,
		},
		{
			Name: "detect module with map variable",
			Input: map[string]string{
				"module.tf": `
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
			},
			Result: map[string]*hclModule{
				"960d94c2f60d34845dc3051edfad76e1": {
					Name:   "ec2_instance",
					Source: "./tf_aws_ec2_instance",
					File:   "module.tf",
					Config: hil.EvalConfig{
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
					Templates: map[string]*hclast.File{},
				},
			},
			Error: false,
		},
		{
			Name: "invalid source",
			Input: map[string]string{
				"module.tf": `
module "ec2_instances" {
    source = ["./tf_aws_ec2_instance", "github.com/wata727/example-module"]
    ami = "ami-12345"
    num = "1"
}`,
			},
			Result: nil,
			Error:  true,
		},
		{
			Name: "module not found",
			Input: map[string]string{
				"module.tf": `
module "ec2_instances" {
    source = "unresolvable_module_source"
    ami = "ami-12345"
    num = "1"
}`,
			},
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
		templates := make(map[string]*hclast.File)
		for k, v := range tc.Input {
			templates[k], _ = parser.Parse([]byte(v))
		}
		evaluator, _ := NewEvaluator(templates, []*hclast.File{}, config.Init())
		result, err := evaluator.detectModules(templates, config.Init())
		if tc.Error && err == nil {
			t.Fatalf("\nshould be happen error.\n\ntestcase: %s", tc.Name)
			continue
		}
		if !tc.Error && err != nil {
			t.Fatalf("\nshould not be happen error.\nError: %s\n\ntestcase: %s", err, tc.Name)
			continue
		}
		// We don't care how the ObjectItem was created
		for _, module := range result {
			module.ObjectItem = nil
		}

		if !reflect.DeepEqual(result, tc.Result) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(result), pp.Sprint(tc.Result), tc.Name)
		}
	}
}

func TestModuleKey(t *testing.T) {
	type Input struct {
		Name   string
		Source string
	}

	cases := []struct {
		Name   string
		Input  Input
		Result string
	}{
		{
			Name: "return module hash",
			Input: Input{
				Name:   "ec2_instance",
				Source: "./tf_aws_ec2_instance",
			},
			Result: "960d94c2f60d34845dc3051edfad76e1",
		},
	}

	for _, tc := range cases {
		result := moduleKey(tc.Input.Name, tc.Input.Source)
		if result != tc.Result {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", result, tc.Result, tc.Name)
		}
	}
}
