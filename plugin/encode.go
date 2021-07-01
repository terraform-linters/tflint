package plugin

import (
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/addrs"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/configs"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/experiments"
	tfplugin "github.com/terraform-linters/tflint-plugin-sdk/tflint/client"
	tfconfigs "github.com/terraform-linters/tflint/terraform/configs"
	"github.com/terraform-linters/tflint/tflint"
	"github.com/zclconf/go-cty/cty/json"
	"github.com/zclconf/go-cty/cty/msgpack"
)

func (s *Server) encodeConfig(config *tfconfigs.Config) (*tfplugin.Config, error) {
	var versionStr string
	if config.Version != nil {
		versionStr = config.Version.String()
	}

	module, err := s.encodeModule(config.Module)
	if err != nil {
		return nil, err
	}

	return &tfplugin.Config{
		Path:            addrs.Module(config.Path),
		Module:          module,
		CallRange:       config.CallRange,
		SourceAddr:      config.SourceAddr,
		SourceAddrRange: config.SourceAddrRange,
		Version:         versionStr,
	}, nil
}

func (s *Server) encodeModule(module *tfconfigs.Module) (*tfplugin.Module, error) {
	versionConstraints := make([]string, len(module.CoreVersionConstraints))
	versionConstraintRanges := make([]hcl.Range, len(module.CoreVersionConstraints))
	for i, v := range module.CoreVersionConstraints {
		versionConstraints[i] = v.Required.String()
		versionConstraintRanges[i] = v.DeclRange
	}

	experimentSet := experiments.Set{}
	for k, v := range module.ActiveExperiments {
		experimentSet[experiments.Experiment(k)] = v
	}

	providers := map[string]*tfplugin.Provider{}
	for k, v := range module.ProviderConfigs {
		providers[k] = s.encodeProvider(v)
	}

	localNames := map[addrs.Provider]string{}
	for k, v := range module.ProviderLocalNames {
		localNames[addrs.Provider(k)] = v
	}

	metas := map[addrs.Provider]*tfplugin.ProviderMeta{}
	for k, v := range module.ProviderMetas {
		metas[addrs.Provider(k)] = s.encodeProviderMeta(v)
	}

	variables := map[string]*tfplugin.Variable{}
	for k, v := range module.Variables {
		var err error
		variables[k], err = s.encodeVariable(v)
		if err != nil {
			return nil, err
		}
	}

	locals := map[string]*tfplugin.Local{}
	for k, v := range module.Locals {
		locals[k] = s.encodeLocal(v)
	}

	outputs := map[string]*tfplugin.Output{}
	for k, v := range module.Outputs {
		outputs[k] = s.encodeOutput(v)
	}

	calls := map[string]*tfplugin.ModuleCall{}
	for k, v := range module.ModuleCalls {
		calls[k] = s.encodeModuleCall(v)
	}

	managed := map[string]*tfplugin.Resource{}
	for k, v := range module.ManagedResources {
		managed[k] = s.encodeResource(v)
	}

	data := map[string]*tfplugin.Resource{}
	for k, v := range module.DataResources {
		data[k] = s.encodeResource(v)
	}

	return &tfplugin.Module{
		SourceDir: module.SourceDir,

		CoreVersionConstraints:      versionConstraints,
		CoreVersionConstraintRanges: versionConstraintRanges,

		ActiveExperiments:    experimentSet,
		Backend:              s.encodeBackend(module.Backend),
		ProviderConfigs:      providers,
		ProviderRequirements: s.encodeRequiredProviders(module.ProviderRequirements),
		ProviderLocalNames:   localNames,
		ProviderMetas:        metas,

		Variables: variables,
		Locals:    locals,
		Outputs:   outputs,

		ModuleCalls: calls,

		ManagedResources: managed,
		DataResources:    data,
	}, nil
}

func (s *Server) encodeResource(resource *tfconfigs.Resource) *tfplugin.Resource {
	configRange := tflint.HCLBodyRange(resource.Config, resource.DeclRange)

	var count []byte
	var countRange hcl.Range
	if resource.Count != nil {
		count = resource.Count.Range().SliceBytes(s.runner.File(resource.Count.Range().Filename).Bytes)
		countRange = resource.Count.Range()
	}

	var forEach []byte
	var forEachRange hcl.Range
	if resource.ForEach != nil {
		forEach = resource.ForEach.Range().SliceBytes(s.runner.File(resource.ForEach.Range().Filename).Bytes)
		forEachRange = resource.ForEach.Range()
	}

	return &tfplugin.Resource{
		Mode:         addrs.ResourceMode(resource.Mode),
		Name:         resource.Name,
		Type:         resource.Type,
		Config:       configRange.SliceBytes(s.runner.File(configRange.Filename).Bytes),
		ConfigRange:  configRange,
		Count:        count,
		CountRange:   countRange,
		ForEach:      forEach,
		ForEachRange: forEachRange,

		ProviderConfigRef: s.encodeProviderConfigRef(resource.ProviderConfigRef),
		Provider: addrs.Provider{
			Type:      resource.Provider.Type,
			Namespace: resource.Provider.Namespace,
			Hostname:  resource.Provider.Hostname,
		},

		Managed: s.encodeManagedResource(resource.Managed),

		DeclRange: resource.DeclRange,
		TypeRange: resource.TypeRange,
	}
}

