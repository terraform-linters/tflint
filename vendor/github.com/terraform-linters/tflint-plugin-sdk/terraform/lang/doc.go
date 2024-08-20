// Package lang is a fork of Terraform's internal/lang package.
//
// This package provides helpers that interprets the Terraform Language's semantics
// in more detail than the HCL Language.
//
// For example, ReferencesInExpr returns a set of references, such as input variables
// and resources, rather than just a set of hcl.Traversal, filtering out invalid references.
package lang
