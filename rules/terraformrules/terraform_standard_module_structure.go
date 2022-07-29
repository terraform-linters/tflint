package terraformrules

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	sdk "github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint/tflint"
)

const (
	filenameMain      = "main.tf"
	filenameVariables = "variables.tf"
	filenameOutputs   = "outputs.tf"
)

// TerraformStandardModuleStructureRule checks whether modules adhere to Terraform's standard module structure
type TerraformStandardModuleStructureRule struct{}

// NewTerraformStandardModuleStructureRule returns a new rule
func NewTerraformStandardModuleStructureRule() *TerraformStandardModuleStructureRule {
	return &TerraformStandardModuleStructureRule{}
}

// Name returns the rule name
func (r *TerraformStandardModuleStructureRule) Name() string {
	return "terraform_standard_module_structure"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformStandardModuleStructureRule) Enabled() bool {
	return false
}

// Severity returns the rule severity
func (r *TerraformStandardModuleStructureRule) Severity() tflint.Severity {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformStandardModuleStructureRule) Link() string {
	return tflint.ReferenceLink(r.Name())
}

// Check emits errors for any missing files and any block types that are included in the wrong file
func (r *TerraformStandardModuleStructureRule) Check(runner *tflint.Runner) error {
	if !runner.TFConfig.Path.IsRoot() {
		// This rule does not evaluate child modules.
		return nil
	}

	log.Printf("[TRACE] Check `%s` rule for `%s` runner", r.Name(), runner.TFConfigPath())

	body, diags := runner.GetModuleContent(&hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type:       "variable",
				LabelNames: []string{"name"},
				Body:       &hclext.BodySchema{},
			},
			{
				Type:       "output",
				LabelNames: []string{"name"},
				Body:       &hclext.BodySchema{},
			},
		},
	}, sdk.GetModuleContentOption{IncludeNotCreated: true})
	if diags.HasErrors() {
		return diags
	}

	blocks := body.Blocks.ByType()

	r.checkFiles(runner, body.Blocks)
	r.checkVariables(runner, blocks["variable"])
	r.checkOutputs(runner, blocks["output"])

	return nil
}

func (r *TerraformStandardModuleStructureRule) checkFiles(runner *tflint.Runner, blocks hclext.Blocks) {
	if r.onlyJSON(runner) {
		return
	}

	f := runner.Files()
	var dir string
	files := make(map[string]*hcl.File, len(f))
	for name, file := range f {
		dir = filepath.Dir(name)
		files[filepath.Base(name)] = file
	}

	log.Printf("[DEBUG] %d files found: %v", len(files), files)

	if files[filenameMain] == nil {
		runner.EmitIssue(
			r,
			fmt.Sprintf("Module should include a %s file as the primary entrypoint", filenameMain),
			hcl.Range{
				Filename: filepath.Join(dir, filenameMain),
				Start:    hcl.InitialPos,
			},
		)
	}

	if files[filenameVariables] == nil && len(blocks.ByType()["variable"]) == 0 {
		runner.EmitIssue(
			r,
			fmt.Sprintf("Module should include an empty %s file", filenameVariables),
			hcl.Range{
				Filename: filepath.Join(dir, filenameVariables),
				Start:    hcl.InitialPos,
			},
		)
	}

	if files[filenameOutputs] == nil && len(blocks.ByType()["output"]) == 0 {
		runner.EmitIssue(
			r,
			fmt.Sprintf("Module should include an empty %s file", filenameOutputs),
			hcl.Range{
				Filename: filepath.Join(dir, filenameOutputs),
				Start:    hcl.InitialPos,
			},
		)
	}
}

func (r *TerraformStandardModuleStructureRule) checkVariables(runner *tflint.Runner, variables hclext.Blocks) {
	for _, variable := range variables {
		if filename := variable.DefRange.Filename; r.shouldMove(filename, filenameVariables) {
			runner.EmitIssue(
				r,
				fmt.Sprintf("variable %q should be moved from %s to %s", variable.Labels[0], filename, filenameVariables),
				variable.DefRange,
			)
		}
	}
}

func (r *TerraformStandardModuleStructureRule) checkOutputs(runner *tflint.Runner, outputs hclext.Blocks) {
	for _, output := range outputs {
		if filename := output.DefRange.Filename; r.shouldMove(filename, filenameOutputs) {
			runner.EmitIssue(
				r,
				fmt.Sprintf("output %q should be moved from %s to %s", output.Labels[0], filename, filenameOutputs),
				output.DefRange,
			)
		}
	}
}

func (r *TerraformStandardModuleStructureRule) onlyJSON(runner *tflint.Runner) bool {
	files := runner.Files()

	if len(files) == 0 {
		return false
	}

	for filename := range files {
		if filepath.Ext(filename) != ".json" {
			return false
		}
	}

	return true
}

func (r *TerraformStandardModuleStructureRule) shouldMove(path string, expected string) bool {
	// json files are likely generated and conventional filenames do not apply
	if filepath.Ext(path) == ".json" {
		return false
	}

	return filepath.Base(path) != expected
}