func (s *Server) encodeManagedResource(resource *tfconfigs.ManagedResource) *tfplugin.ManagedResource {
	if resource == nil {
		return nil
	}

	provisioners := make([]*tfplugin.Provisioner, len(resource.Provisioners))
	for i, provisioner := range resource.Provisioners {
		provisioners[i] = s.encodeProvisioner(provisioner)
	}

	return &tfplugin.ManagedResource{
		Connection:   s.encodeConnection(resource.Connection),
		Provisioners: provisioners,

		CreateBeforeDestroy: resource.CreateBeforeDestroy,
		PreventDestroy:      resource.PreventDestroy,
		IgnoreAllChanges:    resource.IgnoreAllChanges,

		CreateBeforeDestroySet: resource.CreateBeforeDestroySet,
		PreventDestroySet:      resource.PreventDestroySet,
	}
}

func (s *Server) encodeModuleCall(call *tfconfigs.ModuleCall) *tfplugin.ModuleCall {
	configRange := tflint.HCLBodyRange(call.Config, call.DeclRange)

	var count []byte
	var countRange hcl.Range
	if call.Count != nil {
		count = call.Count.Range().SliceBytes(s.runner.File(call.Count.Range().Filename).Bytes)
		countRange = call.Count.Range()
	}

	var forEach []byte
	var forEachRange hcl.Range
	if call.ForEach != nil {
		forEach = call.ForEach.Range().SliceBytes(s.runner.File(call.ForEach.Range().Filename).Bytes)
		forEachRange = call.ForEach.Range()
	}

	providers := []tfplugin.PassedProviderConfig{}
	for _, provider := range call.Providers {
		providers = append(providers, tfplugin.PassedProviderConfig{
			InChild:  s.encodeProviderConfigRef(provider.InChild),
			InParent: s.encodeProviderConfigRef(provider.InParent),
		})
	}

	version := call.Version.Required.String()

	return &tfplugin.ModuleCall{
		Name: call.Name,

		SourceAddr:      call.SourceAddr,
		SourceAddrRange: call.SourceAddrRange,
		SourceSet:       call.SourceSet,

		Version:      version,
		VersionRange: call.Version.DeclRange,

		Config:      configRange.SliceBytes(s.runner.File(configRange.Filename).Bytes),
		ConfigRange: configRange,

		Count:        count,
		CountRange:   countRange,
		ForEach:      forEach,
		ForEachRange: forEachRange,

		Providers: providers,
		DeclRange: call.DeclRange,
	}
}

func (s *Server) encodeProviderConfigRef(ref *tfconfigs.ProviderConfigRef) *configs.ProviderConfigRef {
	if ref == nil {
		return nil
	}

	return &configs.ProviderConfigRef{
		Name:       ref.Name,
		NameRange:  ref.NameRange,
		Alias:      ref.Alias,
		AliasRange: ref.AliasRange,
	}
}

func (s *Server) encodeConnection(connection *tfconfigs.Connection) *tfplugin.Connection {
	if connection == nil {
		return nil
	}

	configRange := tflint.HCLBodyRange(connection.Config, connection.DeclRange)

	return &tfplugin.Connection{
		Config:      configRange.SliceBytes(s.runner.File(configRange.Filename).Bytes),
		ConfigRange: configRange,

		DeclRange: connection.DeclRange,
	}
}

func (s *Server) encodeProvisioner(provisioner *tfconfigs.Provisioner) *tfplugin.Provisioner {
	configRange := tflint.HCLBodyRange(provisioner.Config, provisioner.DeclRange)

	return &tfplugin.Provisioner{
		Type:        provisioner.Type,
		Config:      configRange.SliceBytes(s.runner.File(configRange.Filename).Bytes),
		ConfigRange: configRange,
		Connection:  s.encodeConnection(provisioner.Connection),
		When:        configs.ProvisionerWhen(provisioner.When),
		OnFailure:   configs.ProvisionerOnFailure(provisioner.OnFailure),

		DeclRange: provisioner.DeclRange,
		TypeRange: provisioner.TypeRange,
	}
}

func (s *Server) encodeBackend(backend *tfconfigs.Backend) *tfplugin.Backend {
	if backend == nil {
		return nil
	}

	configRange := tflint.HCLBodyRange(backend.Config, backend.DeclRange)
	config := []byte{}
	if configRange.Empty() {
		configRange.Filename = backend.DeclRange.Filename
	} else {
		config = configRange.SliceBytes(s.runner.File(configRange.Filename).Bytes)
	}

	return &tfplugin.Backend{
		Type:        backend.Type,
		Config:      config,
		ConfigRange: configRange,
		DeclRange:   backend.DeclRange,
		TypeRange:   backend.TypeRange,
	}
}

