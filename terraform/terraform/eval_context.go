package terraform

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/terraform/addrs"
	"github.com/terraform-linters/tflint/terraform/configs/configschema"
	"github.com/terraform-linters/tflint/terraform/lang"
	"github.com/terraform-linters/tflint/terraform/tfdiags"
	"github.com/zclconf/go-cty/cty"
)

// EvalContext is the interface that is given to eval nodes to execute.
type EvalContext interface {
	// Path is the current module path.
	Path() addrs.ModuleInstance

	// EvaluateBlock takes the given raw configuration block and associated
	// schema and evaluates it to produce a value of an object type that
	// conforms to the implied type of the schema.
	//
	// The "self" argument is optional. If given, it is the referenceable
	// address that the name "self" should behave as an alias for when
	// evaluating. Set this to nil if the "self" object should not be available.
	//
	// The "key" argument is also optional. If given, it is the instance key
	// of the current object within the multi-instance container it belongs
	// to. For example, on a resource block with "count" set this should be
	// set to a different addrs.IntKey for each instance created from that
	// block. Set this to addrs.NoKey if not appropriate.
	//
	// The returned body is an expanded version of the given body, with any
	// "dynamic" blocks replaced with zero or more static blocks. This can be
	// used to extract correct source location information about attributes of
	// the returned object value.
	EvaluateBlock(body hcl.Body, schema *configschema.Block, self addrs.Referenceable, keyData InstanceKeyEvalData) (cty.Value, hcl.Body, tfdiags.Diagnostics)

	// EvaluateExpr takes the given HCL expression and evaluates it to produce
	// a value.
	//
	// The "self" argument is optional. If given, it is the referenceable
	// address that the name "self" should behave as an alias for when
	// evaluating. Set this to nil if the "self" object should not be available.
	EvaluateExpr(expr hcl.Expression, wantType cty.Type, self addrs.Referenceable) (cty.Value, tfdiags.Diagnostics)

	// EvaluationScope returns a scope that can be used to evaluate reference
	// addresses in this context.
	EvaluationScope(self addrs.Referenceable, keyData InstanceKeyEvalData) *lang.Scope

	// WithPath returns a copy of the context with the internal path set to the
	// path argument.
	WithPath(path addrs.ModuleInstance) EvalContext
}
