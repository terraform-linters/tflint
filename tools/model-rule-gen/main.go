package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"

	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hclparse"
)

type mappingFile struct {
	Import   string    `hcl:"import"`
	Mappings []mapping `hcl:"mapping,block"`
	Tests    []test    `hcl:"test,block"`
}

type mapping struct {
	Resource string         `hcl:"resource,label"`
	Attrs    hcl.Attributes `hcl:",remain"`
}

type test struct {
	Resource  string `hcl:"resource,label"`
	Attribute string `hcl:"attribute,label"`
	OK        string `hcl:"ok"`
	NG        string `hcl:"ng"`
}

func main() {
	files, err := filepath.Glob("../rules/awsrules/models/mappings/*.hcl")
	if err != nil {
		panic(err)
	}

	mappingFiles := []mappingFile{}
	for _, file := range files {
		parser := hclparse.NewParser()
		f, diags := parser.ParseHCLFile(file)
		if diags.HasErrors() {
			panic(diags)
		}

		var mf mappingFile
		diags = gohcl.DecodeBody(f.Body, nil, &mf)
		if diags.HasErrors() {
			panic(diags)
		}
		mappingFiles = append(mappingFiles, mf)
	}

	generatedRules := []string{}
	for _, mappingFile := range mappingFiles {
		raw, err := ioutil.ReadFile(fmt.Sprintf("../rules/awsrules/models/%s", mappingFile.Import))
		if err != nil {
			panic(err)
		}

		var api map[string]interface{}
		err = json.Unmarshal(raw, &api)
		if err != nil {
			panic(err)
		}
		shapes := api["shapes"].(map[string]interface{})

		for _, mapping := range mappingFile.Mappings {
			for attribute, value := range mapping.Attrs {
				shapeName := value.Expr.Variables()[0].RootName()
				if shapeName == "any" {
					continue
				}
				model := shapes[shapeName].(map[string]interface{})
				if validMapping(model) {
					generateRuleFile(mapping.Resource, attribute, model)
					for _, test := range mappingFile.Tests {
						if mapping.Resource == test.Resource && attribute == test.Attribute {
							generateRuleTestFile(mapping.Resource, attribute, model, test)
						}
					}
					generatedRules = append(generatedRules, fmt.Sprintf("%s_invalid_%s", mapping.Resource, attribute))
				}
			}
		}
	}

	sort.Strings(generatedRules)
	generateProviderFile(generatedRules)
}

func validMapping(model map[string]interface{}) bool {
	switch model["type"].(string) {
	case "string":
		if _, ok := model["max"]; ok {
			return true
		}
		if min, ok := model["min"]; ok && int(min.(float64)) > 2 {
			return true
		}
		if _, ok := model["pattern"]; ok {
			return true
		}
		if _, ok := model["enum"]; ok {
			return true
		}
		return false
	default:
		// Unsupported types
		return false
	}
}
