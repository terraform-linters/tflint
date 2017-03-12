package loader

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/token"
	"github.com/k0kubun/pp"
	"github.com/wata727/tflint/logger"
	"github.com/wata727/tflint/state"
)

func TestLoadHCL(t *testing.T) {
	cases := []struct {
		Name  string
		Input string
		Error bool
	}{
		{
			Name:  "return parsed object",
			Input: "template.tf",
			Error: false,
		},
		{
			Name:  "file not found",
			Input: "not_found.tf",
			Error: true,
		},
		{
			Name:  "invalid syntax file",
			Input: "invalid.tf",
			Error: true,
		},
	}

	for _, tc := range cases {
		prev, _ := filepath.Abs(".")
		dir, _ := os.Getwd()
		defer os.Chdir(prev)
		testDir := dir + "/test-fixtures/files"
		os.Chdir(testDir)

		_, err := loadHCL(tc.Input, logger.Init(false))
		if tc.Error && err == nil {
			t.Fatalf("\nshould be happen error.\n\ntestcase: %s", tc.Name)
		}
		if !tc.Error && err != nil {
			t.Fatalf("\nshould not be happen error.\nError: %s\n\ntestcase: %s", err, tc.Name)
		}
	}
}

func TestLoadJSON(t *testing.T) {
	cases := []struct {
		Name  string
		Input string
		Error bool
	}{
		{
			Name:  "return parsed object",
			Input: "template.json",
			Error: false,
		},
		{
			Name:  "file not found",
			Input: "not_found.json",
			Error: true,
		},
		{
			Name:  "invalid syntax file",
			Input: "invalid.json",
			Error: true,
		},
	}

	for _, tc := range cases {
		prev, _ := filepath.Abs(".")
		dir, _ := os.Getwd()
		defer os.Chdir(prev)
		testDir := dir + "/test-fixtures/files"
		os.Chdir(testDir)

		_, err := loadJSON(tc.Input, logger.Init(false))
		if tc.Error && err == nil {
			t.Fatalf("\nshould be happen error.\n\ntestcase: %s", tc.Name)
		}
		if !tc.Error && err != nil {
			t.Fatalf("\nshould not be happen error.\nError: %s\n\ntestcase: %s", err, tc.Name)
		}
	}
}

func TestLoadTemplate(t *testing.T) {
	type Input struct {
		Templates map[string]*ast.File
		File      string
	}

	cases := []struct {
		Name   string
		Input  Input
		Result map[string]*ast.File
	}{
		{
			Name: "add file list",
			Input: Input{
				Templates: map[string]*ast.File{
					"example.tf": {},
				},
				File: "empty.tf",
			},
			Result: map[string]*ast.File{
				"example.tf": {},
				"empty.tf":   {Node: &ast.ObjectList{}},
			},
		},
	}

	for _, tc := range cases {
		prev, _ := filepath.Abs(".")
		dir, _ := os.Getwd()
		defer os.Chdir(prev)
		testDir := dir + "/test-fixtures/files"
		os.Chdir(testDir)
		load := &Loader{
			Logger:    logger.Init(false),
			Templates: tc.Input.Templates,
		}

		load.LoadTemplate(tc.Input.File)
		if !reflect.DeepEqual(load.Templates, tc.Result) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(load.Templates), pp.Sprint(tc.Result), tc.Name)
		}
	}
}

