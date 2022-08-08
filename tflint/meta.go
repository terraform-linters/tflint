package tflint

import "fmt"

// Version is application version
const Version string = "0.39.2"

// ReferenceLink returns the rule reference link
func ReferenceLink(name string) string {
	return fmt.Sprintf("https://github.com/terraform-linters/tflint/blob/v%s/docs/rules/%s.md", Version, name)
}
