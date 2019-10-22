// +build !linux,!darwin

package plugin

// Find returns empty plugins because pkg/plugin is only supported for Linux and macOS
func Find() ([]*Plugin, error) {
	return []*Plugin{}, nil
}
