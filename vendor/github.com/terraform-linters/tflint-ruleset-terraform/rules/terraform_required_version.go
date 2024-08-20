package rules

import (
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint-ruleset-terraform/project"
)

// TerraformRequiredVersionRule checks whether a terraform version has required_version attribute
type TerraformRequiredVersionRule struct {
	tflint.DefaultRule
}

// NewTerraformRequiredVersionRule returns new rule with default attributes
func NewTerraformRequiredVersionRule() *TerraformRequiredVersionRule {
	return &TerraformRequiredVersionRule{}
}

// Name returns the rule name
func (r *TerraformRequiredVersionRule) Name() string {
	return "terraform_required_version"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformRequiredVersionRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *TerraformRequiredVersionRule) Severity() tflint.Severity {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformRequiredVersionRule) Link() string {
	return project.ReferenceLink(r.Name())
}

// Check Checks whether required_version is set
func (r *TerraformRequiredVersionRule) Check(runner tflint.Runner) error {
	path, err := runner.GetModulePath()
	if err != nil {
		return err
	}
	if !path.IsRoot() {
		// This rule does not evaluate child modules.
		return nil
	}

	files, err := runner.GetFiles()
	if err != nil {
		return err
	}
	if len(files) == 0 {
		// This rule does not run on non-Terraform directory.
		return nil
	}

	body, err := runner.GetModuleContent(&hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type: "terraform",
				Body: &hclext.BodySchema{
					Attributes: []hclext.AttributeSchema{{Name: "required_version"}},
				},
			},
		},
	}, &tflint.GetModuleContentOption{ExpandMode: tflint.ExpandModeNone})
	if err != nil {
		return err
	}

	var exists bool

	for _, block := range body.Blocks {
		_, ok := block.Body.Attributes["required_version"]
		exists = exists || ok
	}

	if exists {
		return nil
	}

	if len(body.Blocks) > 0 {
		return r.emitIssue(body.Blocks[0].DefRange, runner)
	}

	// If there are no "terraform" blocks, create a hcl.Range from the files
	var file string
	for k := range files {
		file = k
		break
	}

	// If there is only one file, use that
	if len(files) == 1 {
		return r.emitIssue(hcl.Range{
			Filename: file,
			Start:    hcl.InitialPos,
			End:      hcl.InitialPos,
		}, runner)
	}

	moduleDirectory := filepath.Dir(file)

	// If there are multiple files, look for terraform.tf or main.tf (in that order)
	for _, basename := range []string{"terraform.tf", "main.tf"} {
		filename := filepath.Join(moduleDirectory, basename)
		if _, ok := files[filename]; ok {
			return r.emitIssue(hcl.Range{
				Filename: filename,
				Start:    hcl.InitialPos,
				End:      hcl.InitialPos,
			}, runner)
		}
	}

	// If none of those are found, point to a nonexistent terraform.tf per the style guide
	return r.emitIssue(hcl.Range{
		Filename: filepath.Join(moduleDirectory, "terraform.tf"),
	}, runner)
}

// emitIssue emits issue for missing terraform require version
func (r *TerraformRequiredVersionRule) emitIssue(missingRange hcl.Range, runner tflint.Runner) error {
	return runner.EmitIssue(
		r,
		`terraform "required_version" attribute is required`,
		missingRange,
	)
}
