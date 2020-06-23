package main

import utils "github.com/terraform-linters/tflint/rules/awsrules/generator-utils"

type providerMeta struct {
	RuleNameCCList []string
}

func generateProviderFile(ruleNames []string) {
	meta := &providerMeta{}

	for _, ruleName := range ruleNames {
		meta.RuleNameCCList = append(meta.RuleNameCCList, utils.ToCamel(ruleName))
	}

	utils.GenerateFile("../../provider_model.go", "../../provider_model.go.tmpl", meta)
}
