// +build tools

package tools

import (
	// This package adds tools to go.mod to ensure that all users have the same versions
	// All imports should be ignored (_)
	_ "golang.org/x/lint/golint"
)
