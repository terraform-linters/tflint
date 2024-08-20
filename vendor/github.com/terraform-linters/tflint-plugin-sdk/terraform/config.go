package terraform

import "strings"

// IsJSONFilename returns true if the filename is a JSON syntax file.
func IsJSONFilename(filename string) bool {
	return strings.HasSuffix(filename, ".tf.json")
}
