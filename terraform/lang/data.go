package lang

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/terraform/addrs"
	"github.com/zclconf/go-cty/cty"
)

// Data is an interface whose implementations can provide cty.Value
// representations of objects identified by referenceable addresses from
// the addrs package.
//
// This interface will grow each time a new type of reference is added, and so
// implementations outside of the Terraform codebases are not advised.
//
// Each method returns a suitable value and optionally some diagnostics. If the
// returned diagnostics contains errors then the type of the returned value is
// used to construct an unknown value of the same type which is then used in
// place of the requested object so that type checking can still proceed. In
// cases where it's not possible to even determine a suitable result type,
// cty.DynamicVal is returned along with errors describing the problem.
type Data interface {
	GetPathAttr(addrs.PathAttr, hcl.Range) (cty.Value, hcl.Diagnostics)
	GetTerraformAttr(addrs.TerraformAttr, hcl.Range) (cty.Value, hcl.Diagnostics)
	GetInputVariable(addrs.InputVariable, hcl.Range) (cty.Value, hcl.Diagnostics)
}
