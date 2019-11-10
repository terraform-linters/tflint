// +build !linux,!darwin

package plugin

import "github.com/terraform-linters/tflint/tflint"

// Find returns empty plugins because pkg/plugin is only supported for Linux and macOS
func Find(c *tflint.Config) ([]*Plugin, error) {
	return []*Plugin{}, nil
}
