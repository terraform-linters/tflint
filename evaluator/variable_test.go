package evaluator

import (
	"reflect"
	"testing"

	hclast "github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/parser"
	hilast "github.com/hashicorp/hil/ast"
	"github.com/k0kubun/pp"
)

func TestDetectVariables(t *testing.T) {
	cases := []struct {
		Name    string
		Input   map[string]string
		Varfile map[string]string
		Result  map[string]hilast.Variable
		Error   bool
	}{
		{
			Name: "return hil variable from hcl object list",
			Input: map[string]string{
				"variable.tf": `
variable "type" {
	default = "name"
}`,
			},
			Varfile: map[string]string{},
			Result: map[string]hilast.Variable{
				"var.type": {
					Type:  hilast.TypeString,
					Value: "name",
				},
			},
			Error: false,
		},
		{
			Name: "return empty from hcl object list when default is not found",
			Input: map[string]string{
				"variable.tf": `variable "type" {}`,
			},
			Varfile: map[string]string{},
			Result:  map[string]hilast.Variable{},
			Error:   false,
		},
		{
			Name: "return empty from hcl object list when variable not found",
			Input: map[string]string{
				"variable.tf": `
provider "aws" {
	region = "us-east-1"
}`,
			},
			Varfile: map[string]string{},
			Result:  map[string]hilast.Variable{},
			Error:   false,
		},
		{
			Name: "return hil variables from multi hcl object list",
			Input: map[string]string{
				"variable1.tf": `
variable "type" {
	default = "name"
}`,
				"variable2.tf": `
variable "stat" {
	default = "usage"
}`,
			},
			Varfile: map[string]string{},
			Result: map[string]hilast.Variable{
				"var.type": {
					Type:  hilast.TypeString,
					Value: "name",
				},
				"var.stat": {
					Type:  hilast.TypeString,
					Value: "usage",
				},
			},
			Error: false,
		},
		{
			Name: "return hil variables from multi hcl object lists when use multi tfvars",
			Input: map[string]string{
				"variable1.tf": `variable "type" {}`,
				"variable2.tf": `variable "mode" {}`,
			},
			Varfile: map[string]string{
				"terraform.tfvars": `type = "t2.micro"`,
				"example.tfvars":   `mode = "complex"`,
			},
			Result: map[string]hilast.Variable{
				"var.type": {
					Type:  hilast.TypeString,
					Value: "t2.micro",
				},
				"var.mode": {
					Type:  hilast.TypeString,
					Value: "complex",
				},
			},
			Error: false,
		},
		{
			Name: "return empty when set in tfvars but not set in tf",
			Input: map[string]string{
				"variable.tf": `variable "type" {}`,
			},
			Varfile: map[string]string{
				"terraform.tfvars": `new_type = "name"`,
			},
			Result: map[string]hilast.Variable{},
			Error:  false,
		},
	}

	for _, tc := range cases {
		templates := make(map[string]*hclast.File)
		varfile := []*hclast.File{}
		for k, v := range tc.Input {
			templates[k], _ = parser.Parse([]byte(v))
		}
		for _, v := range tc.Varfile {
			root, _ := parser.Parse([]byte(v))
			varfile = append(varfile, root)
		}

		result, err := detectVariables(templates, varfile)
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

func TestDecodeTFVars(t *testing.T) {
	cases := []struct {
		Name   string
		Input  []string
		Result []map[string]interface{}
		Error  bool
	}{
		{
			Name: "decode multi tfvars",
			Input: []string{
				`type = "t2.micro"`,
				`name = "test"`,
			},
			Result: []map[string]interface{}{
				{
					"type": "t2.micro",
				},
				{
					"name": "test",
				},
			},
			Error: false,
		},
		{
			Name: "decode complex tfvars",
			Input: []string{`
types = ["t2.nano", "t2.micro"]
complex = {
  foo = "bar"
  baz = {
    nest = true
    list = ["attr"]
  }
}
`,
			},
			Result: []map[string]interface{}{
				{
					"types": []interface{}{"t2.nano", "t2.micro"},
					"complex": []map[string]interface{}{
						{
							"foo": "bar",
							"baz": []map[string]interface{}{
								{
									"nest": true,
									"list": []interface{}{"attr"},
								},
							},
						},
					},
				},
			},
			Error: false,
		},
	}

	for _, tc := range cases {
		varfile := []*hclast.File{}
		for _, v := range tc.Input {
			root, _ := parser.Parse([]byte(v))
			varfile = append(varfile, root)
		}

		result, err := decodeTFVars(varfile)
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

func TestOverriddenVariable(t *testing.T) {
	cases := []struct {
		Name             string
		Input            *hclVariable
		DecodedVariables []map[string]interface{}
		Result           interface{}
	}{
		{
			Name: "overridden variable when not set default",
			Input: &hclVariable{
				Name:    "type",
				Default: nil,
			},
			DecodedVariables: []map[string]interface{}{
				{
					"type": "t2.micro",
				},
			},
			Result: "t2.micro",
		},
		{
			Name: "overridden variable when already set default",
			Input: &hclVariable{
				Name:    "type",
				Default: "t2.micro",
			},
			DecodedVariables: []map[string]interface{}{
				{
					"type": "t2.nano",
				},
			},
			Result: "t2.nano",
		},
		{
			Name: "overridden variable by last specfied varfile when conflict variables",
			Input: &hclVariable{
				Name:    "type",
				Default: nil,
			},
			DecodedVariables: []map[string]interface{}{
				{
					"type": "t2.nano",
				},
				{
					"type": "m4.large",
				},
			},
			Result: "m4.large",
		},
		{
			Name: "merge and overridden complex variable when conflict variables",
			Input: &hclVariable{
				Name:    "complex",
				Default: nil,
			},
			DecodedVariables: []map[string]interface{}{
				{
					"complex": []map[string]interface{}{
						{
							"foo": "bar",
							"baz": []map[string]interface{}{
								{
									"nest": true,
									"list": []interface{}{"attr"},
								},
							},
						},
					},
				},
				{
					"complex": []map[string]interface{}{
						{
							"foo":  "baz",
							"nest": true,
							"baz": []map[string]interface{}{
								{
									"list":  []interface{}{"value"},
									"merge": true,
								},
							},
						},
					},
				},
			},
			Result: []map[string]interface{}{
				{
					"foo":  "baz",
					"nest": true,
					"baz": []map[string]interface{}{
						{
							"list":  []interface{}{"value"},
							"merge": true,
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		overriddenVariable(tc.Input, tc.DecodedVariables)
		if !reflect.DeepEqual(tc.Input.Default, tc.Result) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(tc.Input.Default), pp.Sprint(tc.Result), tc.Name)
		}
	}
}

func TestParseVariable(t *testing.T) {
	type Input struct {
		Val  interface{}
		Type string
	}

	cases := []struct {
		Name   string
		Input  Input
		Result hilast.Variable
	}{
		{
			Name: "parse string with correct type",
			Input: Input{
				Val:  "test",
				Type: "string",
			},
			Result: hilast.Variable{
				Type:  hilast.TypeString,
				Value: "test",
			},
		},
		{
			Name: "parse list with correct type",
			Input: Input{
				Val:  []string{"test1", "test2"},
				Type: "list",
			},
			Result: hilast.Variable{
				Type: hilast.TypeList,
				Value: []hilast.Variable{
					{
						Type:  hilast.TypeString,
						Value: "test1",
					},
					{
						Type:  hilast.TypeString,
						Value: "test2",
					},
				},
			},
		},
		{
			Name: "parse map with correct type",
			Input: Input{
				// HCL map variable is map in slice in Go.
				Val:  []map[string]string{{"test1": "test2", "test3": "test4"}},
				Type: "map",
			},
			Result: hilast.Variable{
				Type: hilast.TypeMap,
				Value: map[string]hilast.Variable{
					"test1": {
						Type:  hilast.TypeString,
						Value: "test2",
					},
					"test3": {
						Type:  hilast.TypeString,
						Value: "test4",
					},
				},
			},
		},
		{
			Name: "parse string with unknown type",
			Input: Input{
				Val:  "test",
				Type: "",
			},
			Result: hilast.Variable{
				Type:  hilast.TypeString,
				Value: "test",
			},
		},
		{
			Name: "parse list with unknown type",
			Input: Input{
				Val:  []string{"test1", "test2"},
				Type: "",
			},
			Result: hilast.Variable{
				Type: hilast.TypeList,
				Value: []hilast.Variable{
					{
						Type:  hilast.TypeString,
						Value: "test1",
					},
					{
						Type:  hilast.TypeString,
						Value: "test2",
					},
				},
			},
		},
		{
			Name: "parse map with unknown type",
			Input: Input{
				Val:  []map[string]string{{"test1": "test2", "test3": "test4"}},
				Type: "",
			},
			Result: hilast.Variable{
				Type: hilast.TypeMap,
				Value: map[string]hilast.Variable{
					"test1": {
						Type:  hilast.TypeString,
						Value: "test2",
					},
					"test3": {
						Type:  hilast.TypeString,
						Value: "test4",
					},
				},
			},
		},
		{
			Name: "parse string with incorrect type",
			Input: Input{
				Val:  "test",
				Type: "map",
			},
			Result: hilast.Variable{
				Type:  hilast.TypeString,
				Value: "test",
			},
		},
		{
			Name: "parse list with incorrect type",
			Input: Input{
				Val:  []string{"test1", "test2"},
				Type: "string",
			},
			Result: hilast.Variable{
				Type: hilast.TypeList,
				Value: []hilast.Variable{
					{
						Type:  hilast.TypeString,
						Value: "test1",
					},
					{
						Type:  hilast.TypeString,
						Value: "test2",
					},
				},
			},
		},
		{
			Name: "parse map with incorrect type",
			Input: Input{
				Val:  []map[string]string{{"test1": "test2", "test3": "test4"}},
				Type: "slice",
			},
			Result: hilast.Variable{
				Type: hilast.TypeMap,
				Value: map[string]hilast.Variable{
					"test1": {
						Type:  hilast.TypeString,
						Value: "test2",
					},
					"test3": {
						Type:  hilast.TypeString,
						Value: "test4",
					},
				},
			},
		},
		{
			Name: "parse complex struct var",
			// ```HCL
			// value = {
			//     a = ["test1", "test2"]
			//     b = {
			//         test3 = 1
			//         test4 = 10
			//     }
			// }
			// ```
			Input: Input{
				Val: []map[string]interface{}{
					{
						"a": []string{"test1", "test2"},
						"b": []map[string]int{
							{
								"test3": 1,
								"test4": 10,
							},
						},
					},
				},
			},
			Result: hilast.Variable{
				Type: hilast.TypeMap,
				Value: map[string]hilast.Variable{
					"a": {
						Type: hilast.TypeList,
						Value: []hilast.Variable{
							{
								Type:  hilast.TypeString,
								Value: "test1",
							},
							{
								Type:  hilast.TypeString,
								Value: "test2",
							},
						},
					},
					"b": {
						Type: hilast.TypeMap,
						Value: map[string]hilast.Variable{
							"test3": {
								Type:  hilast.TypeString,
								Value: 1,
							},
							"test4": {
								Type:  hilast.TypeString,
								Value: 10,
							},
						},
					},
				},
			},
		},
		{
			Name: "parse empty list",
			Input: Input{
				Val:  []string{},
				Type: "list",
			},
			Result: hilast.Variable{
				Type:  hilast.TypeList,
				Value: []hilast.Variable{},
			},
		},
	}

	for _, tc := range cases {
		result := parseVariable(tc.Input.Val, tc.Input.Type)
		if !reflect.DeepEqual(result, tc.Result) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(result), pp.Sprint(tc.Result), tc.Name)
		}
	}
}
