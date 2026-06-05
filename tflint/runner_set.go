package tflint

import (
	"fmt"

	"github.com/terraform-linters/tflint/terraform"
)

// BuildRunners loads the module rooted at dir using the given loader and config,
// returning the root runner and its module runners.
func BuildRunners(loader *terraform.Loader, config *Config, workingDir, dir string) (*Runner, []*Runner, error) {
	rootMod, diags := loader.LoadRootModule(dir)
	if diags.HasErrors() {
		return nil, []*Runner{}, fmt.Errorf("Failed to load the root module; %w", diags)
	}

	files, diags := loader.LoadConfigDirFiles(dir)
	if diags.HasErrors() {
		return nil, []*Runner{}, fmt.Errorf("Failed to list configuration files; %w", diags)
	}
	annotations := map[string]Annotations{}
	for path, file := range files {
		ants, lexDiags := NewAnnotations(path, file)
		diags = diags.Extend(lexDiags)
		annotations[path] = ants
	}
	if diags.HasErrors() {
		return nil, []*Runner{}, fmt.Errorf("Failed to parse annotations; %w", diags)
	}

	variables, diags := loader.LoadValuesFiles(dir, config.Varfiles...)
	if diags.HasErrors() {
		return nil, []*Runner{}, fmt.Errorf("Failed to load values files; %w", diags)
	}
	cliVars, diags := terraform.ParseVariableValues(config.Variables, rootMod.Variables)
	if diags.HasErrors() {
		return nil, []*Runner{}, fmt.Errorf("Failed to parse variables; %w", diags)
	}
	variables = append(variables, cliVars)

	configs, diags := terraform.BuildConfig(
		rootMod,
		loader.ModuleWalker(config.CallModuleType),
		workingDir,
		variables...,
	)
	if diags.HasErrors() {
		return nil, []*Runner{}, fmt.Errorf("Failed to build configurations; %w", diags)
	}

	runner, err := NewRunner(workingDir, config, annotations, configs, variables...)
	if err != nil {
		return nil, []*Runner{}, fmt.Errorf("Failed to initialize a runner; %w", err)
	}

	moduleRunners, err := NewModuleRunners(runner)
	if err != nil {
		return nil, []*Runner{}, fmt.Errorf("Failed to prepare rule checking; %w", err)
	}

	return runner, moduleRunners, nil
}
