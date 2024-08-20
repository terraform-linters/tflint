package toproto

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin/internal/proto"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/lang/marks"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/msgpack"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// BodySchema converts schema.BodySchema to proto.BodySchema
func BodySchema(body *hclext.BodySchema) *proto.BodySchema {
	if body == nil {
		return &proto.BodySchema{}
	}

	attributes := make([]*proto.BodySchema_Attribute, len(body.Attributes))
	for idx, attr := range body.Attributes {
		attributes[idx] = &proto.BodySchema_Attribute{Name: attr.Name, Required: attr.Required}
	}

	blocks := make([]*proto.BodySchema_Block, len(body.Blocks))
	for idx, block := range body.Blocks {
		blocks[idx] = &proto.BodySchema_Block{
			Type:       block.Type,
			LabelNames: block.LabelNames,
			Body:       BodySchema(block.Body),
		}
	}

	return &proto.BodySchema{
		Mode:       SchemaMode(body.Mode),
		Attributes: attributes,
		Blocks:     blocks,
	}
}

// SchemaMode converts hclext.SchemaMode to proto.SchemaMode
func SchemaMode(mode hclext.SchemaMode) proto.SchemaMode {
	switch mode {
	case hclext.SchemaDefaultMode:
		return proto.SchemaMode_SCHEMA_MODE_DEFAULT
	case hclext.SchemaJustAttributesMode:
		return proto.SchemaMode_SCHEMA_MODE_JUST_ATTRIBUTES
	default:
		panic(fmt.Sprintf("invalid SchemaMode: %s", mode))
	}
}

// BodyContent converts schema.BodyContent to proto.BodyContent
func BodyContent(body *hclext.BodyContent, sources map[string][]byte) *proto.BodyContent {
	if body == nil {
		return &proto.BodyContent{}
	}

	attributes := map[string]*proto.BodyContent_Attribute{}
	for idx, attr := range body.Attributes {
		bytes, ok := sources[attr.Range.Filename]
		if !ok {
			panic(fmt.Sprintf("failed to encode to protocol buffers: source code not available: name=%s", attr.Range.Filename))
		}

		attributes[idx] = &proto.BodyContent_Attribute{
			Name:       attr.Name,
			Expression: Expression(attr.Expr, bytes),
			Range:      Range(attr.Range),
			NameRange:  Range(attr.NameRange),
		}
	}

	blocks := make([]*proto.BodyContent_Block, len(body.Blocks))
	for idx, block := range body.Blocks {
		labelRanges := make([]*proto.Range, len(block.LabelRanges))
		for idx, labelRange := range block.LabelRanges {
			labelRanges[idx] = Range(labelRange)
		}

		blocks[idx] = &proto.BodyContent_Block{
			Type:        block.Type,
			Labels:      block.Labels,
			Body:        BodyContent(block.Body, sources),
			DefRange:    Range(block.DefRange),
			TypeRange:   Range(block.TypeRange),
			LabelRanges: labelRanges,
		}
	}

	return &proto.BodyContent{
		Attributes: attributes,
		Blocks:     blocks,
	}
}

// Rule converts tflint.Rule to proto.EmitIssue_Rule
func Rule(rule tflint.Rule) *proto.EmitIssue_Rule {
	if rule == nil {
		panic("failed to encode to protocol buffers: rule should not be nil")
	}
	return &proto.EmitIssue_Rule{
		Name:     rule.Name(),
		Enabled:  rule.Enabled(),
		Severity: Severity(rule.Severity()),
		Link:     rule.Link(),
	}
}

// Expression converts hcl.Expression to proto.Expression
func Expression(expr hcl.Expression, source []byte) *proto.Expression {
	out := &proto.Expression{
		Bytes: expr.Range().SliceBytes(source),
		Range: Range(expr.Range()),
	}

	if boundExpr, ok := expr.(*hclext.BoundExpr); ok {
		val, marks, err := Value(boundExpr.Val, cty.DynamicPseudoType)
		if err != nil {
			panic(fmt.Errorf("cannot marshal the bound expr: %w", err))
		}
		out.Value = val
		out.ValueMarks = marks
	}
	return out
}

// Severity converts severity to proto.EmitIssue_Severity
func Severity(severity tflint.Severity) proto.EmitIssue_Severity {
	switch severity {
	case tflint.ERROR:
		return proto.EmitIssue_SEVERITY_ERROR
	case tflint.WARNING:
		return proto.EmitIssue_SEVERITY_WARNING
	case tflint.NOTICE:
		return proto.EmitIssue_SEVERITY_NOTICE
	}

	return proto.EmitIssue_SEVERITY_ERROR
}

// Range converts hcl.Range to proto.Range
func Range(rng hcl.Range) *proto.Range {
	return &proto.Range{
		Filename: rng.Filename,
		Start:    Pos(rng.Start),
		End:      Pos(rng.End),
	}
}

// Pos converts hcl.Pos to proto.Range_Pos
func Pos(pos hcl.Pos) *proto.Range_Pos {
	return &proto.Range_Pos{
		Line:   int64(pos.Line),
		Column: int64(pos.Column),
		Byte:   int64(pos.Byte),
	}
}

// Value converts cty.Value to msgpack and serialized value marks
func Value(value cty.Value, ty cty.Type) ([]byte, []*proto.ValueMark, error) {
	// Convert first to get the actual cty.Path
	value, err := convert.Convert(value, ty)
	if err != nil {
		return nil, nil, err
	}

	value, pvm := value.UnmarkDeepWithPaths()
	valueMarks := make([]*proto.ValueMark, len(pvm))
	for idx, m := range pvm {
		path, err := AttributePath(m.Path)
		if err != nil {
			return nil, nil, err
		}

		valueMarks[idx] = &proto.ValueMark{Path: path}
		if _, exists := m.Marks[marks.Sensitive]; exists {
			valueMarks[idx].Sensitive = true
		}
	}

	val, err := msgpack.Marshal(value, ty)
	if err != nil {
		return nil, nil, err
	}

	return val, valueMarks, nil
}