func TestLoadModuleFile(t *testing.T) {
	type Input struct {
		Key string
		Src string
	}

	cases := []struct {
		Name   string
		Input  Input
		Result map[string]*ast.File
		Error  bool
	}{
		{
			Name: "load module",
			Input: Input{
				Key: "example",
				Src: "github.com/wata727/example-module",
			},
			Result: map[string]*ast.File{
				"github.com/wata727/example-module/main.tf":   {Node: &ast.ObjectList{}},
				"github.com/wata727/example-module/output.tf": {Node: &ast.ObjectList{}},
			},
			Error: false,
		},
		{
			Name: "module not found",
			Input: Input{
				Key: "not_found",
				Src: "github.com/wata727/example-module",
			},
			Result: make(map[string]*ast.File),
			Error:  true,
		},
	}

	for _, tc := range cases {
		prev, _ := filepath.Abs(".")
		dir, _ := os.Getwd()
		defer os.Chdir(prev)
		testDir := dir + "/test-fixtures/modules"
		os.Chdir(testDir)
		load := NewLoader(false)

		err := load.LoadModuleFile(tc.Input.Key, tc.Input.Src)
		if tc.Error && err == nil {
			t.Fatalf("\nshould be happen error.\n\ntestcase: %s", tc.Name)
			continue
		}
		if !tc.Error && err != nil {
			t.Fatalf("\nshould not be happen error.\nError: %s\n\ntestcase: %s", err, tc.Name)
			continue
		}

		if !reflect.DeepEqual(load.Templates, tc.Result) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(load.Templates), pp.Sprint(tc.Result), tc.Name)
		}
	}
}

func TestLoadAllTemplate(t *testing.T) {
	cases := []struct {
		Name   string
		Input  string
		Result map[string]*ast.File
		Error  bool
	}{
		{
			Name:  "load all files",
			Input: "all-files",
			Result: map[string]*ast.File{
				"all-files/main.tf":   {Node: &ast.ObjectList{}},
				"all-files/output.tf": {Node: &ast.ObjectList{}},
			},
			Error: false,
		},
		{
			Name:   "dir not found",
			Input:  "not_found",
			Result: make(map[string]*ast.File),
			Error:  true,
		},
	}

	for _, tc := range cases {
		prev, _ := filepath.Abs(".")
		dir, _ := os.Getwd()
		defer os.Chdir(prev)
		testDir := dir + "/test-fixtures"
		os.Chdir(testDir)
		load := NewLoader(false)

		err := load.LoadAllTemplate(tc.Input)
		if tc.Error && err == nil {
			t.Fatalf("\nshould be happen error.\n\ntestcase: %s", tc.Name)
			continue
		}
		if !tc.Error && err != nil {
			t.Fatalf("\nshould not be happen error.\nError: %s\n\ntestcase: %s", err, tc.Name)
			continue
		}

		if !reflect.DeepEqual(load.Templates, tc.Result) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(load.Templates), pp.Sprint(tc.Result), tc.Name)
		}
	}
}

