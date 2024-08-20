// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Package typeexpr extends HCL with a convention for describing HCL types
// within configuration files.
//
// The type syntax is processed statically from a hcl.Expression, so it cannot
// use any of the usual language operators. This is similar to type expressions
// in statically-typed programming languages.
//
//     variable "example" {
//       type = list(string)
//     }
package typeexpr
