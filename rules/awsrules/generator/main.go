package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/terraform-linters/tflint/tools/utils"
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

	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filepath.Dir(file))

	utils.GenerateFileWithLogs(fmt.Sprintf("%s/%s.go", dir, ruleName), fmt.Sprintf("%s/rule.go.tmpl", dir), meta)
	utils.GenerateFileWithLogs(fmt.Sprintf("%s/%s_test.go", dir, ruleName), fmt.Sprintf("%s/rule_test.go.tmpl", dir), meta)
}
