package terraformrules

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/terraform-linters/tflint/tflint"
)

// TerraformNamingConventionRule checks whether blocks follow naming convention
type TerraformNamingConventionRule struct{}

type terraformNamingConventionRuleConfig struct {
	Format string `hcl:"format,optional"`
	Custom string `hcl:"custom,optional"`

	CustomFormats []CustomFormatConfig `hcl:"custom_format,block"`

	Data     *BlockFormatConfig `hcl:"data,block"`
	Locals   *BlockFormatConfig `hcl:"locals,block"`
	Module   *BlockFormatConfig `hcl:"module,block"`
	Output   *BlockFormatConfig `hcl:"output,block"`
	Resource *BlockFormatConfig `hcl:"resource,block"`
	Variable *BlockFormatConfig `hcl:"variable,block"`
}

// CustomFormatConfig defines a custom format that can be used instead of the predefined formats
type CustomFormatConfig struct {
	Name        string `hcl:"name,label"`
	Regexp      string `hcl:"regex"`
	Description string `hcl:"description"`
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
	return tflint.NOTICE
}

// Link returns the rule reference link
func (r *TerraformNamingConventionRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

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

	if err := r.checkDataBlocks(runner, &config, defaultNameValidator); err != nil {
		return err
	}

	if err := r.checkLocalValues(runner, &config, defaultNameValidator); err != nil {
		return err
	}

	if err := r.checkModuleBlocks(runner, &config, defaultNameValidator); err != nil {
		return err
	}

	if err := r.checkOutputBlocks(runner, &config, defaultNameValidator); err != nil {
		return err
	}

	if err := r.checkResourceBlocks(runner, &config, defaultNameValidator); err != nil {
		return err
	}

	if err := r.checkVariableBlocks(runner, &config, defaultNameValidator); err != nil {
		return err
	}

	return nil
}

func (r *TerraformNamingConventionRule) checkDataBlocks(runner *tflint.Runner, config *terraformNamingConventionRuleConfig, defaultValidator *NameValidator) error {
	validator := defaultValidator
	if config.Data != nil {
		nameValidator, err := config.Data.getNameValidator(config)
		if err != nil {
			return fmt.Errorf("Invalid data configuration: %v", err)
		}

		validator = nameValidator
	}

	if validator != nil {
		for _, data := range runner.TFConfig.Module.DataResources {
			if !validator.Regexp.MatchString(data.Name) {
				var message string
				if validator.IsNamedFormat {
					message = "data name `%s` must match the following format: %s"
				} else {
					message = "data name `%s` must match the following RegExp: %s"
				}

				runner.EmitIssue(
					r,
					fmt.Sprintf(message, data.Name, validator.Format),
					data.DeclRange,
				)
			}
		}
	}

	return nil
}

func (r *TerraformNamingConventionRule) checkLocalValues(runner *tflint.Runner, config *terraformNamingConventionRuleConfig, defaultValidator *NameValidator) error {
	validator := defaultValidator
	if config.Locals != nil {
		nameValidator, err := config.Locals.getNameValidator(config)
		if err != nil {
			return fmt.Errorf("Invalid locals configuration: %v", err)
		}

		validator = nameValidator
	}

	if validator != nil {
		for _, local := range runner.TFConfig.Module.Locals {
			if !validator.Regexp.MatchString(local.Name) {
				var message string
				if validator.IsNamedFormat {
					message = "local value name `%s` must match the following format: %s"
				} else {
					message = "local value name `%s` must match the following RegExp: %s"
				}

				runner.EmitIssue(
					r,
					fmt.Sprintf(message, local.Name, validator.Format),
					local.DeclRange,
				)
			}
		}
	}

	return nil
}

func (r *TerraformNamingConventionRule) checkModuleBlocks(runner *tflint.Runner, config *terraformNamingConventionRuleConfig, defaultValidator *NameValidator) error {
	validator := defaultValidator
	if config.Module != nil {
		nameValidator, err := config.Module.getNameValidator(config)
		if err != nil {
			return fmt.Errorf("Invalid module configuration: %v", err)
		}

		validator = nameValidator
	}

	if validator != nil {
		for _, module := range runner.TFConfig.Module.ModuleCalls {
			if !validator.Regexp.MatchString(module.Name) {
				var message string
				if validator.IsNamedFormat {
					message = "module name `%s` must match the following format: %s"
				} else {
					message = "module name `%s` must match the following RegExp: %s"
				}

				runner.EmitIssue(
					r,
					fmt.Sprintf(message, module.Name, validator.Format),
					module.DeclRange,
				)
			}
		}
	}

	return nil
}

