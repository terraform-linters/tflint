package tflint

import (
	"fmt"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint/terraform"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

// isEvaluableMetaArguments checks whether the passed resource meta-arguments (count/for_each)
// indicate the resource will be evaluated.
// If `count` is 0 or `for_each` is empty, Terraform will not evaluate the attributes of
// that resource.
// Terraform doesn't expect to pass null for these attributes (it will cause an error),
// but we'll treat them as if they were undefined.
func (r *Runner) isEvaluableResource(resource *hclext.Block) (bool, error) {
	if count, exists := resource.Body.Attributes["count"]; exists {
		return r.isEvaluableCountArgument(count.Expr)
	}

	if forEach, exists := resource.Body.Attributes["for_each"]; exists {
		return r.isEvaluableForEachArgument(forEach.Expr)
	}

	// If `count` or `for_each` is not defined, it will be evaluated by default
	return true, nil
}

// isEvaluableModuleCall checks whether the passed module-call meta-arguments (count/for_each)
// indicate the module-call will be evaluated.
// If `count` is 0 or `for_each` is empty, Terraform will not evaluate the attributes of that module.
// Terraform doesn't expect to pass null for these attributes (it will cause an error),
// but we'll treat them as if they were undefined.
func (r *Runner) isEvaluableModuleCall(moduleCall *terraform.ModuleCall) (bool, error) {
	if moduleCall.Count != nil {
		return r.isEvaluableCountArgument(moduleCall.Count)
	}

	if moduleCall.ForEach != nil {
		return r.isEvaluableForEachArgument(moduleCall.ForEach)
	}

	// If `count` or `for_each` is not defined, it will be evaluated by default
	return true, nil
}

func (r *Runner) isEvaluableCountArgument(expr hcl.Expression) (bool, error) {
	val, diags := r.Ctx.EvaluateExpr(expr, cty.DynamicPseudoType)
	if diags.HasErrors() {
		return false, diags
	}
	val, _ = val.Unmark()

	if val.IsNull() {
		// null value means that attribute is not set
		return true, nil
	}
	if !val.IsKnown() {
		// unknown value is non-deterministic
		return false, nil
	}

	count := 1
	if err := gocty.FromCtyValue(val, &count); err != nil {
		return false, err
	}
	if count == 0 {
		// `count = 0` is not evaluated
		return false, nil
	}
	// `count > 1` is evaluated`
	return true, nil
}

func (r *Runner) isEvaluableForEachArgument(expr hcl.Expression) (bool, error) {
	val, diags := r.Ctx.EvaluateExpr(expr, cty.DynamicPseudoType)
	if diags.HasErrors() {
		return false, diags
	}

	if val.IsNull() {
		// null value means that attribute is not set
		return true, nil
	}
	if !val.IsKnown() {
		// unknown value is non-deterministic
		return false, nil
	}
	if !val.CanIterateElements() {
		// uniteratable values (string, number, etc.) are invalid
		return false, fmt.Errorf("The `for_each` value is not iterable in %s:%d", expr.Range().Filename, expr.Range().Start.Line)
	}
	if val.LengthInt() == 0 {
		// empty `for_each` is not evaluated
		return false, nil
	}
	// `for_each` with non-empty elements is evaluated
	return true, nil
}
