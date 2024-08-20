package tflint

import "github.com/zclconf/go-cty/cty"

// ModuleCtxType represents target module.
//
//go:generate stringer -type=ModuleCtxType
type ModuleCtxType int32

const (
	// SelfModuleCtxType targets the current module. The default is this behavior.
	SelfModuleCtxType ModuleCtxType = iota
	// RootModuleCtxType targets the root module. This is useful when you want to refer to a provider config.
	RootModuleCtxType
)

// ExpandMode represents whether the block retrieved by GetModuleContent is expanded by the meta-arguments.
//
//go:generate stringer -type=ExpandMode
type ExpandMode int32

const (
	// ExpandModeExpand is the mode for expanding blocks based on the meta-arguments. The default is this behavior.
	ExpandModeExpand ExpandMode = iota
	// ExpandModeNone is the mode that does not expand blocks.
	ExpandModeNone
)

// GetModuleContentOption is an option that controls the behavior when getting a module content.
type GetModuleContentOption struct {
	// Specify the module to be acquired.
	ModuleCtx ModuleCtxType
	// Whether resources and modules are expanded by the count/for_each meta-arguments.
	ExpandMode ExpandMode
	// Hint is info for optimizing a query. This is an advanced option and it is not intended to be used directly from plugins.
	Hint GetModuleContentHint
}

// GetModuleContentHint is info for optimizing a query. This is an advanced option and it is not intended to be used directly from plugins.
type GetModuleContentHint struct {
	ResourceType string
}

// EvaluateExprOption is an option that controls the behavior when evaluating an expression.
type EvaluateExprOption struct {
	// Specify what type of value is expected.
	WantType *cty.Type
	// Set the scope of the module to evaluate.
	ModuleCtx ModuleCtxType
}
