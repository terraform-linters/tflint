package terraformrules

import (
	"fmt"
	"log"
	"regexp"

	"github.com/terraform-linters/tflint/tflint"
)

// TerraformNamingConventionRule checks whether blocks follow naming convention
type TerraformNamingConventionRule struct{}

type terraformNamingConventionRuleConfig struct {
	BlockFormatConfig

	Data     *BlockFormatConfig `hcl:"data,block"`
	Locals   *BlockFormatConfig `hcl:"locals,block"`
	Module   *BlockFormatConfig `hcl:"module,block"`
	Output   *BlockFormatConfig `hcl:"output,block"`
	Resource *BlockFormatConfig `hcl:"resource,block"`
	Variable *BlockFormatConfig `hcl:"variable,block"`
}

// BlockFormatConfig defines the pre-defined format or custom regular expression to use
type BlockFormatConfig struct {
	Format string `hcl:"format,optional"`
	Custom string `hcl:"custom,optional"`
}

// NameValidator contains the regular expression to validate block name, if it was a named format, and the format name/regular expression string
type NameValidator struct {
	Format        string
	IsNamedFormat bool
	Regexp        *regexp.Regexp
}

// NewTerraformNamingConventionRule returns new rule with default attributes
func NewTerraformNamingConventionRule() *TerraformNamingConventionRule {
	return &TerraformNamingConventionRule{}
}

// Name returns the rule name
func (r *TerraformNamingConventionRule) Name() string {
	return "terraform_naming_convention"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformNamingConventionRule) Enabled() bool {
	return false
}

// Severity returns the rule severity
func (r *TerraformNamingConventionRule) Severity() string {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformNamingConventionRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

var reSnakeCase = regexp.MustCompile("^[a-z][a-z0-9]*(_[a-z0-9]+)*$")
var reMixedSnakeCase = regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9]*(_[a-zA-Z0-9]+)*$")

