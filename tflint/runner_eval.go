package tflint

import (
	"errors"
	"fmt"
	"log"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/terraform/addrs"
	"github.com/terraform-linters/tflint/terraform/configs"
	"github.com/terraform-linters/tflint/terraform/configs/configschema"
	"github.com/terraform-linters/tflint/terraform/lang"
	"github.com/terraform-linters/tflint/terraform/terraform"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/gocty"
)

// EvaluateExpr evaluates the expression and reflects the result in the value of `ret`.
func (r *Runner) EvaluateExpr(expr hcl.Expression, ret interface{}) error {
	val, err := r.EvalExpr(expr, ret, cty.Type{})
	if err != nil {
		return err
	}

	err = gocty.FromCtyValue(val, ret)
	if err != nil {
		err := &Error{
			Code:  TypeMismatchError,
			Level: ErrorLevel,
			Message: fmt.Sprintf(
				"Invalid type expression in %s:%d",
				expr.Range().Filename,
				expr.Range().Start.Line,
			),
			Cause: err,
		}
		log.Printf("[ERROR] %s", err)
		return err
	}
	return nil
}

// EvalExpr is a wrapper of terraform.BultinEvalContext.EvaluateExpr
// In addition, this method determines whether the expression is evaluable, contains no unknown values, and so on.
// The returned cty.Value is converted according to the value passed as `ret`.
func (r *Runner) EvalExpr(expr hcl.Expression, ret interface{}, wantType cty.Type) (cty.Value, error) {
	evaluable, err := isEvaluableExpr(expr)
	if err != nil {
		err := &Error{
			Code:  EvaluationError,
			Level: ErrorLevel,
			Message: fmt.Sprintf(
				"Failed to parse an expression in %s:%d",
				expr.Range().Filename,
				expr.Range().Start.Line,
			),
			Cause: err,
		}
		log.Printf("[ERROR] %s", err)
		return cty.NullVal(cty.NilType), err
	}

	if !evaluable {
		err := &Error{
			Code:  UnevaluableError,
			Level: WarningLevel,
			Message: fmt.Sprintf(
				"Unevaluable expression found in %s:%d",
				expr.Range().Filename,
				expr.Range().Start.Line,
			),
		}
		log.Printf("[INFO] %s. TFLint ignores unevaluable expressions.", err)
		return cty.NullVal(cty.NilType), err
	}

	if wantType == (cty.Type{}) {
		switch ret.(type) {
		case *string, string:
			wantType = cty.String
		case *int, int:
			wantType = cty.Number
		case *[]string, []string:
			wantType = cty.List(cty.String)
		case *[]int, []int:
			wantType = cty.List(cty.Number)
		case *map[string]string, map[string]string:
			wantType = cty.Map(cty.String)
		case *map[string]int, map[string]int:
			wantType = cty.Map(cty.Number)
		default:
			panic(fmt.Errorf("Unexpected result type: %T", ret))
		}
	}

	val, diags := r.ctx.EvaluateExpr(expr, wantType, nil)
	if diags.HasErrors() {
		err := &Error{
			Code:  EvaluationError,
			Level: ErrorLevel,
			Message: fmt.Sprintf(
				"Failed to eval an expression in %s:%d",
				expr.Range().Filename,
				expr.Range().Start.Line,
			),
			Cause: diags.Err(),
		}
		log.Printf("[ERROR] %s", err)
		return cty.NullVal(cty.NilType), err
	}

	err = cty.Walk(val, func(path cty.Path, v cty.Value) (bool, error) {
		if !v.IsKnown() {
			err := &Error{
				Code:  UnknownValueError,
				Level: WarningLevel,
				Message: fmt.Sprintf(
					"Unknown value found in %s:%d",
					expr.Range().Filename,
					expr.Range().Start.Line,
				),
			}
			log.Printf("[INFO] %s. TFLint can only evaluate provided variables and skips dynamic values.", err)
			return false, err
		}

		if v.IsNull() {
			err := &Error{
				Code:  NullValueError,
				Level: WarningLevel,
				Message: fmt.Sprintf(
					"Null value found in %s:%d",
					expr.Range().Filename,
					expr.Range().Start.Line,
				),
			}
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

// EvaluateBlock is a wrapper of terraform.BultinEvalContext.EvaluateBlock and gocty.FromCtyValue
func (r *Runner) EvaluateBlock(block *hcl.Block, schema *configschema.Block, ret interface{}) error {
	evaluable, err := isEvaluableBlock(block.Body, schema)
	if err != nil {
		err := &Error{
			Code:  EvaluationError,
			Level: ErrorLevel,
			Message: fmt.Sprintf(
				"Failed to parse a block in %s:%d",
				block.DefRange.Filename,
				block.DefRange.Start.Line,
			),
			Cause: err,
		}
		log.Printf("[ERROR] %s", err)
		return err
	}

	if !evaluable {
		err := &Error{
			Code:  UnevaluableError,
			Level: WarningLevel,
			Message: fmt.Sprintf(
				"Unevaluable block found in %s:%d",
				block.DefRange.Filename,
				block.DefRange.Start.Line,
			),
		}
		log.Printf("[INFO] %s. TFLint ignores unevaluable blocks.", err)
		return err
	}

	val, _, diags := r.ctx.EvaluateBlock(block.Body, schema, nil, terraform.EvalDataForNoInstanceKey)
	if diags.HasErrors() {
		err := &Error{
			Code:  EvaluationError,
			Level: ErrorLevel,
			Message: fmt.Sprintf(
				"Failed to eval a block in %s:%d",
				block.DefRange.Filename,
				block.DefRange.Start.Line,
			),
			Cause: diags.Err(),
		}
		log.Printf("[ERROR] %s", err)
		return err
	}

	err = cty.Walk(val, func(path cty.Path, v cty.Value) (bool, error) {
		if !v.IsKnown() {
			err := &Error{
				Code:  UnknownValueError,
				Level: WarningLevel,
				Message: fmt.Sprintf(
					"Unknown value found in %s:%d",
					block.DefRange.Filename,
					block.DefRange.Start.Line,
				),
			}
			log.Printf("[INFO] %s. TFLint can only evaluate provided variables and skips blocks with unknown values.", err)
			return false, err
		}

		return true, nil
	})
	if err != nil {
		return err
	}

	val, err = cty.Transform(val, func(path cty.Path, v cty.Value) (cty.Value, error) {
		if v.IsNull() {
			log.Printf(
				"[DEBUG] Null value found in %s:%d. TFLint treats this value as an empty value.",
				block.DefRange.Filename,
				block.DefRange.Start.Line,
			)
			return cty.StringVal(""), nil
		}
		return v, nil
	})
	if err != nil {
		return err
	}

	switch ret.(type) {
	case *map[string]string:
		val, err = convert.Convert(val, cty.Map(cty.String))
	case *map[string]int:
		val, err = convert.Convert(val, cty.Map(cty.Number))
	}

	if err != nil {
		err := &Error{
			Code:  TypeConversionError,
			Level: ErrorLevel,
			Message: fmt.Sprintf(
				"Invalid type block in %s:%d",
				block.DefRange.Filename,
				block.DefRange.Start.Line,
			),
			Cause: err,
		}
		log.Printf("[ERROR] %s", err)
		return err
	}

	err = gocty.FromCtyValue(val, ret)
	if err != nil {
		err := &Error{
			Code:  TypeMismatchError,
			Level: ErrorLevel,
			Message: fmt.Sprintf(
				"Invalid type block in %s:%d",
				block.DefRange.Filename,
				block.DefRange.Start.Line,
			),
			Cause: err,
		}
		log.Printf("[ERROR] %s", err)
		return err
	}
	return nil
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

func isEvaluableBlock(body hcl.Body, schema *configschema.Block) (bool, error) {
	refs, diags := lang.ReferencesInBlock(body, schema)
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
func (r *Runner) willEvaluateResource(resource *configs.Resource) (bool, error) {
	var err error
	if resource.Count != nil {
		count := 1
		err = r.EvaluateExpr(resource.Count, &count)
		if err == nil && count == 0 {
			return false, nil
		}
	} else if resource.ForEach != nil {
		var forEach cty.Value
		forEach, err = r.EvalExpr(resource.ForEach, nil, cty.DynamicPseudoType)
		if err == nil {
			if !forEach.CanIterateElements() {
				return false, fmt.Errorf("The `for_each` value is not iterable in %s:%d", resource.ForEach.Range().Filename, resource.ForEach.Range().Start.Line)
			}
			if forEach.LengthInt() == 0 {
				return false, nil
			}
		}
	}

	if err == nil {
		return true, nil
	}
	var appErr *Error
	if errors.As(err, &appErr) {
		switch appErr.Level {
		case WarningLevel:
			return false, nil
		case ErrorLevel:
			return false, err
		default:
			panic(appErr)
		}
	} else {
		return false, err
	}
}
