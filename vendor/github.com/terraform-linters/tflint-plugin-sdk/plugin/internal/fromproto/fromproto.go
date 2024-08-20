package fromproto

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/internal/proto"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/lang/marks"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/msgpack"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// BodySchema converts proto.BodySchema to hclext.BodySchema
func BodySchema(body *proto.BodySchema) *hclext.BodySchema {
	if body == nil {
		return nil
	}

	attributes := make([]hclext.AttributeSchema, len(body.Attributes))
	for idx, attr := range body.Attributes {
		attributes[idx] = hclext.AttributeSchema{Name: attr.Name, Required: attr.Required}
	}

	blocks := make([]hclext.BlockSchema, len(body.Blocks))
	for idx, block := range body.Blocks {
		blocks[idx] = hclext.BlockSchema{
			Type:       block.Type,
			LabelNames: block.LabelNames,
			Body:       BodySchema(block.Body),
		}
	}

	return &hclext.BodySchema{
		Mode:       SchemaMode(body.Mode),
		Attributes: attributes,
		Blocks:     blocks,
	}
}

// SchemaMode converts proto.SchemaMode to hclext.SchemaMode
func SchemaMode(mode proto.SchemaMode) hclext.SchemaMode {
	switch mode {
	case proto.SchemaMode_SCHEMA_MODE_UNSPECIFIED:
		return hclext.SchemaDefaultMode
	case proto.SchemaMode_SCHEMA_MODE_DEFAULT:
		return hclext.SchemaDefaultMode
	case proto.SchemaMode_SCHEMA_MODE_JUST_ATTRIBUTES:
		return hclext.SchemaJustAttributesMode
	default:
		panic(fmt.Sprintf("invalid SchemaMode: %s", mode))
	}
}

// BodyContent converts proto.BodyContent to hclext.BodyContent
func BodyContent(body *proto.BodyContent) (*hclext.BodyContent, hcl.Diagnostics) {
	if body == nil {
		return nil, nil
	}
	diags := hcl.Diagnostics{}

	attributes := hclext.Attributes{}
	for key, attr := range body.Attributes {
		expr, exprDiags := Expression(attr.Expression)
		diags = diags.Extend(exprDiags)

		attributes[key] = &hclext.Attribute{
			Name:      attr.Name,
			Expr:      expr,
			Range:     Range(attr.Range),
			NameRange: Range(attr.NameRange),
		}
	}

	blocks := make(hclext.Blocks, len(body.Blocks))
	for idx, block := range body.Blocks {
		blockBody, contentDiags := BodyContent(block.Body)
		diags = diags.Extend(contentDiags)

		labelRanges := make([]hcl.Range, len(block.LabelRanges))
		for idx, labelRange := range block.LabelRanges {
			labelRanges[idx] = Range(labelRange)
		}

		blocks[idx] = &hclext.Block{
			Type:        block.Type,
			Labels:      block.Labels,
			Body:        blockBody,
			DefRange:    Range(block.DefRange),
			TypeRange:   Range(block.TypeRange),
			LabelRanges: labelRanges,
		}
	}

	return &hclext.BodyContent{
		Attributes: attributes,
		Blocks:     blocks,
	}, diags
}

// RuleObject is an intermediate representation that satisfies the Rule interface.
type RuleObject struct {
	tflint.DefaultRule
	Data struct {
		Name     string
		Enabled  bool
		Severity tflint.Severity
		Link     string
	}
}

// Name returns the rule name
func (r *RuleObject) Name() string { return r.Data.Name }

// Enabled returns whether the rule is enabled
func (r *RuleObject) Enabled() bool { return r.Data.Enabled }

// Severity returns the severify of the rule
func (r *RuleObject) Severity() tflint.Severity { return r.Data.Severity }

// Link returns the link of the rule documentation if exists
func (r *RuleObject) Link() string { return r.Data.Link }

