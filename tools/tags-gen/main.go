package main

import (
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws"

	"github.com/terraform-linters/tflint/tools/utils"
)

type tagRuleMeta struct {
	RuleName     string
	ResourceName string
	ResourceType string
}

type providerMeta struct {
	ResourceNames []string
}

var awsProvider = aws.Provider().(*schema.Provider)

func main() {
	providerMeta := &providerMeta{}

	for k, v := range awsProvider.ResourcesMap {
		if _, ok := v.Schema["tags"]; ok {
			ruleName := "aws_resource_tags_" + k
			meta := &tagRuleMeta{
				RuleName:     ruleName,
				ResourceName: utils.ToCamel(k),
				ResourceType: k,
			}
			utils.GenerateFile(
				fmt.Sprintf("../rules/awsrules/tags/%s.go", ruleName),
				"../rules/awsrules/tags/aws_resource_tags.go.tmpl",
				meta,
			)

			providerMeta.ResourceNames = append(providerMeta.ResourceNames, utils.ToCamel(k))
		}
	}

	sort.Strings(providerMeta.ResourceNames)
	utils.GenerateFile(
		"../rules/provider_tags.go",
		"../rules/provider_tags.go.tmpl",
		providerMeta,
	)
}
