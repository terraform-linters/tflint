package tflint

import (
	"errors"
	"fmt"
	"log"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	sdk "github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint/terraform/addrs"
	"github.com/terraform-linters/tflint/terraform/lang"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

// EvaluateExpr is a wrapper of terraform.BultinEvalContext.EvaluateExpr
// In addition, it returns an error if expr cannot be evaluated, if it contains an unknown value,
// or if it contains null. However, it allows null and unknown only for DynamicPseudoType.
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

	val, diags := r.ctx.EvaluateExpr(expr, wantType, nil)
	if diags.HasErrors() {
		err := fmt.Errorf(
			"failed to eval an expression in %s:%d; %w",
			expr.Range().Filename,
			expr.Range().Start.Line,
			diags.Err(),
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
		return false, diags.Err()
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

// willEvaluateResource checks whether the passed resource will be evaluated.
// If `count` is 0 or `for_each` is empty, Terraform will not evaluate the attributes of that resource.
// Terraform doesn't expect to pass null for these attributes (it will cause an error),
// but we'll treat them as if they were undefined.
func (r *Runner) willEvaluateResource(resource *hclext.Block) (bool, error) {
	if attr, exists := resource.Body.Attributes["count"]; exists {
		val, err := r.EvaluateExpr(attr.Expr, cty.Number)
		if err != nil {
			return willEvaluateResourceOnError(err)
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

	if attr, exists := resource.Body.Attributes["for_each"]; exists {
		forEach, err := r.EvaluateExpr(attr.Expr, cty.DynamicPseudoType)
		if err != nil {
			return willEvaluateResourceOnError(err)
		}

		if forEach.IsNull() {
			// null value means that attribute is not set
			return true, nil
		}
		if !forEach.IsKnown() {
			// unknown value is non-deterministic
			return false, nil
		}
		if !forEach.CanIterateElements() {
			// uniteratable values (string, number, etc.) are invalid
			return false, fmt.Errorf("The `for_each` value is not iterable in %s:%d", attr.Expr.Range().Filename, attr.Expr.Range().Start.Line)
		}
		if forEach.LengthInt() == 0 {
			// empty `for_each` is not evaluated
			return false, nil
		}
		// `for_each` with non-empty elements is evaluated
		return true, nil
	}

	// If `count` or `for_each` is not defined, it will be evaluated by default
	return true, nil
}

func willEvaluateResourceOnError(err error) (bool, error) {
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
