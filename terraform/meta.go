package terraform

import (
	"bytes"
	"log"
	"os"
	"path/filepath"

	"github.com/terraform-linters/tflint/internal/workingdir"
)

func dataDir() string {
	dir := os.Getenv("TF_DATA_DIR")
	if dir != "" {
		log.Printf("[INFO] TF_DATA_DIR environment variable found: %s", dir)
	} else {
		// The default data dir is always `.terraform` in the current directory
		dir = ".terraform"
	}

	return dir
}

func Workspace() string {
	return WorkspaceAt("")
}

// WorkspaceAt resolves a relative TF_DATA_DIR against dir.
func WorkspaceAt(dir string) string {
	if envVar := os.Getenv("TF_WORKSPACE"); envVar != "" {
		log.Printf("[INFO] TF_WORKSPACE environment variable found: %s", envVar)
		return envVar
	}

	envData, _ := os.ReadFile(workingdir.Resolve(dir, filepath.Join(dataDir(), "environment")))
	current := string(bytes.TrimSpace(envData))
	if current != "" {
		log.Printf("[INFO] environment file found: %s", current)
	} else {
		current = "default"
	}

	return current
}
