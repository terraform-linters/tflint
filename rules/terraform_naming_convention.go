package rules

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint-ruleset-terraform/project"
	"github.com/terraform-linters/tflint-ruleset-terraform/terraform"
)

// TerraformNamingConventionRule checks whether blocks follow naming convention
type TerraformNamingConventionRule struct {
	tflint.DefaultRule
}

type terraformNamingConventionRuleConfig struct {
	Format string `hclext:"format,optional"`
	Custom string `hclext:"custom,optional"`

	CustomFormats map[string]*CustomFormatConfig `hclext:"custom_formats,optional"`

	Data     *BlockFormatConfig `hclext:"data,block"`
	Locals   *BlockFormatConfig `hclext:"locals,block"`
	Module   *BlockFormatConfig `hclext:"module,block"`
	Output   *BlockFormatConfig `hclext:"output,block"`
	Resource *BlockFormatConfig `hclext:"resource,block"`
	Variable *BlockFormatConfig `hclext:"variable,block"`
}

// CustomFormatConfig defines a custom format that can be used instead of the predefined formats
type CustomFormatConfig struct {
	Regexp      string `cty:"regex"`
	Description string `cty:"description"`
}

// BlockFormatConfig defines the pre-defined format or custom regular expression to use
type BlockFormatConfig struct {
	Format string `hclext:"format,optional"`
	Custom string `hclext:"custom,optional"`
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
	return true
}

// Severity returns the rule severity
func (r *TerraformNamingConventionRule) Severity() tflint.Severity {
	return tflint.NOTICE
}

// Link returns the rule reference link
func (r *TerraformNamingConventionRule) Link() string {
	return project.ReferenceLink(r.Name())
}

// Check checks whether blocks follow naming convention
func (r *TerraformNamingConventionRule) Check(rr tflint.Runner) error {
	runner := rr.(*terraform.Runner)

	path, err := runner.GetModulePath()
	if err != nil {
		return err
	}
	if !path.IsRoot() {
		// This rule does not evaluate child modules.
		return nil
	}

	config := terraformNamingConventionRuleConfig{}
	config.Format = "snake_case"
	if err := runner.DecodeRuleConfig(r.Name(), &config); err != nil {
		return err
	}

	defaultNameValidator, err := config.getNameValidator()
	if err != nil {
		return fmt.Errorf("Invalid default configuration: %v", err)
	}

	var nameValidator *NameValidator

	body, err := runner.GetModuleContent(&hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type:       "data",
				LabelNames: []string{"type", "name"},
				Body:       &hclext.BodySchema{},
			},
			{
				Type:       "module",
				LabelNames: []string{"name"},
				Body:       &hclext.BodySchema{},
			},
			{
				Type:       "output",
				LabelNames: []string{"name"},
				Body:       &hclext.BodySchema{},
			},
			{
				Type:       "resource",
				LabelNames: []string{"type", "name"},
				Body:       &hclext.BodySchema{},
			},
			{
				Type:       "variable",
				LabelNames: []string{"name"},
				Body:       &hclext.BodySchema{},
			},
		},
	}, &tflint.GetModuleContentOption{IncludeNotCreated: true})
	if err != nil {
		return err
	}
	blocks := body.Blocks.ByType()

	// data
	dataBlockName := "data"
	nameValidator, err = config.Data.getNameValidator(defaultNameValidator, &config, dataBlockName)
	if err != nil {
		return err
	}
	for _, block := range blocks[dataBlockName] {
		if err := nameValidator.checkBlock(runner, r, dataBlockName, block.Labels[1], &block.DefRange); err != nil {
			return err
		}
	}

	// modules
	moduleBlockName := "module"
	nameValidator, err = config.Module.getNameValidator(defaultNameValidator, &config, moduleBlockName)
	if err != nil {
		return err
	}
	for _, block := range blocks[moduleBlockName] {
		if err := nameValidator.checkBlock(runner, r, moduleBlockName, block.Labels[0], &block.DefRange); err != nil {
			return err
		}
	}

	// outputs
	outputBlockName := "output"
	nameValidator, err = config.Output.getNameValidator(defaultNameValidator, &config, outputBlockName)
	if err != nil {
		return err
	}
	for _, block := range blocks[outputBlockName] {
		if err := nameValidator.checkBlock(runner, r, outputBlockName, block.Labels[0], &block.DefRange); err != nil {
			return err
		}
	}

	// resources
	resourceBlockName := "resource"
	nameValidator, err = config.Resource.getNameValidator(defaultNameValidator, &config, resourceBlockName)
	if err != nil {
		return err
	}
	for _, block := range blocks[resourceBlockName] {
		if err := nameValidator.checkBlock(runner, r, resourceBlockName, block.Labels[1], &block.DefRange); err != nil {
			return err
		}
	}

	// variables
	variableBlockName := "variable"
	nameValidator, err = config.Variable.getNameValidator(defaultNameValidator, &config, variableBlockName)
	if err != nil {
		return err
	}
	for _, block := range blocks[variableBlockName] {
		if err := nameValidator.checkBlock(runner, r, variableBlockName, block.Labels[0], &block.DefRange); err != nil {
			return err
		}
	}

	// locals
	localBlockName := "local value"
	nameValidator, err = config.Locals.getNameValidator(defaultNameValidator, &config, localBlockName)
	if err != nil {
		return err
	}
	locals, diags := runner.GetLocals()
	if diags.HasErrors() {
		return diags
	}
	for name, local := range locals {
		if err := nameValidator.checkBlock(runner, r, localBlockName, name, &local.DefRange); err != nil {
			return err
		}
	}

	return nil
}

