package terraform

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
)

var DefaultVarsFilename = "terraform.tfvars"

func DataDir() string {
	dir := os.Getenv("TF_DATA_DIR")
	if dir != "" {
		log.Printf("[INFO] TF_DATA_DIR environment variable found: %s", dir)
	} else {
		dir = ".terraform"
	}

	return dir
}

func ModuleDir() string {
	return filepath.Join(DataDir(), "modules")
}

func ModuleManifestPath() string {
	return filepath.Join(ModuleDir(), "modules.json")
}

func Workspace() string {
	if envVar := os.Getenv("TF_WORKSPACE"); envVar != "" {
		log.Printf("[INFO] TF_WORKSPACE environment variable found: %s", envVar)
		return envVar
	}

	envData, _ := os.ReadFile(filepath.Join(DataDir(), "environment"))
	current := string(bytes.TrimSpace(envData))
	if current != "" {
		log.Printf("[INFO] environment file found: %s", current)
	} else {
		current = "default"
	}

	return current
}
