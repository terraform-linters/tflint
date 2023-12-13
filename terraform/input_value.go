package terraform

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

type InputValue struct {
	Value cty.Value
}

type InputValues map[string]*InputValue

func (vv InputValues) Override(others ...InputValues) InputValues {
	ret := make(InputValues)
	for k, v := range vv {
		ret[k] = v
	}
	for _, other := range others {
		for k, v := range other {
			ret[k] = v
		}
	}
	return ret
}

// DefaultVariableValues returns InputValues using the default values
// of variables declared in the configuration.
func DefaultVariableValues(configs map[string]*Variable) InputValues {
	ret := make(InputValues)
	for k, c := range configs {
		val := c.Default
		// cty.NilVal means no default declared in the variable. Terraform collects this value interactively,
		// while TFLint marks it as unknown and continues inspection.
		if c.Default == cty.NilVal {
			val = cty.UnknownVal(c.Type)
		}

		ret[k] = &InputValue{
			Value: val,
		}
	}
	return ret
}

// EnvironmentVariableValues looks up `TF_VAR_*` env variables and returns InputValues.
// Declared variables are required because the parsing mode of the variable value is type-dependent.
func EnvironmentVariableValues(declVars map[string]*Variable) (InputValues, hcl.Diagnostics) {
	envVariables := make(InputValues)
	var diags hcl.Diagnostics

	for _, e := range os.Environ() {
		idx := strings.Index(e, "=")
		envKey := e[:idx]
		envVal := e[idx+1:]

		if strings.HasPrefix(envKey, "TF_VAR_") {
			log.Printf("[INFO] TF_VAR_* environment variable found: key=%s", envKey)
			varName := strings.Replace(envKey, "TF_VAR_", "", 1)

			var mode VariableParsingMode
			declVar, declared := declVars[varName]
			if declared {
				mode = declVar.ParsingMode
			} else {
				mode = VariableParseLiteral
			}

			val, parseDiags := mode.Parse(varName, envVal)
			if parseDiags.HasErrors() {
				diags = diags.Extend(parseDiags)
				continue
			}

			envVariables[varName] = &InputValue{
				Value: val,
			}
		}
	}

	return envVariables, diags
}

// ParseVariableValues parses the variable values passed as CLI flags and returns InputValues.
// Declared variables are required because the parsing mode of the variable value is type-dependent.
func ParseVariableValues(vars []string, declVars map[string]*Variable) (InputValues, hcl.Diagnostics) {
	variables := make(InputValues)
	var diags hcl.Diagnostics

	for _, raw := range vars {
		idx := strings.Index(raw, "=")
		if idx == -1 {
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "invalid variable value format",
				Detail:   fmt.Sprintf(`"%s" is invalid. Variables must be "key=value" format`, raw),
			})
			continue
		}
		name := raw[:idx]
		rawVal := raw[idx+1:]

		var mode VariableParsingMode
		declVar, declared := declVars[name]
		if declared {
			mode = declVar.ParsingMode
		} else {
			mode = VariableParseLiteral
		}

		val, parseDiags := mode.Parse(name, rawVal)
		if parseDiags.HasErrors() {
			diags = diags.Extend(parseDiags)
			continue
		}

		variables[name] = &InputValue{
			Value: val,
		}
	}

	return variables, diags
}

// VariableValues returns a value map based on configuration, environment variables,
// and external input values. External input values take precedence over configuration defaults,
// environment variables, and the last one passed takes precedence.
func VariableValues(config *Config, values ...InputValues) (map[string]map[string]cty.Value, hcl.Diagnostics) {
	moduleKey := config.Path.UnkeyedInstanceShim().String()
	variableValues := make(map[string]map[string]cty.Value)
	variableValues[moduleKey] = make(map[string]cty.Value)

	variables := DefaultVariableValues(config.Module.Variables)
	envVars, diags := EnvironmentVariableValues(config.Module.Variables)
	if diags.HasErrors() {
		return variableValues, diags
	}
	overrideVariables := variables.Override(envVars).Override(values...)

	for k, iv := range overrideVariables {
		variableValues[moduleKey][k] = iv.Value
	}
	return variableValues, nil
}
