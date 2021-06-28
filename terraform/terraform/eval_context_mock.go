package terraform

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/terraform-linters/tflint/terraform/addrs"
	"github.com/terraform-linters/tflint/terraform/configs/configschema"
	"github.com/terraform-linters/tflint/terraform/instances"
	"github.com/terraform-linters/tflint/terraform/lang"
	"github.com/terraform-linters/tflint/terraform/plans"
	"github.com/terraform-linters/tflint/terraform/providers"
	"github.com/terraform-linters/tflint/terraform/provisioners"
	"github.com/terraform-linters/tflint/terraform/states"
	"github.com/terraform-linters/tflint/terraform/tfdiags"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

// MockEvalContext is a mock version of EvalContext that can be used
// for tests.
type MockEvalContext struct {
	StoppedCalled bool
	StoppedValue  <-chan struct{}

	HookCalled bool
	HookHook   Hook
	HookError  error

	InputCalled bool
	InputInput  UIInput

	InitProviderCalled   bool
	InitProviderType     string
	InitProviderAddr     addrs.AbsProviderConfig
	InitProviderProvider providers.Interface
	InitProviderError    error

	ProviderCalled   bool
	ProviderAddr     addrs.AbsProviderConfig
	ProviderProvider providers.Interface

	ProviderSchemaCalled bool
	ProviderSchemaAddr   addrs.AbsProviderConfig
	ProviderSchemaSchema *ProviderSchema

	CloseProviderCalled   bool
	CloseProviderAddr     addrs.AbsProviderConfig
	CloseProviderProvider providers.Interface

	ProviderInputCalled bool
	ProviderInputAddr   addrs.AbsProviderConfig
	ProviderInputValues map[string]cty.Value

	SetProviderInputCalled bool
	SetProviderInputAddr   addrs.AbsProviderConfig
	SetProviderInputValues map[string]cty.Value

	ConfigureProviderFn func(
		addr addrs.AbsProviderConfig,
		cfg cty.Value) tfdiags.Diagnostics // overrides the other values below, if set
	ConfigureProviderCalled bool
	ConfigureProviderAddr   addrs.AbsProviderConfig
	ConfigureProviderConfig cty.Value
	ConfigureProviderDiags  tfdiags.Diagnostics

	ProvisionerCalled      bool
	ProvisionerName        string
	ProvisionerProvisioner provisioners.Interface

	ProvisionerSchemaCalled bool
	ProvisionerSchemaName   string
	ProvisionerSchemaSchema *configschema.Block

	CloseProvisionersCalled bool

	EvaluateBlockCalled     bool
	EvaluateBlockBody       hcl.Body
	EvaluateBlockSchema     *configschema.Block
	EvaluateBlockSelf       addrs.Referenceable
	EvaluateBlockKeyData    InstanceKeyEvalData
	EvaluateBlockResultFunc func(
		body hcl.Body,
		schema *configschema.Block,
		self addrs.Referenceable,
		keyData InstanceKeyEvalData,
	) (cty.Value, hcl.Body, tfdiags.Diagnostics) // overrides the other values below, if set
	EvaluateBlockResult       cty.Value
	EvaluateBlockExpandedBody hcl.Body
	EvaluateBlockDiags        tfdiags.Diagnostics

	EvaluateExprCalled     bool
	EvaluateExprExpr       hcl.Expression
	EvaluateExprWantType   cty.Type
	EvaluateExprSelf       addrs.Referenceable
	EvaluateExprResultFunc func(
		expr hcl.Expression,
		wantType cty.Type,
		self addrs.Referenceable,
	) (cty.Value, tfdiags.Diagnostics) // overrides the other values below, if set
	EvaluateExprResult cty.Value
	EvaluateExprDiags  tfdiags.Diagnostics

	EvaluationScopeCalled  bool
	EvaluationScopeSelf    addrs.Referenceable
	EvaluationScopeKeyData InstanceKeyEvalData
	EvaluationScopeScope   *lang.Scope

	PathCalled bool
	PathPath   addrs.ModuleInstance

	SetModuleCallArgumentsCalled bool
	SetModuleCallArgumentsModule addrs.ModuleCallInstance
	SetModuleCallArgumentsValues map[string]cty.Value

	GetVariableValueCalled bool
	GetVariableValueAddr   addrs.AbsInputVariableInstance
	GetVariableValueValue  cty.Value

	ChangesCalled  bool
	ChangesChanges *plans.ChangesSync

	StateCalled bool
	StateState  *states.SyncState

	RefreshStateCalled bool
	RefreshStateState  *states.SyncState

	PrevRunStateCalled bool
	PrevRunStateState  *states.SyncState

	InstanceExpanderCalled   bool
	InstanceExpanderExpander *instances.Expander
}