func (s *Server) encodeProvider(provider *tfconfigs.Provider) *tfplugin.Provider {
	configRange := tflint.HCLBodyRange(provider.Config, provider.DeclRange)
	config := []byte{}
	if configRange.Empty() {
		configRange.Filename = provider.DeclRange.Filename
	} else {
		config = configRange.SliceBytes(s.runner.File(configRange.Filename).Bytes)
	}

	return &tfplugin.Provider{
		Name:       provider.Name,
		NameRange:  provider.NameRange,
		Alias:      provider.Alias,
		AliasRange: provider.AliasRange,

		Version:      provider.Version.Required.String(),
		VersionRange: provider.Version.DeclRange,

		Config:      config,
		ConfigRange: configRange,

		DeclRange: provider.DeclRange,
	}
}

func (s *Server) encodeRequiredProviders(providers *tfconfigs.RequiredProviders) *tfplugin.RequiredProviders {
	if providers == nil {
		return nil
	}

	ret := map[string]*tfplugin.RequiredProvider{}
	for k, v := range providers.RequiredProviders {
		ret[k] = s.encodeRequiredProvider(v)
	}

	return &tfplugin.RequiredProviders{
		RequiredProviders: ret,
		DeclRange:         providers.DeclRange,
	}
}

func (s *Server) encodeRequiredProvider(provider *tfconfigs.RequiredProvider) *tfplugin.RequiredProvider {
	return &tfplugin.RequiredProvider{
		Name:             provider.Name,
		Source:           provider.Source,
		Type:             addrs.Provider(provider.Type),
		Requirement:      provider.Requirement.Required.String(),
		RequirementRange: provider.Requirement.DeclRange,
		DeclRange:        provider.DeclRange,
	}
}

func (s *Server) encodeProviderMeta(meta *tfconfigs.ProviderMeta) *tfplugin.ProviderMeta {
	configRange := tflint.HCLBodyRange(meta.Config, meta.DeclRange)

	return &tfplugin.ProviderMeta{
		Provider:    meta.Provider,
		Config:      configRange.SliceBytes(s.runner.File(configRange.Filename).Bytes),
		ConfigRange: configRange,

		ProviderRange: meta.ProviderRange,
		DeclRange:     meta.DeclRange,
	}
}

func (s *Server) encodeVariable(variable *tfconfigs.Variable) (*tfplugin.Variable, error) {
	validations := make([]*tfplugin.VariableValidation, len(variable.Validations))
	for i, v := range variable.Validations {
		validations[i] = s.encodeVariableValidation(v)
	}

	defaultVal, err := msgpack.Marshal(variable.Default, variable.Default.Type())
	if err != nil {
		return nil, err
	}
	// We need to use json because cty.DynamicPseudoType cannot be encoded by gob
	typeVal, err := json.MarshalType(variable.Type)
	if err != nil {
		return nil, err
	}
	return &tfplugin.Variable{
		Name:        variable.Name,
		Description: variable.Description,
		Default:     defaultVal,
		Type:        typeVal,
		ParsingMode: configs.VariableParsingMode(variable.ParsingMode),
		Validations: validations,
		Sensitive:   variable.Sensitive,

		DescriptionSet: variable.DescriptionSet,
		SensitiveSet:   variable.SensitiveSet,

		DeclRange: variable.DeclRange,
	}, nil
}

func (s *Server) encodeVariableValidation(validation *tfconfigs.VariableValidation) *tfplugin.VariableValidation {
	var condition []byte
	var conditionRange hcl.Range
	if validation.Condition != nil {
		condition = validation.Condition.Range().SliceBytes(s.runner.File(validation.Condition.Range().Filename).Bytes)
		conditionRange = validation.Condition.Range()
	}

	return &tfplugin.VariableValidation{
		Condition:      condition,
		ConditionRange: conditionRange,

		ErrorMessage: validation.ErrorMessage,

		DeclRange: validation.DeclRange,
	}
}

func (s *Server) encodeLocal(local *tfconfigs.Local) *tfplugin.Local {
	var expr []byte
	var exprRange hcl.Range
	if local.Expr != nil {
		expr = local.Expr.Range().SliceBytes(s.runner.File(local.Expr.Range().Filename).Bytes)
		exprRange = local.Expr.Range()
	}

	return &tfplugin.Local{
		Name:      local.Name,
		Expr:      expr,
		ExprRange: exprRange,

		DeclRange: local.DeclRange,
	}
}

func (s *Server) encodeOutput(output *tfconfigs.Output) *tfplugin.Output {
	var expr []byte
	var exprRange hcl.Range
	if output.Expr != nil {
		expr = output.Expr.Range().SliceBytes(s.runner.File(output.Expr.Range().Filename).Bytes)
		exprRange = output.Expr.Range()
	}

	return &tfplugin.Output{
		Name:        output.Name,
		Description: output.Description,
		Expr:        expr,
		ExprRange:   exprRange,
		Sensitive:   output.Sensitive,

		DescriptionSet: output.DescriptionSet,
		SensitiveSet:   output.SensitiveSet,

		DeclRange: output.DeclRange,
	}
}
