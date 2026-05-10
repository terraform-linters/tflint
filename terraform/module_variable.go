package terraform

import (
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/terraform/addrs"
	"github.com/terraform-linters/tflint/terraform/lang"
)

type moduleVariable struct {
	Root      bool
	Parents   []*moduleVariable
	Callers   []*moduleVariable
	DeclRange hcl.Range
}

func (m *moduleVariable) roots() []*moduleVariable {
	if m.Root {
		return []*moduleVariable{m}
	}

	ret := []*moduleVariable{}
	for _, parent := range m.Parents {
		for _, parentRoot := range parent.roots() {
			parentRoot.Callers = append(parentRoot.Callers, m)
			ret = append(ret, parentRoot)
		}
	}
	return ret
}

func (m *moduleVariable) callers() []hcl.Range {
	ret := make([]hcl.Range, len(m.Callers)+1)
	ret[0] = m.DeclRange

	for idx, caller := range m.Callers {
		ret[idx+1] = caller.DeclRange
	}
	return ret
}

// listVarRefs returns the references in the expression.
// If the expression is not a valid expression, it returns an empty map.
func listVarRefs(expr hcl.Expression) map[string]addrs.InputVariable {
	ret := map[string]addrs.InputVariable{}
	refs, diags := lang.ReferencesInExpr(expr)

	if diags.HasErrors() {
		// If we cannot determine the references in the expression, it is likely a valid HCL expression, but not a valid Terraform expression.
		// The declaration range of a block with no labels is its name, which is syntactically valid as an HCL expression, but is not a valid Terraform reference.
		return ret
	}

	for _, ref := range refs {
		if varRef, ok := ref.Subject.(addrs.InputVariable); ok {
			ret[varRef.String()] = varRef
		}
	}

	return ret
}