func (validator *NameValidator) checkBlock(runner tflint.Runner, r *TerraformNamingConventionRule, blockTypeName string, blockName string, blockDeclRange *hcl.Range) error {
	if validator != nil && !validator.Regexp.MatchString(blockName) {
		var formatType string
		if validator.IsNamedFormat {
			formatType = "format"
		} else {
			formatType = "RegExp"
		}

		return runner.EmitIssue(
			r,
			fmt.Sprintf("%s name `%s` must match the following %s: %s", blockTypeName, blockName, formatType, validator.Format),
			*blockDeclRange,
		)
	}
	return nil
}

func (blockFormatConfig *BlockFormatConfig) getNameValidator(defaultValidator *NameValidator, config *terraformNamingConventionRuleConfig, blockName string) (*NameValidator, error) {
	validator := defaultValidator
	if blockFormatConfig != nil {
		nameValidator, err := getNameValidator(blockFormatConfig.Custom, blockFormatConfig.Format, config)
		if err != nil {
			return nil, fmt.Errorf("Invalid %s configuration: %v", blockName, err)
		}

		validator = nameValidator
	}
	return validator, nil
}

func (config *terraformNamingConventionRuleConfig) getNameValidator() (*NameValidator, error) {
	return getNameValidator(config.Custom, config.Format, config)
}

var predefinedFormats = map[string]*regexp.Regexp{
	"snake_case":       regexp.MustCompile("^[a-z][a-z0-9]*(_[a-z0-9]+)*$"),
	"mixed_snake_case": regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9]*(_[a-zA-Z0-9]+)*$"),
}

func getNameValidator(custom string, format string, config *terraformNamingConventionRuleConfig) (*NameValidator, error) {
	// Prefer custom format if specified
	if custom != "" {
		return getCustomNameValidator(false, custom, custom)
	} else if format != "none" {
		customFormats := config.CustomFormats
		customFormatConfig, exists := customFormats[format]
		if exists {
			return getCustomNameValidator(true, customFormatConfig.Description, customFormatConfig.Regexp)
		}

		regex, exists := predefinedFormats[strings.ToLower(format)]
		if exists {
			nameValidator := &NameValidator{
				IsNamedFormat: true,
				Format:        format,
				Regexp:        regex,
			}
			return nameValidator, nil
		}
		return nil, fmt.Errorf("`%s` is unsupported format", format)
	}

	return nil, nil
}

func getCustomNameValidator(isNamed bool, format, expression string) (*NameValidator, error) {
	regex, err := regexp.Compile(expression)
	nameValidator := &NameValidator{
		IsNamedFormat: isNamed,
		Format:        format,
		Regexp:        regex,
	}
	return nameValidator, err
}
