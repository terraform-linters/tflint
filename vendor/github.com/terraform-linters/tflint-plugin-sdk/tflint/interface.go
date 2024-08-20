package tflint

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/addrs"
	"github.com/zclconf/go-cty/cty"
)

// RuleSet is a list of rules that a plugin should provide.
// Normally, plugins can use BuiltinRuleSet directly,
// but you can also use custom rulesets that satisfy this interface.
// The actual implementation can be found in plugin/host2plugin.GRPCServer.
type RuleSet interface {
	// RuleSetName is the name of the ruleset. This method is not expected to be overridden.
	RuleSetName() string

	// RuleSetVersion is the version of the plugin. This method is not expected to be overridden.
	RuleSetVersion() string

	// RuleNames is a list of rule names provided by the plugin. This method is not expected to be overridden.
	RuleNames() []string

	// VersionConstraint declares the version of TFLint the plugin will work with. Default is no constraint.
	VersionConstraint() string

	// ConfigSchema returns the ruleset plugin config schema.
	// If you return a schema, TFLint will extract the config from .tflint.hcl based on that schema
	// and pass it to ApplyConfig. This schema should be a schema inside of "plugin" block.
	// If you don't need a config that controls the entire plugin, you don't need to override this method.
	//
	// It is recommended to use hclext.ImpliedBodySchema to generate the schema from the structure:
	//
	// ```
	// type myPluginConfig struct {
	//   Style       string `hclext:"style"`
	//   Description string `hclext:"description,optional"`
	//   Detail      Detail `hclext:"detail,block"`
	// }
	//
	// config := &myPluginConfig{}
	// hclext.ImpliedBodySchema(config)
	// ```
	ConfigSchema() *hclext.BodySchema

	// ApplyGlobalConfig applies the common config to the ruleset.
	// This is not supposed to be overridden from custom rulesets.
	// Override the ApplyConfig if you want to apply the plugin's custom configuration.
	ApplyGlobalConfig(*Config) error

	// ApplyConfig applies the configuration to the ruleset.
	// Custom rulesets can override this method to reflect the plugin's custom configuration.
	//
	// You can reflect the body in the structure by using hclext.DecodeBody:
	//
	// ```
	// type myPluginConfig struct {
	//   Style       string `hclext:"style"`
	//   Description string `hclext:"description,optional"`
	//   Detail      Detail `hclext:"detail,block"`
	// }
	//
	// config := &myPluginConfig{}
	// hclext.DecodeBody(body, nil, config)
	// ```
	ApplyConfig(*hclext.BodyContent) error

	// NewRunner returns a new runner based on the original runner.
	// Custom rulesets can override this method to inject a custom runner.
	NewRunner(Runner) (Runner, error)

	// BuiltinImpl returns the receiver itself as BuiltinRuleSet.
	// This is not supposed to be overridden from custom rulesets.
	BuiltinImpl() *BuiltinRuleSet

	// All Ruleset must embed the builtin ruleset.
	mustEmbedBuiltinRuleSet()
}

