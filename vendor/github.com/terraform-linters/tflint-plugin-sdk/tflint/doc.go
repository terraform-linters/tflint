// Package tflint contains implementations and interfaces for
// plugin developers.
//
// Each rule can use the gRPC client that satisfies the Runner
// interface as an argument. Through this client, developers
// can get attributes, blocks, and resources to be analyzed
// and send issues to TFLint.
//
// All rules must be implemented to satisfy the Rule interface
// and a plugin must serve the RuleSet that bundles the rules.
package tflint
