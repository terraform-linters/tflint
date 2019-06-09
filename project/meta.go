package project

import "fmt"

// Version is application version
const Version string = "0.8.3"

// ReferenceLink returns the rule reference link
func ReferenceLink(name string) string {
	return fmt.Sprintf("https://github.com/wata727/tflint/blob/v%s/docs/rules/%s.md", Version, name)
}
