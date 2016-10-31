package loader

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/hashicorp/hcl/hcl/ast"
)

func TestLoad(t *testing.T) {
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

		_, err := load(tc.Input)
		if tc.Error == true && err == nil {
			t.Fatalf("should be happen error.\n\ntestcase: %s", tc.Name)
		}
		if tc.Error == false && err != nil {
			t.Fatalf("should not be happen error.\nError: %s\n\ntestcase: %s", err, tc.Name)
		}
	}
}

func TestLoadFile(t *testing.T) {
	type Input struct {
		ListMap map[string]*ast.ObjectList
		File    string
	}

	cases := []struct {
		Name   string
		Input  Input
		Result map[string]*ast.ObjectList
	}{
		{
			Name: "add file list",
			Input: Input{
				ListMap: map[string]*ast.ObjectList{
					"example.tf": &ast.ObjectList{},
				},
				File: "empty.tf",
			},
			Result: map[string]*ast.ObjectList{
				"example.tf": &ast.ObjectList{},
				"empty.tf":   &ast.ObjectList{},
			},
		},
	}

	for _, tc := range cases {
		prev, _ := filepath.Abs(".")
		dir, _ := os.Getwd()
		defer os.Chdir(prev)
		testDir := dir + "/test-fixtures/files"
		os.Chdir(testDir)

		listMap, _ := LoadFile(tc.Input.ListMap, tc.Input.File)
		if !reflect.DeepEqual(listMap, tc.Result) {
			t.Fatalf("Bad: %s\nExpected: %s\n\ntestcase: %s", listMap, tc.Result, tc.Name)
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
		Result map[string]*ast.ObjectList
		Error  bool
	}{
		{
			Name: "load module",
			Input: Input{
				Key: "example",
				Src: "github.com/wata727/example-module",
			},
			Result: map[string]*ast.ObjectList{
				"github.com/wata727/example-module/main.tf":   &ast.ObjectList{},
				"github.com/wata727/example-module/output.tf": &ast.ObjectList{},
			},
			Error: false,
		},
		{
			Name: "module not found",
			Input: Input{
				Key: "not_found",
				Src: "github.com/wata727/example-module",
			},
			Result: nil,
			Error:  true,
		},
	}

	for _, tc := range cases {
		prev, _ := filepath.Abs(".")
		dir, _ := os.Getwd()
		defer os.Chdir(prev)
		testDir := dir + "/test-fixtures/modules"
		os.Chdir(testDir)

		listMap, err := LoadModuleFile(tc.Input.Key, tc.Input.Src)
		if tc.Error == true && err == nil {
			t.Fatalf("should be happen error.\n\ntestcase: %s", tc.Name)
			continue
		}
		if tc.Error == false && err != nil {
			t.Fatalf("should not be happen error.\nError: %s\n\ntestcase: %s", err, tc.Name)
			continue
		}

		if !reflect.DeepEqual(listMap, tc.Result) {
			t.Fatalf("Bad: %s\nExpected: %s\n\ntestcase: %s", listMap, tc.Result, tc.Name)
		}
	}
}

func TestLoadAllFile(t *testing.T) {
	cases := []struct {
		Name   string
		Input  string
		Result map[string]*ast.ObjectList
		Error  bool
	}{
		{
			Name:  "load all files",
			Input: "all-files",
			Result: map[string]*ast.ObjectList{
				"all-files/main.tf":   &ast.ObjectList{},
				"all-files/output.tf": &ast.ObjectList{},
			},
			Error: false,
		},
		{
			Name:   "dir not found",
			Input:  "not_found",
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

		listMap, err := LoadAllFile(tc.Input)
		if tc.Error == true && err == nil {
			t.Fatalf("should be happen error.\n\ntestcase: %s", tc.Name)
			continue
		}
		if tc.Error == false && err != nil {
			t.Fatalf("should not be happen error.\nError: %s\n\ntestcase: %s", err, tc.Name)
			continue
		}

		if !reflect.DeepEqual(listMap, tc.Result) {
			t.Fatalf("Bad: %s\nExpected: %s\n\ntestcase: %s", listMap, tc.Result, tc.Name)
		}
	}
}
