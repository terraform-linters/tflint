package tflint

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/terraform/configs"
	"github.com/terraform-linters/tflint/terraform/terraform"
)

var defaultValuesFile = "terraform.tfvars"

// ParseTFVariables parses the passed Terraform variable CLI arguments, and returns terraform.InputValues
func ParseTFVariables(vars []string, declVars map[string]*configs.Variable) (terraform.InputValues, error) {
	variables := make(terraform.InputValues)
	for _, raw := range vars {
		idx := strings.Index(raw, "=")
		if idx == -1 {
			return variables, fmt.Errorf("`%s` is invalid. Variables must be `key=value` format", raw)
		}
		name := raw[:idx]
		rawVal := raw[idx+1:]

		var mode configs.VariableParsingMode
		declVar, declared := declVars[name]
		if declared {
			mode = declVar.ParsingMode
		} else {
			mode = configs.VariableParseLiteral
		}

		val, diags := mode.Parse(name, rawVal)
		if diags.HasErrors() {
			return variables, diags
		}

		variables[name] = &terraform.InputValue{
			Value:      val,
			SourceType: terraform.ValueFromCLIArg,
		}
	}

	return variables, nil
}

func getTFDataDir() string {
	dir := os.Getenv("TF_DATA_DIR")
	if dir != "" {
		log.Printf("[INFO] TF_DATA_DIR environment variable found: %s", dir)
	} else {
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
		log.Printf("[INFO] TF_WORKSPACE environment variable found: %s", envVar)
		return envVar
	}

	envData, _ := os.ReadFile(filepath.Join(getTFDataDir(), "environment"))
	current := string(bytes.TrimSpace(envData))
	if current != "" {
		log.Printf("[INFO] environment file found: %s", current)
	} else {
		current = "default"
	}

	return current
}

func getTFEnvVariables(declVars map[string]*configs.Variable) (terraform.InputValues, hcl.Diagnostics) {
	envVariables := make(terraform.InputValues)
	var diags hcl.Diagnostics

	for _, e := range os.Environ() {
		idx := strings.Index(e, "=")
		envKey := e[:idx]
		envVal := e[idx+1:]

		if strings.HasPrefix(envKey, "TF_VAR_") {
			log.Printf("[INFO] TF_VAR_* environment variable found: key=%s", envKey)
			varName := strings.Replace(envKey, "TF_VAR_", "", 1)

			var mode configs.VariableParsingMode
			declVar, declared := declVars[varName]
			if declared {
				mode = declVar.ParsingMode
			} else {
				mode = configs.VariableParseLiteral
			}

			val, parseDiags := mode.Parse(varName, envVal)
			if parseDiags.HasErrors() {
				diags = diags.Extend(parseDiags)
				continue
			}

			envVariables[varName] = &terraform.InputValue{
				Value:      val,
				SourceType: terraform.ValueFromEnvVar,
			}
		}
	}

	return envVariables, diags
}
