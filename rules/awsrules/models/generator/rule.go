// +build generators

package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	utils "github.com/terraform-linters/tflint/rules/awsrules/generator-utils"
)

type ruleMeta struct {
	RuleName      string
	RuleNameCC    string
	ResourceType  string
	AttributeName string
	Sensitive     bool
	Max           int
	Min           int
	Pattern       string
	Enum          []string
	TestOK        string
	TestNG        string
}

func generateRuleFile(resource, attribute string, model map[string]interface{}, schema *schema.Schema) {
	ruleName := makeRuleName(resource, attribute)

	meta := &ruleMeta{
		RuleName:      ruleName,
		RuleNameCC:    utils.ToCamel(ruleName),
		ResourceType:  resource,
		AttributeName: attribute,
		Sensitive:     schema.Sensitive,
		Max:           fetchNumber(model, "max"),
		Min:           fetchNumber(model, "min"),
		Pattern:       replacePattern(fetchString(model, "pattern")),
		Enum:          fetchStrings(model, "enum"),
	}

	// Testing generated regexp
	regexp.MustCompile(meta.Pattern)

	utils.GenerateFile(fmt.Sprintf("%s.go", ruleName), "pattern_rule.go.tmpl", meta)
}

func generateRuleTestFile(resource, attribute string, model map[string]interface{}, test test) {
	ruleName := makeRuleName(resource, attribute)

	meta := &ruleMeta{
		RuleName:      ruleName,
		RuleNameCC:    utils.ToCamel(ruleName),
		ResourceType:  resource,
		AttributeName: attribute,
		Max:           fetchNumber(model, "max"),
		Min:           fetchNumber(model, "min"),
		Pattern:       replacePattern(fetchString(model, "pattern")),
		Enum:          fetchStrings(model, "enum"),
		TestOK:        formatTest(test.OK),
		TestNG:        formatTest(test.NG),
	}

	// Testing generated regexp
	regexp.MustCompile(meta.Pattern)

	utils.GenerateFile(fmt.Sprintf("%s_test.go", ruleName), "pattern_rule_test.go.tmpl", meta)
}

func makeRuleName(resource, attribute string) string {
	// XXX: Change the naming convention for the backward compatibility.
	if resource == "aws_instance" && attribute == "instance_type" {
		return "aws_instance_invalid_type"
	}
	if resource == "aws_launch_configuration" && attribute == "instance_type" {
		return "aws_launch_configuration_invalid_type"
	}
	return fmt.Sprintf("%s_invalid_%s", resource, attribute)
}

func fetchNumber(model map[string]interface{}, key string) int {
	if v, ok := model[key]; ok {
		return int(v.(float64))
	}
	return 0
}

func fetchStrings(model map[string]interface{}, key string) []string {
	if raw, ok := model[key]; ok {
		list := raw.([]interface{})
		ret := make([]string, len(list))
		for i, v := range list {
			ret[i] = v.(string)
		}
		return ret
	}
	return []string{}
}

func fetchString(model map[string]interface{}, key string) string {
	if v, ok := model[key]; ok {
		return v.(string)
	}
	return ""
}

func replacePattern(pattern string) string {
	if pattern == "" {
		return pattern
	}
	reg := regexp.MustCompile(`\\u([0-9A-F]{4})`)
	replaced := reg.ReplaceAllString(pattern, `\x{$1}`)
	if !strings.HasPrefix(replaced, "^") && !strings.HasSuffix(replaced, "$") {
		return fmt.Sprintf("^%s$", replaced)
	}
	return replaced
}

func formatTest(body string) string {
	if strings.Contains(body, "\n") {
		return fmt.Sprintf("<<TEXT\n%sTEXT", body)
	}
	return fmt.Sprintf(`"%s"`, body)
}
