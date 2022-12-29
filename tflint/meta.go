package tflint

import (
	"fmt"

	version "github.com/hashicorp/go-version"
)

// Version is application version
var Version *version.Version = version.Must(version.NewVersion("0.44.1"))

// ReferenceLink returns the rule reference link
func ReferenceLink(name string) string {
	return fmt.Sprintf("https://github.com/terraform-linters/tflint/blob/v%s/docs/rules/%s.md", Version, name)
}
