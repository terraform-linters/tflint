// Package toproto contains an implementation to encode a Go structure
// into a structure generated from *.proto. This package is not intended
// to be used directly from plugins.
//
// Many primitives can be handled as-is, but some interfaces and errors
// require special encoding. The `hcl.Expression` encodes into the range
// and the text representation as bytes. Error is encoded into gRPC error
// details to represent wrapped errors.
package toproto
