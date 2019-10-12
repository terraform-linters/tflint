package main

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws"
	"github.com/wata727/tflint/tools/utils"
)

type definition struct {
	Rules []rule `hcl:"rule,block"`
}

type rule struct {
	Name         string `hcl:"name,label"`
	Resource     string `hcl:"resource"`
	Attribute    string `hcl:"attribute"`
	SourceAction string `hcl:"source_action"`
	Template     string `hcl:"template"`
}

type ruleMeta struct {
	RuleName      string
	RuleNameCC    string
	ResourceType  string
	AttributeName string
	DataType      string
	ActionName    string
	Template      string
}

type providerMeta struct {
	RuleNameCCList []string
}

var awsProvider = aws.Provider().(*schema.Provider)

func main() {
	files, err := filepath.Glob("../rules/awsrules/api/definitions/*.hcl")
	if err != nil {
		panic(err)
	}

	providerMeta := &providerMeta{}
	for _, file := range files {
		parser := hclparse.NewParser()
		f, diags := parser.ParseHCLFile(file)
		if diags.HasErrors() {
			panic(diags)
		}

		var def definition
		diags = gohcl.DecodeBody(f.Body, nil, &def)
		if diags.HasErrors() {
			panic(diags)
		}

		for _, rule := range def.Rules {
			meta := &ruleMeta{
				RuleName:      rule.Name,
				RuleNameCC:    utils.ToCamel(rule.Name),
				ResourceType:  rule.Resource,
				AttributeName: rule.Attribute,
				DataType:      dataType(rule.Resource, rule.Attribute),
				ActionName:    rule.SourceAction,
				Template:      rule.Template,
			}

			utils.GenerateFile(
				fmt.Sprintf("../rules/awsrules/api/%s.go", rule.Name),
				"../rules/awsrules/api/rule.go.tmpl",
				meta,
			)

			providerMeta.RuleNameCCList = append(providerMeta.RuleNameCCList, meta.RuleNameCC)
		}
	}

	sort.Strings(providerMeta.RuleNameCCList)
	utils.GenerateFile(
		"../rules/provider_api.go",
		"../rules/provider_api.go.tmpl",
		providerMeta,
	)
}

func dataType(resource, attribute string) string {
	resourceSchema, ok := awsProvider.ResourcesMap[resource]
	if !ok {
		panic(fmt.Sprintf("resource `%s` not found in the Terraform schema", resource))
	}
	attrSchema, ok := resourceSchema.Schema[attribute]
	if !ok {
		panic(fmt.Sprintf("`%s.%s` not found in the Terraform schema", resource, attribute))
	}

	switch attrSchema.Type {
	case schema.TypeString:
		return "string"
	case schema.TypeSet:
		return "list"
	default:
		panic(fmt.Errorf("Unexpected data type: %#v", attrSchema.Type))
	}
}
