package aws

import (
	"github.com/hashicorp/hcl/hcl/ast"
	eval "github.com/wata727/tflint/evaluator"
)

type AwsDetector struct {
	List       *ast.ObjectList
	File       string
	EvalConfig *eval.Evaluator
}