// Check does nothing. This is just a method to satisfy the interface
func (r *RuleObject) Check(tflint.Runner) error { return nil }

// Rule converts proto.EmitIssue_Rule to RuleObject
func Rule(rule *proto.EmitIssue_Rule) *RuleObject {
	if rule == nil {
		return nil
	}

	return &RuleObject{
		Data: struct {
			Name     string
			Enabled  bool
			Severity tflint.Severity
			Link     string
		}{
			Name:     rule.Name,
			Enabled:  rule.Enabled,
			Severity: Severity(rule.Severity),
			Link:     rule.Link,
		},
	}
}

// Expression converts proto.Expression to hcl.Expression
func Expression(expr *proto.Expression) (hcl.Expression, hcl.Diagnostics) {
	parsed, diags := hclext.ParseExpression(expr.Bytes, expr.Range.Filename, Pos(expr.Range.Start))
	if diags.HasErrors() {
		return nil, diags
	}
	if expr.Value != nil {
		val, err := Value(expr.Value, cty.DynamicPseudoType, expr.ValueMarks)
		if err != nil {
			panic(fmt.Errorf("cannot unmarshal the bound expr: %w", err))
		}
		parsed = hclext.BindValue(val, parsed)
	}
	return parsed, diags
}

// Severity converts proto.EmitIssue_Severity to severity
func Severity(severity proto.EmitIssue_Severity) tflint.Severity {
	switch severity {
	case proto.EmitIssue_SEVERITY_ERROR:
		return tflint.ERROR
	case proto.EmitIssue_SEVERITY_WARNING:
		return tflint.WARNING
	case proto.EmitIssue_SEVERITY_NOTICE:
		return tflint.NOTICE
	}

	return tflint.ERROR
}

// Range converts proto.Range to hcl.Range
func Range(rng *proto.Range) hcl.Range {
	if rng == nil {
		return hcl.Range{}
	}

	return hcl.Range{
		Filename: rng.Filename,
		Start:    Pos(rng.Start),
		End:      Pos(rng.End),
	}
}

// Pos converts proto.Range_Pos to hcl.Pos
func Pos(pos *proto.Range_Pos) hcl.Pos {
	if pos == nil {
		return hcl.Pos{}
	}

	return hcl.Pos{
		Line:   int(pos.Line),
		Column: int(pos.Column),
		Byte:   int(pos.Byte),
	}
}

// Value converts msgpack and []proto.ValueMark to cty.Value
func Value(value []byte, ty cty.Type, valueMarks []*proto.ValueMark) (cty.Value, error) {
	val, err := msgpack.Unmarshal(value, ty)
	if err != nil {
		return cty.NilVal, err
	}

	pvm := make([]cty.PathValueMarks, len(valueMarks))
	for idx, mark := range valueMarks {
		pvm[idx] = cty.PathValueMarks{
			Path: AttributePath(mark.Path),
		}
		if mark.Sensitive {
			pvm[idx].Marks = cty.NewValueMarks(marks.Sensitive)
		}
	}

	return val.MarkWithPaths(pvm), nil
}

// AttributePath converts proto.AttributePath to cty.Path
func AttributePath(path *proto.AttributePath) cty.Path {
	ret := cty.Path{}

	for _, step := range path.Steps {
		switch s := step.Selector.(type) {
		case *proto.AttributePath_Step_ElementKeyString:
			ret = ret.IndexString(s.ElementKeyString)
		case *proto.AttributePath_Step_ElementKeyInt:
			ret = ret.IndexInt(int(s.ElementKeyInt))
		case *proto.AttributePath_Step_AttributeName:
			ret = ret.GetAttr(s.AttributeName)
		}
	}
	return ret
}