// Check checks whether blocks follow naming convention
func (r *TerraformNamingConventionRule) Check(runner *tflint.Runner) error {
	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	config := terraformNamingConventionRuleConfig{}
	config.Format = "snake_case"
	if err := runner.DecodeRuleConfig(r.Name(), &config); err != nil {
		return err
	}

	defaultNameValidator, err := config.getNameValidator()
	if err != nil {
		return fmt.Errorf("Invalid default configuration: %v", err)
	}

	dataNameValidator := defaultNameValidator
	if config.Data != nil {
		nameValidator, err := config.Data.getNameValidator()
		if err != nil {
			return fmt.Errorf("Invalid data configuration: %v", err)
		}

		dataNameValidator = nameValidator
	}

	localsNameValidator := defaultNameValidator
	if config.Locals != nil {
		nameValidator, err := config.Locals.getNameValidator()
		if err != nil {
			return fmt.Errorf("Invalid locals configuration: %v", err)
		}

		localsNameValidator = nameValidator
	}

	moduleNameValidator := defaultNameValidator
	if config.Module != nil {
		nameValidator, err := config.Module.getNameValidator()
		if err != nil {
			return fmt.Errorf("Invalid module configuration: %v", err)
		}

		moduleNameValidator = nameValidator
	}

	outputNameValidator := defaultNameValidator
	if config.Output != nil {
		nameValidator, err := config.Output.getNameValidator()
		if err != nil {
			return fmt.Errorf("Invalid output configuration: %v", err)
		}

		outputNameValidator = nameValidator
	}

	resourceNameValidator := defaultNameValidator
	if config.Resource != nil {
		nameValidator, err := config.Resource.getNameValidator()
		if err != nil {
			return fmt.Errorf("Invalid resource configuration: %v", err)
		}

		resourceNameValidator = nameValidator
	}

	variableNameValidator := defaultNameValidator
	if config.Variable != nil {
		nameValidator, err := config.Variable.getNameValidator()
		if err != nil {
			return fmt.Errorf("Invalid variable configuration: %v", err)
		}

		variableNameValidator = nameValidator
	}

	// Actually run any checks for modules
	if dataNameValidator != nil {
		for _, data := range runner.TFConfig.Module.DataResources {
			if !dataNameValidator.Regexp.MatchString(data.Name) {
				var message string
				if dataNameValidator.IsNamedFormat {
					message = "data name `%s` must match the following format: %s"
				} else {
					message = "data name `%s` must match the following RegExp: %s"
				}

				runner.EmitIssue(
					r,
					fmt.Sprintf(message, data.Name, dataNameValidator.Format),
					data.DeclRange,
				)
			}
		}
	}

	if localsNameValidator != nil {
		for _, local := range runner.TFConfig.Module.Locals {
			if !localsNameValidator.Regexp.MatchString(local.Name) {
				var message string
				if localsNameValidator.IsNamedFormat {
					message = "local value name `%s` must match the following format: %s"
				} else {
					message = "local value name `%s` must match the following RegExp: %s"
				}

				runner.EmitIssue(
					r,
					fmt.Sprintf(message, local.Name, localsNameValidator.Format),
					local.DeclRange,
				)
			}
		}
	}

	if moduleNameValidator != nil {
		for _, module := range runner.TFConfig.Module.ModuleCalls {
			if !moduleNameValidator.Regexp.MatchString(module.Name) {
				var message string
				if moduleNameValidator.IsNamedFormat {
					message = "module name `%s` must match the following format: %s"
				} else {
					message = "module name `%s` must match the following RegExp: %s"
				}

				runner.EmitIssue(
					r,
					fmt.Sprintf(message, module.Name, moduleNameValidator.Format),
					module.DeclRange,
				)
			}
		}
	}

	if outputNameValidator != nil {
		for _, output := range runner.TFConfig.Module.Outputs {
			if !outputNameValidator.Regexp.MatchString(output.Name) {
				var message string
				if outputNameValidator.IsNamedFormat {
					message = "output name `%s` must match the following format: %s"
				} else {
					message = "output name `%s` must match the following RegExp: %s"
				}

				runner.EmitIssue(
					r,
					fmt.Sprintf(message, output.Name, outputNameValidator.Format),
					output.DeclRange,
				)
			}
		}
	}

	if resourceNameValidator != nil {
		for _, resource := range runner.TFConfig.Module.ManagedResources {
			if !resourceNameValidator.Regexp.MatchString(resource.Name) {
				var message string
				if resourceNameValidator.IsNamedFormat {
					message = "resource name `%s` must match the following format: %s"
				} else {
					message = "resource name `%s` must match the following RegExp: %s"
				}

				runner.EmitIssue(
					r,
					fmt.Sprintf(message, resource.Name, resourceNameValidator.Format),
					resource.DeclRange,
				)
			}
		}
	}

	if variableNameValidator != nil {
		for _, variable := range runner.TFConfig.Module.Variables {
			if !variableNameValidator.Regexp.MatchString(variable.Name) {
				var message string
				if variableNameValidator.IsNamedFormat {
					message = "variable name `%s` must match the following format: %s"
				} else {
					message = "variable name `%s` must match the following RegExp: %s"
				}

				runner.EmitIssue(
					r,
					fmt.Sprintf(message, variable.Name, variableNameValidator.Format),
					variable.DeclRange,
				)
			}
		}
	}

	return nil
}

func (config *BlockFormatConfig) getNameValidator() (*NameValidator, error) {
	if config.Custom != "" {
		nameValidator := &NameValidator{
			IsNamedFormat: false,
			Format:        config.Custom,
			Regexp:        regexp.MustCompile(config.Custom),
		}

		return nameValidator, nil
	} else if config.Format != "" {
		switch config.Format {
		case "snake_case":
			nameValidator := &NameValidator{
				IsNamedFormat: true,
				Format:        config.Format,
				Regexp:        reSnakeCase,
			}

			return nameValidator, nil
		case "mixed_snake_case":
			nameValidator := &NameValidator{
				IsNamedFormat: true,
				Format:        config.Format,
				Regexp:        reSnakeCase,
			}

			return nameValidator, nil
		default:
			return nil, fmt.Errorf("`%s` is unsupported format", config.Format)
		}
	}

	return nil, nil
}
