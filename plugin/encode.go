package plugin

import (
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/terraform/configs"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform"
	tfplugin "github.com/terraform-linters/tflint-plugin-sdk/tflint/client"
	"github.com/terraform-linters/tflint/tflint"
)

func (s *Server) encodeResource(resource *configs.Resource) *tfplugin.Resource {
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
		Mode:         terraform.ResourceMode(resource.Mode),
		Name:         resource.Name,
		Type:         resource.Type,
		Config:       configRange.SliceBytes(s.runner.File(configRange.Filename).Bytes),
		ConfigRange:  configRange,
		Count:        count,
		CountRange:   countRange,
		ForEach:      forEach,
		ForEachRange: forEachRange,

		ProviderConfigRef: s.encodeProviderConfigRef(resource.ProviderConfigRef),

		Managed: s.encodeManagedResource(resource.Managed),

		DeclRange: resource.DeclRange,
		TypeRange: resource.TypeRange,
	}
}

func (s *Server) encodeManagedResource(resource *configs.ManagedResource) *tfplugin.ManagedResource {
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

func (s *Server) encodeProviderConfigRef(ref *configs.ProviderConfigRef) *terraform.ProviderConfigRef {
	if ref == nil {
		return nil
	}

	return &terraform.ProviderConfigRef{
		Name:       ref.Name,
		NameRange:  ref.NameRange,
		Alias:      ref.Alias,
		AliasRange: ref.AliasRange,
	}
}

func (s *Server) encodeConnection(connection *configs.Connection) *tfplugin.Connection {
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

func (s *Server) encodeProvisioner(provisioner *configs.Provisioner) *tfplugin.Provisioner {
	configRange := tflint.HCLBodyRange(provisioner.Config, provisioner.DeclRange)

	return &tfplugin.Provisioner{
		Type:        provisioner.Type,
		Config:      configRange.SliceBytes(s.runner.File(configRange.Filename).Bytes),
		ConfigRange: configRange,
		Connection:  s.encodeConnection(provisioner.Connection),
		When:        terraform.ProvisionerWhen(provisioner.When),
		OnFailure:   terraform.ProvisionerOnFailure(provisioner.OnFailure),

		DeclRange: provisioner.DeclRange,
		TypeRange: provisioner.TypeRange,
	}
}

func (s *Server) encodeBackend(backend *configs.Backend) *tfplugin.Backend {
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