// MockEvalContext implements EvalContext
var _ EvalContext = (*MockEvalContext)(nil)

func (c *MockEvalContext) Stopped() <-chan struct{} {
	c.StoppedCalled = true
	return c.StoppedValue
}

func (c *MockEvalContext) Hook(fn func(Hook) (HookAction, error)) error {
	c.HookCalled = true
	if c.HookHook != nil {
		if _, err := fn(c.HookHook); err != nil {
			return err
		}
	}

	return c.HookError
}

func (c *MockEvalContext) Input() UIInput {
	c.InputCalled = true
	return c.InputInput
}

func (c *MockEvalContext) InitProvider(addr addrs.AbsProviderConfig) (providers.Interface, error) {
	c.InitProviderCalled = true
	c.InitProviderType = addr.String()
	c.InitProviderAddr = addr
	return c.InitProviderProvider, c.InitProviderError
}

func (c *MockEvalContext) Provider(addr addrs.AbsProviderConfig) providers.Interface {
	c.ProviderCalled = true
	c.ProviderAddr = addr
	return c.ProviderProvider
}

func (c *MockEvalContext) ProviderSchema(addr addrs.AbsProviderConfig) *ProviderSchema {
	c.ProviderSchemaCalled = true
	c.ProviderSchemaAddr = addr
	return c.ProviderSchemaSchema
}

func (c *MockEvalContext) CloseProvider(addr addrs.AbsProviderConfig) error {
	c.CloseProviderCalled = true
	c.CloseProviderAddr = addr
	return nil
}

func (c *MockEvalContext) ConfigureProvider(addr addrs.AbsProviderConfig, cfg cty.Value) tfdiags.Diagnostics {

	c.ConfigureProviderCalled = true
	c.ConfigureProviderAddr = addr
	c.ConfigureProviderConfig = cfg
	if c.ConfigureProviderFn != nil {
		return c.ConfigureProviderFn(addr, cfg)
	}
	return c.ConfigureProviderDiags
}

func (c *MockEvalContext) ProviderInput(addr addrs.AbsProviderConfig) map[string]cty.Value {
	c.ProviderInputCalled = true
	c.ProviderInputAddr = addr
	return c.ProviderInputValues
}

func (c *MockEvalContext) SetProviderInput(addr addrs.AbsProviderConfig, vals map[string]cty.Value) {
	c.SetProviderInputCalled = true
	c.SetProviderInputAddr = addr
	c.SetProviderInputValues = vals
}

func (c *MockEvalContext) Provisioner(n string) (provisioners.Interface, error) {
	c.ProvisionerCalled = true
	c.ProvisionerName = n
	return c.ProvisionerProvisioner, nil
}

func (c *MockEvalContext) ProvisionerSchema(n string) *configschema.Block {
	c.ProvisionerSchemaCalled = true
	c.ProvisionerSchemaName = n
	return c.ProvisionerSchemaSchema
}

func (c *MockEvalContext) CloseProvisioners() error {
	c.CloseProvisionersCalled = true
	return nil
}

func (c *MockEvalContext) EvaluateBlock(body hcl.Body, schema *configschema.Block, self addrs.Referenceable, keyData InstanceKeyEvalData) (cty.Value, hcl.Body, tfdiags.Diagnostics) {
	c.EvaluateBlockCalled = true
	c.EvaluateBlockBody = body
	c.EvaluateBlockSchema = schema
	c.EvaluateBlockSelf = self
	c.EvaluateBlockKeyData = keyData
	if c.EvaluateBlockResultFunc != nil {
		return c.EvaluateBlockResultFunc(body, schema, self, keyData)
	}
	return c.EvaluateBlockResult, c.EvaluateBlockExpandedBody, c.EvaluateBlockDiags
}

