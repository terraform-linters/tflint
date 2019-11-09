// +build !linux,!darwin

package plugin

import "github.com/wata727/tflint/tflint"

// Find returns empty plugins because pkg/plugin is only supported for Linux and macOS
func Find(c *tflint.Config) ([]*Plugin, error) {
	return []*Plugin{}, nil
}
