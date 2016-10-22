package aws

import (
	"github.com/hashicorp/hcl/hcl/ast"
	eval "github.com/wata727/tflint/evaluator"
)

type AwsDetector struct {
	ListMap    map[string]*ast.ObjectList
	EvalConfig *eval.Evaluator
}
