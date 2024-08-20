// Package hclext is an extension of package hcl for TFLint.
//
// The goal of this package is to work with nested hcl.BodyContent.
// In the various functions provided by the package hcl, hcl.Block
// nests hcl.Body as body. However, since hcl.Body is an interface,
// the nested body cannot be sent over a wire protocol.
//
// In this package, redefine hcl.Block as hclext.Block nests BodyContent,
// not Body, which is an interface. Some functions and related structures
// have been redefined to make hclext.Block behave like the package hcl.
//
// For example, Content/PartialContent takes hclext.BodySchema instead of
// hcl.BodySchema and returns hclext.BodyContent. In hclext.BodySchema,
// you can declare the structure of the nested body as the block schema.
// This allows you to send the schema and its results of configurations
// that contain nested bodies via gRPC.
package hclext
