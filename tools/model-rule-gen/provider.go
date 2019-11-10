package main

import "github.com/terraform-linters/tflint/tools/utils"

type providerMeta struct {
	RuleNameCCList []string
}

func generateProviderFile(ruleNames []string) {
	meta := &providerMeta{}

	for _, ruleName := range ruleNames {
		meta.RuleNameCCList = append(meta.RuleNameCCList, utils.ToCamel(ruleName))
	}

	utils.GenerateFile("../rules/provider_model.go", "../rules/provider_model.go.tmpl", meta)
}