func (r *TerraformNamingConventionRule) checkOutputBlocks(runner *tflint.Runner, config *terraformNamingConventionRuleConfig, defaultValidator *NameValidator) error {
	validator := defaultValidator
	if config.Output != nil {
		nameValidator, err := config.Output.getNameValidator(config)
		if err != nil {
			return fmt.Errorf("Invalid output configuration: %v", err)
		}

		validator = nameValidator
	}

	if validator != nil {
		for _, output := range runner.TFConfig.Module.Outputs {
			if !validator.Regexp.MatchString(output.Name) {
				var message string
				if validator.IsNamedFormat {
					message = "output name `%s` must match the following format: %s"
				} else {
					message = "output name `%s` must match the following RegExp: %s"
				}

				runner.EmitIssue(
					r,
					fmt.Sprintf(message, output.Name, validator.Format),
					output.DeclRange,
				)
			}
		}
	}

	return nil
}

func (r *TerraformNamingConventionRule) checkResourceBlocks(runner *tflint.Runner, config *terraformNamingConventionRuleConfig, defaultValidator *NameValidator) error {
	validator := defaultValidator
	if config.Resource != nil {
		nameValidator, err := config.Resource.getNameValidator(config)
		if err != nil {
			return fmt.Errorf("Invalid resource configuration: %v", err)
		}

		validator = nameValidator
	}

	if validator != nil {
		for _, resource := range runner.TFConfig.Module.ManagedResources {
			if !validator.Regexp.MatchString(resource.Name) {
				var message string
				if validator.IsNamedFormat {
					message = "resource name `%s` must match the following format: %s"
				} else {
					message = "resource name `%s` must match the following RegExp: %s"
				}

				runner.EmitIssue(
					r,
					fmt.Sprintf(message, resource.Name, validator.Format),
					resource.DeclRange,
				)
			}
		}
	}

	return nil
}

func (r *TerraformNamingConventionRule) checkVariableBlocks(runner *tflint.Runner, config *terraformNamingConventionRuleConfig, defaultValidator *NameValidator) error {
	validator := defaultValidator
	if config.Variable != nil {
		nameValidator, err := config.Variable.getNameValidator(config)
		if err != nil {
			return fmt.Errorf("Invalid variable configuration: %v", err)
		}

		validator = nameValidator
	}

	if validator != nil {
		for _, variable := range runner.TFConfig.Module.Variables {
			if !validator.Regexp.MatchString(variable.Name) {
				var message string
				if validator.IsNamedFormat {
					message = "variable name `%s` must match the following format: %s"
				} else {
					message = "variable name `%s` must match the following RegExp: %s"
				}

				runner.EmitIssue(
					r,
					fmt.Sprintf(message, variable.Name, validator.Format),
					variable.DeclRange,
				)
			}
		}
	}

	return nil
}

func (config *BlockFormatConfig) getNameValidator(ruleConfig *terraformNamingConventionRuleConfig) (*NameValidator, error) {
	return getNameValidator(config.Custom, config.Format, ruleConfig.CustomFormats)
}

func (config *terraformNamingConventionRuleConfig) getNameValidator() (*NameValidator, error) {
	return getNameValidator(config.Custom, config.Format, config.CustomFormats)
}

var snakeCaseRegex = regexp.MustCompile("^[a-z][a-z0-9]*(_[a-z0-9]+)*$")
var mixedSnakeCaseRegex = regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9]*(_[a-zA-Z0-9]+)*$")

func getNameValidator(custom string, format string, customFormats []CustomFormatConfig) (*NameValidator, error) {
	// Prefer custom format if specified
	if custom != "" {
		customRegex, err := regexp.Compile(custom)
		nameValidator := &NameValidator{
			IsNamedFormat: false,
			Format:        custom,
			Regexp:        customRegex,
		}

		return nameValidator, err
	} else if format != "none" {
		for _, customFormat := range customFormats {

			if customFormat.Name == format {
				customRegex, err := regexp.Compile(customFormat.Regexp)
				nameValidator := &NameValidator{
					IsNamedFormat: true,
					Format:        customFormat.Description,
					Regexp:        customRegex,
				}
				return nameValidator, err
			}
		}
		switch strings.ToLower(format) {
		case "snake_case":
			nameValidator := &NameValidator{
				IsNamedFormat: true,
				Format:        format,
				Regexp:        snakeCaseRegex,
			}

			return nameValidator, nil
		case "mixed_snake_case":
			nameValidator := &NameValidator{
				IsNamedFormat: true,
				Format:        format,
				Regexp:        mixedSnakeCaseRegex,
			}

			return nameValidator, nil
		default:
			return nil, fmt.Errorf("`%s` is unsupported format", format)
		}
	}

	return nil, nil
}