func (c *MockEvalContext) EvaluateExpr(expr hcl.Expression, wantType cty.Type, self addrs.Referenceable) (cty.Value, tfdiags.Diagnostics) {
	c.EvaluateExprCalled = true
	c.EvaluateExprExpr = expr
	c.EvaluateExprWantType = wantType
	c.EvaluateExprSelf = self
	if c.EvaluateExprResultFunc != nil {
		return c.EvaluateExprResultFunc(expr, wantType, self)
	}
	return c.EvaluateExprResult, c.EvaluateExprDiags
}

// installSimpleEval is a helper to install a simple mock implementation of
// both EvaluateBlock and EvaluateExpr into the receiver.
//
// These default implementations will either evaluate the given input against
// the scope in field EvaluationScopeScope or, if it is nil, with no eval
// context at all so that only constant values may be used.
//
// This function overwrites any existing functions installed in fields
// EvaluateBlockResultFunc and EvaluateExprResultFunc.
func (c *MockEvalContext) installSimpleEval() {
	c.EvaluateBlockResultFunc = func(body hcl.Body, schema *configschema.Block, self addrs.Referenceable, keyData InstanceKeyEvalData) (cty.Value, hcl.Body, tfdiags.Diagnostics) {
		if scope := c.EvaluationScopeScope; scope != nil {
			// Fully-functional codepath.
			var diags tfdiags.Diagnostics
			body, diags = scope.ExpandBlock(body, schema)
			if diags.HasErrors() {
				return cty.DynamicVal, body, diags
			}
			val, evalDiags := c.EvaluationScopeScope.EvalBlock(body, schema)
			diags = diags.Append(evalDiags)
			if evalDiags.HasErrors() {
				return cty.DynamicVal, body, diags
			}
			return val, body, diags
		}

		// Fallback codepath supporting constant values only.
		val, hclDiags := hcldec.Decode(body, schema.DecoderSpec(), nil)
		return val, body, tfdiags.Diagnostics(nil).Append(hclDiags)
	}
	c.EvaluateExprResultFunc = func(expr hcl.Expression, wantType cty.Type, self addrs.Referenceable) (cty.Value, tfdiags.Diagnostics) {
		if scope := c.EvaluationScopeScope; scope != nil {
			// Fully-functional codepath.
			return scope.EvalExpr(expr, wantType)
		}

		// Fallback codepath supporting constant values only.
		var diags tfdiags.Diagnostics
		val, hclDiags := expr.Value(nil)
		diags = diags.Append(hclDiags)
		if hclDiags.HasErrors() {
			return cty.DynamicVal, diags
		}
		var err error
		val, err = convert.Convert(val, wantType)
		if err != nil {
			diags = diags.Append(err)
			return cty.DynamicVal, diags
		}
		return val, diags
	}
}

func (c *MockEvalContext) EvaluationScope(self addrs.Referenceable, keyData InstanceKeyEvalData) *lang.Scope {
	c.EvaluationScopeCalled = true
	c.EvaluationScopeSelf = self
	c.EvaluationScopeKeyData = keyData
	return c.EvaluationScopeScope
}

func (c *MockEvalContext) WithPath(path addrs.ModuleInstance) EvalContext {
	newC := *c
	newC.PathPath = path
	return &newC
}

func (c *MockEvalContext) Path() addrs.ModuleInstance {
	c.PathCalled = true
	return c.PathPath
}

func (c *MockEvalContext) SetModuleCallArguments(n addrs.ModuleCallInstance, values map[string]cty.Value) {
	c.SetModuleCallArgumentsCalled = true
	c.SetModuleCallArgumentsModule = n
	c.SetModuleCallArgumentsValues = values
}

func (c *MockEvalContext) GetVariableValue(addr addrs.AbsInputVariableInstance) cty.Value {
	c.GetVariableValueCalled = true
	c.GetVariableValueAddr = addr
	return c.GetVariableValueValue
}

func (c *MockEvalContext) Changes() *plans.ChangesSync {
	c.ChangesCalled = true
	return c.ChangesChanges
}

func (c *MockEvalContext) State() *states.SyncState {
	c.StateCalled = true
	return c.StateState
}

func (c *MockEvalContext) RefreshState() *states.SyncState {
	c.RefreshStateCalled = true
	return c.RefreshStateState
}

func (c *MockEvalContext) PrevRunState() *states.SyncState {
	c.PrevRunStateCalled = true
	return c.PrevRunStateState
}

func (c *MockEvalContext) InstanceExpander() *instances.Expander {
	c.InstanceExpanderCalled = true
	return c.InstanceExpanderExpander
}
