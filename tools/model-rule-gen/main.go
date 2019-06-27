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
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws"
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

	awsProvider := aws.Provider().(*schema.Provider)

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
				checkAttributeType(mapping.Resource, attribute, model, awsProvider)
				if validMapping(model) {
					generateRuleFile(mapping.Resource, attribute, model)
					for _, test := range mappingFile.Tests {
						if mapping.Resource == test.Resource && attribute == test.Attribute {
							generateRuleTestFile(mapping.Resource, attribute, model, test)
						}
					}
					generatedRules = append(generatedRules, makeRuleName(mapping.Resource, attribute))
				}
			}
		}
	}

	sort.Strings(generatedRules)
	generateProviderFile(generatedRules)
}

func checkAttributeType(resource, attribute string, model map[string]interface{}, provider *schema.Provider) {
	resourceSchema, ok := provider.ResourcesMap[resource]
	if !ok {
		panic(fmt.Sprintf("resource `%s` not found in the Terraform schema", resource))
	}
	attrSchema, ok := resourceSchema.Schema[attribute]
	if !ok {
		panic(fmt.Sprintf("`%s.%s` not found in the Terraform schema", resource, attribute))
	}

	switch model["type"].(string) {
	case "string":
		if attrSchema.Type != schema.TypeString {
			panic(fmt.Sprintf("`%s.%s` is expected as string, but not", resource, attribute))
		}
	default:
		// noop
	}
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
