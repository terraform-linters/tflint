package main

import "github.com/wata727/tflint/tools/utils"

type rulesetMeta struct {
	RuleNameCCList []string
}

func generateRuleSetFile(ruleNames []string) {
	meta := &rulesetMeta{}

	for _, ruleName := range ruleNames {
		meta.RuleNameCCList = append(meta.RuleNameCCList, utils.ToCamel(ruleName))
	}

	utils.GenerateFile("../rules/ruleset_model.go", "../rules/ruleset_model.go.tmpl", meta)
}
