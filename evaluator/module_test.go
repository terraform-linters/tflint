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
				"960d94c2f60d34845dc3051edfad76e1": &hclModule{
					Name:   "ec2_instance",
					Source: "./tf_aws_ec2_instance",
					Config: hil.EvalConfig{
						GlobalScope: &hilast.BasicScope{
							VarMap: map[string]hilast.Variable{
								"var.ami": hilast.Variable{
									Type:  hilast.TypeString,
									Value: "ami-12345",
								},
								"var.num": hilast.Variable{
									Type:  hilast.TypeString,
									Value: "1",
								},
							},
						},
					},
					ListMap: map[string]*hclast.ObjectList{},
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
				"960d94c2f60d34845dc3051edfad76e1": &hclModule{
					Name:   "ec2_instance",
					Source: "./tf_aws_ec2_instance",
					Config: hil.EvalConfig{
						GlobalScope: &hilast.BasicScope{
							VarMap: map[string]hilast.Variable{
								"var.ami": hilast.Variable{
									Type:  hilast.TypeString,
									Value: "ami-12345",
								},
								"var.num": hilast.Variable{
									Type:  hilast.TypeString,
									Value: "1",
								},
							},
						},
					},
					ListMap: map[string]*hclast.ObjectList{},
				},
				"0cf2d4dab02de8de33c7058799b6f81e": &hclModule{
					Name:   "ec2_instance",
					Source: "github.com/wata727/example-module",
					Config: hil.EvalConfig{
						GlobalScope: &hilast.BasicScope{
							VarMap: map[string]hilast.Variable{
								"var.ami": hilast.Variable{
									Type:  hilast.TypeString,
									Value: "ami-54321",
								},
							},
						},
					},
					ListMap: map[string]*hclast.ObjectList{},
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

		listMap := make(map[string]*hclast.ObjectList)
		for k, v := range tc.Input {
			root, _ := parser.Parse([]byte(v))
			list, _ := root.Node.(*hclast.ObjectList)
			listMap[k] = list
		}
		result, err := detectModules(listMap, config.Init())
		if tc.Error && err == nil {
			t.Fatalf("\nshould be happen error.\n\ntestcase: %s", tc.Name)
			continue
		}
		if !tc.Error && err != nil {
			t.Fatalf("\nshould not be happen error.\nError: %s\n\ntestcase: %s", err, tc.Name)
			continue
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
