package tflint

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/json"
	"github.com/terraform-linters/tflint/terraform/configs"
	"github.com/terraform-linters/tflint/terraform/terraform"
	"github.com/zclconf/go-cty/cty"
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

// ParseExpression is a wrapper for a function that parses JSON and HCL expressions
func ParseExpression(src []byte, filename string, start hcl.Pos) (hcl.Expression, hcl.Diagnostics) {
	if strings.HasSuffix(filename, ".tf") {
		// HACK: Always add a newline to avoid heredoc parse errors.
		// @see https://github.com/hashicorp/hcl/issues/441
		src = []byte(string(src) + "\n")
		return hclsyntax.ParseExpression(src, filename, start)
	}

	if strings.HasSuffix(filename, ".tf.json") {
		return json.ParseExpressionWithStartPos(src, filename, start)
	}

	panic(fmt.Sprintf("Unexpected file: %s", filename))
}

// HCLBodyRange attempts to find a range of the passed body
func HCLBodyRange(body hcl.Body, defRange hcl.Range) hcl.Range {
	if strings.HasSuffix(defRange.Filename, ".tf") {
		var bodyRange hcl.Range
		bodyRange.Filename = defRange.Filename

		// Estimate the range of the body from the range of all attributes and blocks.
		hclBody, ok := body.(*hclsyntax.Body)
		if !ok {
			// BUG: If the body is overridden, the structure that satisfies hcl.Body interface may be configs.mergeBody.
			// In that case, the definition range of the body cannot be acquired by this way because the range of the body spans multiple files.
			// To avoid panic, here we return an empty range with only the filename set.
			// As a result, plugins that use this range to get hcl.Body may have incorrect results.
			// This issue will be fixed by changing the way of transffering the hcl.Body.
			// See also https://github.com/terraform-linters/tflint-plugin-sdk/issues/89.
			return bodyRange
		}

		for _, attr := range hclBody.Attributes {
			if bodyRange.Empty() {
				bodyRange = attr.Range()
			} else {
				bodyRange = hcl.RangeOver(bodyRange, attr.Range())
			}
		}
		for _, block := range hclBody.Blocks {
			if bodyRange.Empty() {
				bodyRange = block.Range()
			} else {
				bodyRange = hcl.RangeOver(bodyRange, block.Range())
			}
		}
		return bodyRange
	}

	if strings.HasSuffix(defRange.Filename, ".tf.json") {
		// HACK: In JSON syntax, DefRange corresponds to open brace and MissingItemRange corresponds to close brace.
		return hcl.RangeOver(defRange, body.MissingItemRange())
	}

	panic(fmt.Sprintf("Unexpected file: %s", defRange.Filename))
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

	envData, _ := ioutil.ReadFile(filepath.Join(getTFDataDir(), "environment"))
	current := string(bytes.TrimSpace(envData))
	if current != "" {
		log.Printf("[INFO] environment file found: %s", current)
	} else {
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
			log.Printf("[INFO] TF_VAR_* environment variable found: key=%s", envKey)
			varName := strings.Replace(envKey, "TF_VAR_", "", 1)

			envVariables[varName] = &terraform.InputValue{
				Value:      cty.StringVal(envVal),
				SourceType: terraform.ValueFromEnvVar,
			}
		}
	}

	return envVariables
}