// Runner acts as a client for each plugin to query the host process about the Terraform configurations.
// The actual implementation can be found in plugin/plugin2host.GRPCClient.
type Runner interface {
	// GetOriginalwd returns the original working directory.
	// Normally this is equal to os.Getwd(), but differs if --chdir or --recursive is used.
	// If you need the absolute path of the file, joining with the original working directory is appropriate.
	GetOriginalwd() (string, error)

	// GetModulePath returns the current module path address.
	GetModulePath() (addrs.Module, error)

	// GetResourceContent retrieves the content of resources based on the passed schema.
	// The schema allows you to specify attributes and blocks that describe the structure needed for the inspection:
	//
	// ```
	// runner.GetResourceContent("aws_instance", &hclext.BodySchema{
	//   Attributes: []hclext.AttributeSchema{{Name: "instance_type"}},
	//   Blocks: []hclext.BlockSchema{
	//     {
	//       Type: "ebs_block_device",
	//       Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "volume_size"}}},
	//     },
	//   },
	// }, nil)
	// ```
	GetResourceContent(resourceName string, schema *hclext.BodySchema, option *GetModuleContentOption) (*hclext.BodyContent, error)

	// GetProviderContent retrieves the content of providers based on the passed schema.
	// This method is GetResourceContent for providers.
	GetProviderContent(providerName string, schema *hclext.BodySchema, option *GetModuleContentOption) (*hclext.BodyContent, error)

	// GetModuleContent retrieves the content of the module based on the passed schema.
	// GetResourceContent/GetProviderContent are syntactic sugar for GetModuleContent, which you can use to access other structures.
	GetModuleContent(schema *hclext.BodySchema, option *GetModuleContentOption) (*hclext.BodyContent, error)

	// GetFile returns the hcl.File object.
	// This is low level API for accessing information such as comments and syntax.
	// When accessing resources, expressions, etc, it is recommended to use high-level APIs.
	GetFile(filename string) (*hcl.File, error)

	// GetFiles returns a map[string]hcl.File object, where the key is the file name.
	// This is low level API for accessing information such as comments and syntax.
	GetFiles() (map[string]*hcl.File, error)

	// WalkExpressions traverses expressions in all files by the passed walker.
	// The walker can be passed any structure that satisfies the `tflint.ExprWalker`
	// interface, or a `tflint.ExprWalkFunc`. Example of passing function:
	//
	// ```
	// runner.WalkExpressions(tflint.ExprWalkFunc(func (expr hcl.Expression) hcl.Diagnostics {
	//   // Write code here
	// }))
	// ```
	//
	// If you pass ExprWalkFunc, the function will be called for every expression.
	// Note that it behaves differently in native HCL syntax and JSON syntax.
	//
	// In the HCL syntax, `var.foo` and `var.bar` in `[var.foo, var.bar]` are
	// also passed to the walker. In other words, it traverses expressions recursively.
	// To avoid redundant checks, the walker should check the kind of expression.
	//
	// In the JSON syntax, only an expression of an attribute seen from the top
	// level of the file is passed. In other words, it doesn't traverse expressions
	// recursively. This is a limitation of JSON syntax.
	WalkExpressions(walker ExprWalker) hcl.Diagnostics

	// DecodeRuleConfig fetches the rule's configuration and reflects the result in the 2nd argument.
	// The argument is expected to be a pointer to a structure tagged with hclext:
	//
	// ```
	// type myRuleConfig struct {
	//   Style       string `hclext:"style"`
	//   Description string `hclext:"description,optional"`
	//   Detail      Detail `hclext:"detail,block"`
	// }
	//
	// config := &myRuleConfig{}
	// runner.DecodeRuleConfig("my_rule", config)
	// ```
	//
	// See the hclext.DecodeBody documentation and examples for more details.
	DecodeRuleConfig(ruleName string, ret interface{}) error

	// EvaluateExpr evaluates an expression and assigns its value to a Go value target,
	// which must be a pointer or a function. Any other type of target will trigger a panic.
	//
	// For pointers, if the expression value cannot be assigned to the target, an error is returned.
	// Some examples of this include unknown values (like variables without defaults or
	// aws_instance.foo.arn), null values, and sensitive values (for variables with sensitive = true).
	//
	// These errors be handled with errors.Is():
	//
	// ```
	// var val string
	// err := runner.EvaluateExpr(expr, &val, nil)
	// if err != nil {
	//   if errors.Is(err, tflint.ErrUnknownValue) {
	//     // Ignore unknown values
	//	   return nil
	//   }
	//   if errors.Is(err, tflint.ErrNullValue) {
	//     // Ignore null values because null means that the value is not set
	//	   return nil
	//   }
	//   if errors.Is(err, tflint.ErrSensitive) {
	//     // Ignore sensitive values
	//     return nil
	//   }
	//   return err
	// }
	// ```
	//
	// However, if the target is cty.Value, these errors will not be returned.
	//
	// Here are the types that can be passed as the target: string, int, bool, []string,
	// []int, []bool, map[string]string, map[string]int, map[string]bool, and cty.Value.
	// Passing any other type will result in a panic, but you can make an exception by
	// passing wantType as an option.
	//
	// ```
	// type complexVal struct {
	//   Key     string `cty:"key"`
	//   Enabled bool   `cty:"enabled"`
	// }
	//
	// wantType := cty.List(cty.Object(map[string]cty.Type{
	//   "key":     cty.String,
	//   "enabled": cty.Bool,
	// }))
	//
	// var complexVals []complexVal
	// runner.EvaluateExpr(expr, &compleVals, &tflint.EvaluateExprOption{WantType: &wantType})
	// ```
	//
	// For functions (callbacks), the assigned value is used as an argument to execute
	// the function. If a value cannot be assigned to the argument type, the execution
	// is skipped instead of returning an error. This is useful when it's always acceptable
	// to ignore exceptional values.
	//
	// Here's an example of how you can pass a function to EvaluateExpr:
	//
	// ```
	// runner.EvaluateExpr(expr, func (val string) error {
	//   // Test value
	// }, nil)
	// ```
	EvaluateExpr(expr hcl.Expression, target interface{}, option *EvaluateExprOption) error

	// EmitIssue sends an issue to TFLint. You need to pass the message of the issue and the range.
	EmitIssue(rule Rule, message string, issueRange hcl.Range) error

	// EmitIssueWithFix is similar to EmitIssue, but it also supports autofix.
	// If you pass a function that rewrites the source code to the last argument,
	// TFLint will apply the fix when the --fix option is specified.
	//
	// The function is passed a tflint.Fixer that can be used to rewrite the source code.
	// See the tflint.Fixer interface for more details.
	//
	// Issues emitted using this function are automatically marked as fixable.
	// However, if you don't want to support autofix only under certain conditions (e.g. JSON syntax),
	// you can return tflint.ErrFixNotSupported from the fix function.
	// In this case, the issue will not be marked as fixable and the fix will not be applied.
	//
	// As a best practice for autofix, we recommend minimizing the amount of code changed at once.
	// If fixes for the same range conflict within the same rule, Fixer will return an error.
	EmitIssueWithFix(rule Rule, message string, issueRange hcl.Range, fixFunc func(f Fixer) error) error

	// EnsureNoError is a helper for error handling. Depending on the type of error generated by EvaluateExpr,
	// determine whether to exit, skip, or continue. If it is continued, the passed function will be executed.
	//
	// Deprecated: Use EvaluateExpr with a function callback. e.g. EvaluateExpr(expr, func (val T) error {}, ...)
	EnsureNoError(error, func() error) error
}