func TestLoadState(t *testing.T) {
	cases := []struct {
		Name   string
		Dir    string
		Result *state.TFState
	}{
		{
			Name: "load local state",
			Dir:  "local-state",
			Result: &state.TFState{
				Modules: []*state.Module{
					{
						Resources: map[string]*state.Resource{
							"aws_db_parameter_group.production": {
								Type:         "aws_db_parameter_group",
								Dependencies: []string{},
								Primary: &state.Instance{
									ID: "production",
									Attributes: map[string]string{
										"arn":         "arn:aws:rds:us-east-1:hogehoge:pg:production",
										"description": "production-db-parameter-group",
										"family":      "mysql5.6",
										"id":          "production",
										"name":        "production",
										"parameter.#": "0",
										"tags.%":      "0",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Name: "load remote state",
			Dir:  "remote-state",
			Result: &state.TFState{
				Modules: []*state.Module{
					{
						Resources: map[string]*state.Resource{
							"aws_db_parameter_group.staging": {
								Type:         "aws_db_parameter_group",
								Dependencies: []string{},
								Primary: &state.Instance{
									ID: "staging",
									Attributes: map[string]string{
										"arn":         "arn:aws:rds:us-east-1:hogehoge:pg:staging",
										"description": "staging-db-parameter-group",
										"family":      "mysql5.6",
										"id":          "staging",
										"name":        "staging",
										"parameter.#": "0",
										"tags.%":      "0",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			Name:   "state not found",
			Dir:    "files",
			Result: &state.TFState{},
		},
	}

	for _, tc := range cases {
		prev, _ := filepath.Abs(".")
		dir, _ := os.Getwd()
		defer os.Chdir(prev)
		testDir := dir + "/test-fixtures/" + tc.Dir
		os.Chdir(testDir)
		load := NewLoader(false)

		load.LoadState()
		if !reflect.DeepEqual(load.State, tc.Result) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(load.State), pp.Sprint(tc.Result), tc.Name)
		}
		os.Chdir(prev)
	}
}

func TestLoadTFVars(t *testing.T) {
	cases := []struct {
		Name   string
		Input  []string
		Result []*ast.File
	}{
		{
			Name:  "Load multi tfvars",
			Input: []string{"terraform.tfvars", "tfvars.json"},
			Result: []*ast.File{
				{
					Node: &ast.ObjectList{
						Items: []*ast.ObjectItem{
							{
								Keys: []*ast.ObjectKey{
									{
										Token: token.Token{
											Type: 4,
											Pos: token.Pos{
												Line:   1,
												Column: 1,
											},
											Text: "type",
											JSON: false,
										},
									},
								},
								Assign: token.Pos{
									Offset: 5,
									Line:   1,
									Column: 6,
								},
								Val: &ast.LiteralType{
									Token: token.Token{
										Type: 9,
										Pos: token.Pos{
											Offset: 7,
											Line:   1,
											Column: 8,
										},
										Text: "\"t2.micro\"",
										JSON: false,
									},
								},
							},
						},
					},
				},
				{
					Node: &ast.ObjectList{
						Items: []*ast.ObjectItem{
							{
								Keys: []*ast.ObjectKey{
									{
										Token: token.Token{
											Type: 9,
											Pos:  token.Pos{},
											Text: "\"name\"",
											JSON: true,
										},
									},
								},
								Assign: token.Pos{
									Offset: 10,
									Line:   2,
									Column: 9,
								},
								Val: &ast.LiteralType{
									Token: token.Token{
										Type: 9,
										Pos:  token.Pos{},
										Text: "\"awesome-app\"",
										JSON: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		prev, _ := filepath.Abs(".")
		dir, _ := os.Getwd()
		defer os.Chdir(prev)
		testDir := dir + "/test-fixtures/tfvars"
		os.Chdir(testDir)
		load := &Loader{
			Logger: logger.Init(false),
		}

		load.LoadTFVars(tc.Input)
		if !reflect.DeepEqual(load.TFVars, tc.Result) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(load.TFVars), pp.Sprint(tc.Result), tc.Name)
		}
	}
}

func TestDump(t *testing.T) {
	load := NewLoader(false)
	templates := map[string]*ast.File{
		"main.tf":   {},
		"output.tf": {},
	}
	state := &state.TFState{
		Modules: []*state.Module{
			{
				Resources: map[string]*state.Resource{
					"aws_db_parameter_group.production": {
						Type:         "aws_db_parameter_group",
						Dependencies: []string{},
						Primary: &state.Instance{
							ID: "production",
							Attributes: map[string]string{
								"arn":         "arn:aws:rds:us-east-1:hogehoge:pg:production",
								"description": "production-db-parameter-group",
								"family":      "mysql5.6",
								"id":          "production",
								"name":        "production",
								"parameter.#": "0",
								"tags.%":      "0",
							},
						},
					},
				},
			},
		},
	}
	tfvars := []*ast.File{
		{
			Node: &ast.ObjectList{},
		},
	}
	load.Templates = templates
	load.State = state
	load.TFVars = tfvars

	dumpTemplates, dumpState, dumpTFvars := load.Dump()
	if !reflect.DeepEqual(dumpTemplates, templates) {
		t.Fatalf("\nBad: %s\nExpected: %s\n\n", pp.Sprint(dumpTemplates), pp.Sprint(templates))
	}
	if !reflect.DeepEqual(dumpState, state) {
		t.Fatalf("\nBad: %s\nExpected: %s\n\n", pp.Sprint(dumpState), pp.Sprint(state))
	}
	if !reflect.DeepEqual(dumpTFvars, tfvars) {
		t.Fatalf("\nBad: %s\nExpected: %s\n\n", pp.Sprint(dumpTFvars), pp.Sprint(tfvars))
	}
}
