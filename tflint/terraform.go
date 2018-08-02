package tflint

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform/terraform"
	"github.com/zclconf/go-cty/cty"
)

func getTFDataDir() string {
	dir := os.Getenv("TF_DATA_DIR")
	if dir == "" {
		dir = ".terraform"
	}

	return dir
}

func getTFModuleDir() string {
	return filepath.Join(getTFDataDir(), "modules")
}

func getTFModuleManifestPath() string {
	return filepath.Join(getTFModuleDir(), "modules.json")
}

func getTFWorkspace() string {
	if envVar := os.Getenv("TF_WORKSPACE"); envVar != "" {
		return envVar
	}

	envData, _ := ioutil.ReadFile(filepath.Join(getTFDataDir(), "environment"))
	current := string(bytes.TrimSpace(envData))
	if current == "" {
		current = "default"
	}

	return current
}

func getTFEnvVariables() terraform.InputValues {
	envVariables := make(terraform.InputValues)
	for _, e := range os.Environ() {
		idx := strings.Index(e, "=")
		envKey := e[:idx]
		envVal := e[idx+1:]

		if strings.HasPrefix(envKey, "TF_VAR_") {
			varName := strings.Replace(envKey, "TF_VAR_", "", 1)

			envVariables[varName] = &terraform.InputValue{
				Value:      cty.StringVal(envVal),
				SourceType: terraform.ValueFromEnvVar,
			}
		}
	}

	return envVariables
}