// AttributePath converts cty.Path to proto.AttributePath
func AttributePath(path cty.Path) (*proto.AttributePath, error) {
	steps := make([]*proto.AttributePath_Step, len(path))

	for idx, step := range path {
		switch s := step.(type) {
		case cty.IndexStep:
			switch s.Key.Type() {
			case cty.String:
				steps[idx] = &proto.AttributePath_Step{
					Selector: &proto.AttributePath_Step_ElementKeyString{ElementKeyString: s.Key.AsString()},
				}
			case cty.Number:
				v, _ := s.Key.AsBigFloat().Int64()
				steps[idx] = &proto.AttributePath_Step{
					Selector: &proto.AttributePath_Step_ElementKeyInt{ElementKeyInt: v},
				}
			default:
				return nil, fmt.Errorf("unknown index step key type: %s", s.Key.Type().GoString())
			}
		case cty.GetAttrStep:
			steps[idx] = &proto.AttributePath_Step{
				Selector: &proto.AttributePath_Step_AttributeName{AttributeName: s.Name},
			}
		default:
			return nil, fmt.Errorf("unknown attribute path step: %T", s)
		}
	}

	return &proto.AttributePath{Steps: steps}, nil
}

// Config converts tflint.Config to proto.ApplyGlobalConfig_Config
func Config(config *tflint.Config) *proto.ApplyGlobalConfig_Config {
	if config == nil {
		return &proto.ApplyGlobalConfig_Config{Rules: make(map[string]*proto.ApplyGlobalConfig_RuleConfig)}
	}

	rules := map[string]*proto.ApplyGlobalConfig_RuleConfig{}
	for name, rule := range config.Rules {
		rules[name] = &proto.ApplyGlobalConfig_RuleConfig{Name: rule.Name, Enabled: rule.Enabled}
	}
	return &proto.ApplyGlobalConfig_Config{
		Rules:             rules,
		DisabledByDefault: config.DisabledByDefault,
		Only:              config.Only,
		Fix:               config.Fix,
	}
}

// GetModuleContentOption converts tflint.GetModuleContentOption to proto.GetModuleContent_Option
func GetModuleContentOption(opts *tflint.GetModuleContentOption) *proto.GetModuleContent_Option {
	if opts == nil {
		return &proto.GetModuleContent_Option{}
	}

	return &proto.GetModuleContent_Option{
		ModuleCtx:  ModuleCtxType(opts.ModuleCtx),
		ExpandMode: ExpandMode(opts.ExpandMode),
		Hint:       GetModuleContentHint(opts.Hint),
	}
}

// ModuleCtxType converts tflint.ModuleCtxType to proto.ModuleCtxType
func ModuleCtxType(ty tflint.ModuleCtxType) proto.ModuleCtxType {
	switch ty {
	case tflint.SelfModuleCtxType:
		return proto.ModuleCtxType_MODULE_CTX_TYPE_SELF
	case tflint.RootModuleCtxType:
		return proto.ModuleCtxType_MODULE_CTX_TYPE_ROOT
	default:
		panic(fmt.Sprintf("invalid ModuleCtxType: %s", ty.String()))
	}
}

// ExpandMode converts tflint.ExpandMode to proto.GetModuleContent_ExpandMode
func ExpandMode(mode tflint.ExpandMode) proto.GetModuleContent_ExpandMode {
	switch mode {
	case tflint.ExpandModeExpand:
		return proto.GetModuleContent_EXPAND_MODE_EXPAND
	case tflint.ExpandModeNone:
		return proto.GetModuleContent_EXPAND_MODE_NONE
	default:
		panic(fmt.Sprintf("invalid ExpandMode: %s", mode))
	}
}

// GetModuleContentHint converts tflint.GetModuleContentHint to proto.GetModuleContentHint
func GetModuleContentHint(hint tflint.GetModuleContentHint) *proto.GetModuleContent_Hint {
	return &proto.GetModuleContent_Hint{
		ResourceType: hint.ResourceType,
	}
}

// Error converts error to gRPC error status with details
func Error(code codes.Code, err error) error {
	if err == nil {
		return nil
	}

	var errCode proto.ErrorCode
	if errors.Is(err, tflint.ErrUnknownValue) {
		errCode = proto.ErrorCode_ERROR_CODE_UNKNOWN_VALUE
	} else if errors.Is(err, tflint.ErrNullValue) {
		errCode = proto.ErrorCode_ERROR_CODE_NULL_VALUE
	} else if errors.Is(err, tflint.ErrUnevaluable) {
		errCode = proto.ErrorCode_ERROR_CODE_UNEVALUABLE
	} else if errors.Is(err, tflint.ErrSensitive) {
		errCode = proto.ErrorCode_ERROR_CODE_SENSITIVE
	}

	if errCode == proto.ErrorCode_ERROR_CODE_UNSPECIFIED {
		return status.Error(code, err.Error())
	}

	st := status.New(code, err.Error())
	dt, err := st.WithDetails(&proto.ErrorDetail{Code: errCode})
	if err != nil {
		return status.Error(codes.Unknown, fmt.Sprintf("Failed to add ErrorDetail: code=%d error=%s", code, err.Error()))
	}

	return dt.Err()
}