// Fixer is a tool to rewrite HCL source code.
// The actual implementation is in the internal.Fixer.
type Fixer interface {
	// ReplaceText rewrites the given range of source code to a new text.
	// If the range is overlapped with a previous rewrite range, it returns an error.
	//
	// Either string or tflint.TextNode is valid as an argument.
	// TextNode can be obtained with fixer.TextAt(range).
	// If the argument is a TextNode, and the range is contained in the replacement range,
	// this function automatically minimizes the replacement range as much as possible.
	//
	// For example, if the source code is "(foo)", ReplaceText(range, "[foo]")
	// rewrites the whole "(foo)". But ReplaceText(range, "[", TextAt(fooRange), "]")
	// rewrites only "(" and ")". This is useful to avoid unintended conflicts.
	ReplaceText(hcl.Range, ...any) error

	// InsertTextBefore inserts the given text before the given range.
	InsertTextBefore(hcl.Range, string) error

	// InsertTextAfter inserts the given text after the given range.
	InsertTextAfter(hcl.Range, string) error

	// Remove removes the given range of source code.
	Remove(hcl.Range) error

	// RemoveAttribute removes the given attribute from the source code.
	// The difference from Remove is that it removes the attribute
	// and the associated newlines, indentations, and comments.
	// This only works for HCL native syntax. JSON syntax is not supported
	// and returns tflint.ErrFixNotSupported.
	RemoveAttribute(*hcl.Attribute) error

	// RemoveBlock removes the given block from the source code.
	// The difference from Remove is that it removes the block
	// and the associated newlines, indentations, and comments.
	// This only works for HCL native syntax. JSON syntax is not supported
	// and returns tflint.ErrFixNotSupported.
	RemoveBlock(*hcl.Block) error

	// RemoveExtBlock removes the given block from the source code.
	// This is similar to RemoveBlock, but it works for hclext.Block.
	RemoveExtBlock(*hclext.Block) error

	// TextAt returns a text node at the given range.
	// This is expected to be passed as an argument to ReplaceText.
	// Note this doesn't take into account the changes made by the fixer in a rule.
	TextAt(hcl.Range) TextNode

	// ValueText returns a text representation of the given cty.Value.
	// Values are always converted to a single line. For more pretty-printing,
	// implement your own conversion function.
	//
	// This function is inspired by hclwrite.TokensForValue.
	// https://github.com/hashicorp/hcl/blob/v2.16.2/hclwrite/generate.go#L26
	ValueText(cty.Value) string

	// RangeTo returns a range from the given start position to the given text.
	// Note that it doesn't check if the text is actually in the range.
	RangeTo(to string, filename string, start hcl.Pos) hcl.Range
}

// Rule is the interface that the plugin's rules should satisfy.
type Rule interface {
	// Name will be displayed with a message of an issue and will be the identifier used to control
	// the behavior of this rule in the configuration file etc.
	// Therefore, it is expected that this will not duplicate the rule names provided by other plugins.
	Name() string

	// Enabled indicates whether the rule is enabled by default.
	Enabled() bool

	// Severity indicates the severity of the rule.
	Severity() Severity

	// Link allows you to add a reference link to the rule.
	Link() string

	// Metadata allows you to set any metadata to the rule.
	// This value is never referenced by the SDK and can be used for your custom ruleset.
	Metadata() interface{}

	// Check is the entrypoint of the rule. You can fetch Terraform configurations and send issues via Runner.
	Check(Runner) error

	// All rules must embed the default rule.
	mustEmbedDefaultRule()
}
