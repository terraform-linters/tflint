// Package ctydebug contains some utilities for cty that are aimed at debugging
// and test code rather than at production code.
//
// A common characteristic of the functions here is that they are optimized
// for ease of use by having good defaults, as opposed to flexibility via
// lots of configuration arguments.
//
// Don't depend on the exact output of any functions in this package in tests,
// because the details may change in future versions in order to improve the
// output for human readers.
package ctydebug
