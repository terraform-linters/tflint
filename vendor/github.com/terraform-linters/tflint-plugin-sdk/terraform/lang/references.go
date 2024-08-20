package lang

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/addrs"
)

// References finds all of the references in the given set of traversals,
// returning diagnostics if any of the traversals cannot be interpreted as a
// reference.
//
// This function does not do any de-duplication of references, since references
// have source location information embedded in them and so any invalid
// references that are duplicated should have errors reported for each
// occurence.
//
// If the returned diagnostics contains errors then the result may be
// incomplete or invalid. Otherwise, the returned slice has one reference per
// given traversal, though it is not guaranteed that the references will
// appear in the same order as the given traversals.
func References(traversals []hcl.Traversal) ([]*addrs.Reference, hcl.Diagnostics) {
	if len(traversals) == 0 {
		return nil, nil
	}

	var diags hcl.Diagnostics
	refs := make([]*addrs.Reference, 0, len(traversals))

	for _, traversal := range traversals {
		ref, refDiags := addrs.ParseRef(traversal)
		diags = diags.Extend(refDiags)
		if ref == nil {
			continue
		}
		refs = append(refs, ref)
	}

	return refs, diags
}

// ReferencesInExpr is a helper wrapper around References that first searches
// the given expression for traversals, before converting those traversals
// to references.
//
// This function is almost identical to the Terraform internal API of the same name,
// except that it does not return diagnostics if it contains an invalid reference.
// This is because expressions with invalid traversals as references, such as
// `ignore_changes`, may be parsed. Developers should take advantage of the possible
// incomplete results returned by this function.
//
// Low-level APIs such as addrs.ParseRef are recommended if the expression is
// guaranteed not to contain invalid traversals, and analysis should stop in that case.
func ReferencesInExpr(expr hcl.Expression) []*addrs.Reference {
	if expr == nil {
		return nil
	}
	traversals := expr.Variables()
	refs, _ := References(traversals)
	return refs
}
