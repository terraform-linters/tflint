// Package fromproto contains an implementation to decode a structure
// generated from *.proto into a real Go structure. This package is not
// intended to be used directly from plugins.
//
// Many primitives can be handled as-is, but some interfaces and errors
// require special decoding. The `hcl.Expression` restores the interface
// by reparsed based on the bytes and their range. The `tflint.Rule`
// restores the interface by filling the value in a pseudo-structure that
// satisfies the interface. Error makes use of gRPC error details to recover
// the wrapped error. Rewrap the error based on the error code obtained
// from details.
package fromproto
