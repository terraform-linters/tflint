package tflint

import (
	"errors"
	"fmt"
	"log"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	sdk "github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint/terraform"
	"github.com/terraform-linters/tflint/terraform/addrs"
	"github.com/terraform-linters/tflint/terraform/lang"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

// EvaluateExpr is a wrapper of terraform.BultinEvalContext.EvaluateExpr
// However, unlike the original implementation, it returns an error instead of
// returning the values below:
//
// - Unevaluable values (e.g. `module.<MODULE_NAME>`, `data.<DATA TYPE>.<NAME>`, `each.key`)
// - Unknown values
// - Null values
//
// An error is returned if these values are contained. This ensures that
// the caller can get the value via `gocty.FromCtyValue` unless an error occurs.
//
// As an exception, only unknown and null values can be returned as values by specifying
// `cty.DynamicPseudoType` for the type.
func (r *Runner) EvaluateExpr(expr hcl.Expression, wantType cty.Type) (cty.Value, error) {
	evaluable, err := isEvaluableExpr(expr)
	if err != nil {
		err := fmt.Errorf(
			"failed to parse an expression in %s:%d; %w",
			expr.Range().Filename,
			expr.Range().Start.Line,
			err,
		)
		log.Printf("[ERROR] %s", err)
		return cty.NullVal(cty.NilType), err
	}

	if !evaluable {
		err := fmt.Errorf(
			"unevaluable expression found in %s:%d%w",
			expr.Range().Filename,
			expr.Range().Start.Line,
			sdk.ErrUnevaluable,
		)
		log.Printf("[INFO] %s. TFLint ignores unevaluable expressions.", err)
		return cty.NullVal(cty.NilType), err
	}

	val, diags := r.ctx.EvaluateExpr(expr, wantType)
	if diags.HasErrors() {
		err := fmt.Errorf(
			"failed to eval an expression in %s:%d; %w",
			expr.Range().Filename,
			expr.Range().Start.Line,
			diags,
		)
		log.Printf("[ERROR] %s", err)
		return cty.NullVal(cty.NilType), err
	}

	if wantType == cty.DynamicPseudoType {
		return val, nil
	}

	err = cty.Walk(val, func(path cty.Path, v cty.Value) (bool, error) {
		if !v.IsKnown() {
			err := fmt.Errorf(
				"unknown value found in %s:%d%w",
				expr.Range().Filename,
				expr.Range().Start.Line,
				sdk.ErrUnknownValue,
			)
			log.Printf("[INFO] %s. TFLint can only evaluate provided variables and skips dynamic values.", err)
			return false, err
		}

		if v.IsNull() {
			err := fmt.Errorf(
				"null value found in %s:%d%w",
				expr.Range().Filename,
				expr.Range().Start.Line,
				sdk.ErrNullValue,
			)
			log.Printf("[INFO] %s. TFLint ignores expressions with null values.", err)
			return false, err
		}

		return true, nil
	})

	if err != nil {
		return cty.NullVal(cty.NilType), err
	}

	return val, nil
}

func isEvaluableExpr(expr hcl.Expression) (bool, error) {
	refs, diags := lang.ReferencesInExpr(expr)
	if diags.HasErrors() {
		return false, diags
	}
	for _, ref := range refs {
		if !isEvaluableRef(ref) {
			return false, nil
		}
	}
	return true, nil
}

func isEvaluableRef(ref *addrs.Reference) bool {
	switch ref.Subject.(type) {
	case addrs.InputVariable:
		return true
	case addrs.TerraformAttr:
		return true
	case addrs.PathAttr:
		return true
	default:
		return false
	}
}

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
	val, err := r.EvaluateExpr(expr, cty.Number)
	if err != nil {
		return isEvaluableMetaArgumentsOnError(err)
	}
	val, _ = val.Unmark()

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
	val, err := r.EvaluateExpr(expr, cty.DynamicPseudoType)
	if err != nil {
		return isEvaluableMetaArgumentsOnError(err)
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

func isEvaluableMetaArgumentsOnError(err error) (bool, error) {
	if errors.Is(err, sdk.ErrNullValue) {
		// null value means that attribute is not set
		return true, nil
	}
	if errors.Is(err, sdk.ErrUnknownValue) || errors.Is(err, sdk.ErrUnevaluable) {
		// unknown or unevaluable values are non-deterministic
		return false, nil
	}
	return false, err
}
