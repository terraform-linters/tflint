package tflint

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hclparse"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/terraform"
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

// getAutoLoadValuesFiles returns all files which match *.auto.tfvars present in the given directory
// The list is sorted alphabetically except `terraform.tfvars`. This is equivalent to priority
func getAutoLoadValuesFiles(dir string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	ret := []string{}
	var foundDefaultFile bool
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if file.Name() == defaultValuesFile {
			foundDefaultFile = true
		} else if isAutoloadValuesFile(file.Name()) {
			ret = append(ret, file.Name())
		}
	}
	sort.Strings(ret)

	if foundDefaultFile {
		// `terraform.tfvars` has the lowest priority
		ret = append([]string{defaultValuesFile}, ret...)
	}

	return ret, nil
}

func isConfigFile(name string) bool {
	return strings.HasSuffix(name, ".tf") || strings.HasSuffix(name, ".tf.json")
}

func isOverrideConfigFile(name string) bool {
	var ext string
	if strings.HasSuffix(name, ".tf") {
		ext = ".tf"
	} else if strings.HasSuffix(name, ".tf.json") {
		ext = ".tf.json"
	}

	if ext == "" {
		return false
	}
	basename := name[:len(name)-len(ext)]
	return basename == "override" || strings.HasSuffix(basename, "_override")
}

func isValuesFile(name string) bool {
	return strings.HasSuffix(name, ".tfvars") || strings.HasSuffix(name, ".tfvars.json")
}

func isAutoloadValuesFile(name string) bool {
	return strings.HasSuffix(name, ".auto.tfvars") || strings.HasSuffix(name, ".auto.tfvars.json") || name == defaultValuesFile
}

/************************************************************************
  The following functions are expected to be merged into Terraform core
*************************************************************************/

// https://github.com/hashicorp/terraform/blob/v0.12.6/configs/parser.go
func ParseHCLFile(path string, src []byte) (hcl.Body, hcl.Diagnostics) {
	parser := hclparse.NewParser()

	var file *hcl.File
	var diags hcl.Diagnostics
	switch {
	case strings.HasSuffix(path, ".json"):
		file, diags = parser.ParseJSON(src, path)
	default:
		file, diags = parser.ParseHCL(src, path)
	}

	// If the returned file or body is nil, then we'll return a non-nil empty
	// body so we'll meet our contract that nil means an error reading the file.
	if file == nil || file.Body == nil {
		return hcl.EmptyBody(), diags
	}

	return file.Body, diags
}

// https://github.com/hashicorp/terraform/blob/v0.12.6/configs/parser_config.go
func BuildConfigFile(body hcl.Body, override bool) (*configs.File, hcl.Diagnostics) {
	file := &configs.File{}
	var diags hcl.Diagnostics

	var reqDiags hcl.Diagnostics
	file.CoreVersionConstraints, reqDiags = sniffCoreVersionRequirements(body)
	diags = append(diags, reqDiags...)

	content, contentDiags := body.Content(configFileSchema)
	diags = append(diags, contentDiags...)

	for _, block := range content.Blocks {
		switch block.Type {

		case "terraform":
			content, contentDiags := block.Body.Content(terraformBlockSchema)
			diags = append(diags, contentDiags...)

			// We ignore the "terraform_version" attribute here because
			// sniffCoreVersionRequirements already dealt with that above.

			for _, innerBlock := range content.Blocks {
				switch innerBlock.Type {

				case "backend":
					backendCfg, cfgDiags := decodeBackendBlock(innerBlock)
					diags = append(diags, cfgDiags...)
					if backendCfg != nil {
						file.Backends = append(file.Backends, backendCfg)
					}

				case "required_providers":
					reqs, reqsDiags := decodeRequiredProvidersBlock(innerBlock)
					diags = append(diags, reqsDiags...)
					file.ProviderRequirements = append(file.ProviderRequirements, reqs...)

				default:
					// Should never happen because the above cases should be exhaustive
					// for all block type names in our schema.
					continue

				}
			}

		case "provider":
			cfg, cfgDiags := decodeProviderBlock(block)
			diags = append(diags, cfgDiags...)
			if cfg != nil {
				file.ProviderConfigs = append(file.ProviderConfigs, cfg)
			}

		case "variable":
			cfg, cfgDiags := decodeVariableBlock(block, override)
			diags = append(diags, cfgDiags...)
			if cfg != nil {
				file.Variables = append(file.Variables, cfg)
			}

		case "locals":
			defs, defsDiags := decodeLocalsBlock(block)
			diags = append(diags, defsDiags...)
			file.Locals = append(file.Locals, defs...)

		case "output":
			cfg, cfgDiags := decodeOutputBlock(block, override)
			diags = append(diags, cfgDiags...)
			if cfg != nil {
				file.Outputs = append(file.Outputs, cfg)
			}

		case "module":
			cfg, cfgDiags := decodeModuleBlock(block, override)
			diags = append(diags, cfgDiags...)
			if cfg != nil {
				file.ModuleCalls = append(file.ModuleCalls, cfg)
			}

		case "resource":
			cfg, cfgDiags := decodeResourceBlock(block)
			diags = append(diags, cfgDiags...)
			if cfg != nil {
				file.ManagedResources = append(file.ManagedResources, cfg)
			}

		case "data":
			cfg, cfgDiags := decodeDataBlock(block)
			diags = append(diags, cfgDiags...)
			if cfg != nil {
				file.DataResources = append(file.DataResources, cfg)
			}

		default:
			// Should never happen because the above cases should be exhaustive
			// for all block type names in our schema.
			continue

		}
	}

	return file, diags
}

// https://github.com/hashicorp/terraform/blob/v0.12.6/configs/parser_values.go
func BuildValuesFile(body hcl.Body) (map[string]cty.Value, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	vals := make(map[string]cty.Value)
	attrs, attrDiags := body.JustAttributes()
	diags = append(diags, attrDiags...)
	if attrs == nil {
		return vals, diags
	}

	for name, attr := range attrs {
		val, valDiags := attr.Expr.Value(nil)
		diags = append(diags, valDiags...)
		vals[name] = val
	}

	return vals, diags
}
