package terraform

import (
	"github.com/terraform-linters/tflint/terraform/addrs"
)

// GraphNodeModuleInstance says that a node is part of a graph with a
// different path, and the context should be adjusted accordingly.
type GraphNodeModuleInstance interface {
	Path() addrs.ModuleInstance
}

// GraphNodeModulePath is implemented by all referenceable nodes, to indicate
// their configuration path in unexpanded modules.
type GraphNodeModulePath interface {
	ModulePath() addrs.Module
}