// Config converts proto.ApplyGlobalConfig_Config to tflint.Config
func Config(config *proto.ApplyGlobalConfig_Config) *tflint.Config {
	if config == nil {
		return &tflint.Config{Rules: make(map[string]*tflint.RuleConfig)}
	}

	rules := map[string]*tflint.RuleConfig{}
	for name, rule := range config.Rules {
		rules[name] = &tflint.RuleConfig{Name: rule.Name, Enabled: rule.Enabled}
	}
	return &tflint.Config{
		Rules:             rules,
		DisabledByDefault: config.DisabledByDefault,
		Only:              config.Only,
		Fix:               config.Fix,
	}
}

// GetModuleContentOption converts proto.GetModuleContent_Option to tflint.GetModuleContentOption
func GetModuleContentOption(opts *proto.GetModuleContent_Option) tflint.GetModuleContentOption {
	if opts == nil {
		return tflint.GetModuleContentOption{}
	}

	return tflint.GetModuleContentOption{
		ModuleCtx:  ModuleCtxType(opts.ModuleCtx),
		ExpandMode: ExpandMode(opts.ExpandMode),
		Hint:       GetModuleContentHint(opts.Hint),
	}
}

// ModuleCtxType converts proto.ModuleCtxType to tflint.ModuleCtxType
func ModuleCtxType(ty proto.ModuleCtxType) tflint.ModuleCtxType {
	switch ty {
	case proto.ModuleCtxType_MODULE_CTX_TYPE_UNSPECIFIED:
		return tflint.SelfModuleCtxType
	case proto.ModuleCtxType_MODULE_CTX_TYPE_SELF:
		return tflint.SelfModuleCtxType
	case proto.ModuleCtxType_MODULE_CTX_TYPE_ROOT:
		return tflint.RootModuleCtxType
	default:
		panic(fmt.Sprintf("invalid ModuleCtxType: %s", ty))
	}
}

// ExpandMode converts proto.GetModuleContent_ExpandMode to tflint.ExpandMode
func ExpandMode(mode proto.GetModuleContent_ExpandMode) tflint.ExpandMode {
	switch mode {
	case proto.GetModuleContent_EXPAND_MODE_UNSPECIFIED:
		return tflint.ExpandModeExpand
	case proto.GetModuleContent_EXPAND_MODE_EXPAND:
		return tflint.ExpandModeExpand
	case proto.GetModuleContent_EXPAND_MODE_NONE:
		return tflint.ExpandModeNone
	default:
		panic(fmt.Sprintf("invalid ExpandMode: %s", mode))
	}
}

// GetModuleContentHint converts proto.GetModuleContent_Hint to tflint.GetModuleContentHint
func GetModuleContentHint(hint *proto.GetModuleContent_Hint) tflint.GetModuleContentHint {
	if hint == nil {
		return tflint.GetModuleContentHint{}
	}

	return tflint.GetModuleContentHint{
		ResourceType: hint.ResourceType,
	}
}

// Error converts gRPC error status to wrapped error
func Error(err error) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return err
	}

	// Unimplemented is an unexpected error, so return as-is.
	if st.Code() == codes.Unimplemented {
		return err
	}

	// If the error status has no details, return an error from the gRPC error status.
	// Remove the status code because some statuses are expected and should not be shown to users.
	if len(st.Details()) == 0 {
		return errors.New(st.Message())
	}

	// It is not supposed to have multiple details. The detail have an error code and will be wrapped as an error.
	switch t := st.Details()[0].(type) {
	case *proto.ErrorDetail:
		switch t.Code {
		case proto.ErrorCode_ERROR_CODE_UNKNOWN_VALUE:
			return tflint.ErrUnknownValue
		case proto.ErrorCode_ERROR_CODE_NULL_VALUE:
			return tflint.ErrNullValue
		case proto.ErrorCode_ERROR_CODE_UNEVALUABLE:
			return fmt.Errorf("%s%w", st.Message(), tflint.ErrUnevaluable)
		case proto.ErrorCode_ERROR_CODE_SENSITIVE:
			return tflint.ErrSensitive
		}
	}

	return err
}
