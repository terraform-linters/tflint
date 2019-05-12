package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/wata727/tflint/tools/utils"
)

type metadata struct {
	RuleName   string
	RuleNameCC string
}

func main() {
	buf := bufio.NewReader(os.Stdin)
	fmt.Print("Rule name? (e.g. aws_instance_invalid_type): ")
	ruleName, err := buf.ReadString('\n')
	if err != nil {
		panic(err)
	}
	ruleName = strings.Trim(ruleName, "\n")

	meta := &metadata{RuleNameCC: utils.ToCamel(ruleName), RuleName: ruleName}

	utils.GenerateFileWithLogs(fmt.Sprintf("rules/awsrules/%s.go", ruleName), "rules/rule.go.tmpl", meta)
	utils.GenerateFileWithLogs(fmt.Sprintf("rules/awsrules/%s_test.go", ruleName), "rules/rule_test.go.tmpl", meta)
}
