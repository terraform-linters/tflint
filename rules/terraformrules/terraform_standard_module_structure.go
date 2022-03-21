package terraformrules

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
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

	r.checkFiles(runner)
	r.checkVariables(runner)
	r.checkOutputs(runner)

	return nil
}

func (r *TerraformStandardModuleStructureRule) checkFiles(runner *tflint.Runner) {
	if r.onlyJSON(runner) {
		return
	}

	f := runner.Files()
	files := make(map[string]*hcl.File, len(f))
	for name, file := range f {
		files[filepath.Base(name)] = file
	}

	log.Printf("[DEBUG] %d files found: %v", len(files), files)

	if files[filenameMain] == nil {
		runner.EmitIssue(
			r,
			fmt.Sprintf("Module should include a %s file as the primary entrypoint", filenameMain),
			hcl.Range{
				Filename: filepath.Join(runner.TFConfig.Module.SourceDir, filenameMain),
				Start:    hcl.InitialPos,
			},
		)
	}

	if files[filenameVariables] == nil && len(runner.TFConfig.Module.Variables) == 0 {
		runner.EmitIssue(
			r,
			fmt.Sprintf("Module should include an empty %s file", filenameVariables),
			hcl.Range{
				Filename: filepath.Join(runner.TFConfig.Module.SourceDir, filenameVariables),
				Start:    hcl.InitialPos,
			},
		)
	}

	if files[filenameOutputs] == nil && len(runner.TFConfig.Module.Outputs) == 0 {
		runner.EmitIssue(
			r,
			fmt.Sprintf("Module should include an empty %s file", filenameOutputs),
			hcl.Range{
				Filename: filepath.Join(runner.TFConfig.Module.SourceDir, filenameOutputs),
				Start:    hcl.InitialPos,
			},
		)
	}
}

func (r *TerraformStandardModuleStructureRule) checkVariables(runner *tflint.Runner) {
	for _, variable := range runner.TFConfig.Module.Variables {
		if filename := variable.DeclRange.Filename; r.shouldMove(filename, filenameVariables) {
			runner.EmitIssue(
				r,
				fmt.Sprintf("variable %q should be moved from %s to %s", variable.Name, filename, filenameVariables),
				variable.DeclRange,
			)
		}
	}
}

func (r *TerraformStandardModuleStructureRule) checkOutputs(runner *tflint.Runner) {
	for _, variable := range runner.TFConfig.Module.Outputs {
		if filename := variable.DeclRange.Filename; r.shouldMove(filename, filenameOutputs) {
			runner.EmitIssue(
				r,
				fmt.Sprintf("output %q should be moved from %s to %s", variable.Name, filename, filenameOutputs),
				variable.DeclRange,
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
