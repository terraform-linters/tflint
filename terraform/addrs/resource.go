package addrs

import (
	"fmt"
	"strings"
)

// Resource is an address for a resource block within configuration, which
// contains potentially-multiple resource instances if that configuration
// block uses "count" or "for_each".
type Resource struct {
	referenceable
	Mode ResourceMode
	Type string
	Name string
}

func (r Resource) String() string {
	switch r.Mode {
	case ManagedResourceMode:
		return fmt.Sprintf("%s.%s", r.Type, r.Name)
	case DataResourceMode:
		return fmt.Sprintf("data.%s.%s", r.Type, r.Name)
	default:
		// Should never happen, but we'll return a string here rather than
		// crashing just in case it does.
		return fmt.Sprintf("<invalid>.%s.%s", r.Type, r.Name)
	}
}

func (r Resource) Equal(o Resource) bool {
	return r.Mode == o.Mode && r.Name == o.Name && r.Type == o.Type
}

func (r Resource) UniqueKey() UniqueKey {
	return r // A Resource is its own UniqueKey
}

func (r Resource) uniqueKeySigil() {}

// Instance produces the address for a specific instance of the receiver
// that is idenfied by the given key.
func (r Resource) Instance(key InstanceKey) ResourceInstance {
	return ResourceInstance{
		Resource: r,
		Key:      key,
	}
}

// Absolute returns an AbsResource from the receiver and the given module
// instance address.
func (r Resource) Absolute(module ModuleInstance) AbsResource {
	return AbsResource{
		Module:   module,
		Resource: r,
	}
}

// InModule returns a ConfigResource from the receiver and the given module
// address.
func (r Resource) InModule(module Module) ConfigResource {
	return ConfigResource{
		Module:   module,
		Resource: r,
	}
}

// ImpliedProvider returns the implied provider type name, for e.g. the "aws" in
// "aws_instance"
func (r Resource) ImpliedProvider() string {
	typeName := r.Type
	if under := strings.Index(typeName, "_"); under != -1 {
		typeName = typeName[:under]
	}

	return typeName
}

// ResourceInstance is an address for a specific instance of a resource.
// When a resource is defined in configuration with "count" or "for_each" it
// produces zero or more instances, which can be addressed using this type.
type ResourceInstance struct {
	referenceable
	Resource Resource
	Key      InstanceKey
}

func (r ResourceInstance) ContainingResource() Resource {
	return r.Resource
}

func (r ResourceInstance) String() string {
	if r.Key == NoKey {
		return r.Resource.String()
	}
	return r.Resource.String() + r.Key.String()
}

func (r ResourceInstance) Equal(o ResourceInstance) bool {
	return r.Key == o.Key && r.Resource.Equal(o.Resource)
}

func (r ResourceInstance) UniqueKey() UniqueKey {
	return r // A ResourceInstance is its own UniqueKey
}

func (r ResourceInstance) uniqueKeySigil() {}

// Absolute returns an AbsResourceInstance from the receiver and the given module
// instance address.
func (r ResourceInstance) Absolute(module ModuleInstance) AbsResourceInstance {
	return AbsResourceInstance{
		Module:   module,
		Resource: r,
	}
}

// AbsResource is an absolute address for a resource under a given module path.
type AbsResource struct {
	targetable
	Module   ModuleInstance
	Resource Resource
}

// Resource returns the address of a particular resource within the receiver.
func (m ModuleInstance) Resource(mode ResourceMode, typeName string, name string) AbsResource {
	return AbsResource{
		Module: m,
		Resource: Resource{
			Mode: mode,
			Type: typeName,
			Name: name,
		},
	}
}

// Instance produces the address for a specific instance of the receiver
// that is idenfied by the given key.
func (r AbsResource) Instance(key InstanceKey) AbsResourceInstance {
	return AbsResourceInstance{
		Module:   r.Module,
		Resource: r.Resource.Instance(key),
	}
}

// Config returns the unexpanded ConfigResource for this AbsResource.
func (r AbsResource) Config() ConfigResource {
	return ConfigResource{
		Module:   r.Module.Module(),
		Resource: r.Resource,
	}
}

// TargetContains implements Targetable by returning true if the given other
// address is either equal to the receiver or is an instance of the
// receiver.
func (r AbsResource) TargetContains(other Targetable) bool {
	switch to := other.(type) {

	case AbsResource:
		// We'll use our stringification as a cheat-ish way to test for equality.
		return to.String() == r.String()

	case ConfigResource:
		// if an absolute resource from parsing a target address contains a
		// ConfigResource, the string representation will match
		return to.String() == r.String()

	case AbsResourceInstance:
		return r.TargetContains(to.ContainingResource())

	default:
		return false

	}
}

func (r AbsResource) AddrType() TargetableAddrType {
	return AbsResourceAddrType
}

func (r AbsResource) String() string {
	if len(r.Module) == 0 {
		return r.Resource.String()
	}
	return fmt.Sprintf("%s.%s", r.Module.String(), r.Resource.String())
}

// AffectedAbsResource returns the AbsResource.
func (r AbsResource) AffectedAbsResource() AbsResource {
	return r
}

func (r AbsResource) Equal(o AbsResource) bool {
	return r.Module.Equal(o.Module) && r.Resource.Equal(o.Resource)
}

func (r AbsResource) absMoveableSigil() {
	// AbsResource is moveable
}

type absResourceKey string

func (r absResourceKey) uniqueKeySigil() {}

func (r AbsResource) UniqueKey() UniqueKey {
	return absResourceKey(r.String())
}

// AbsResourceInstance is an absolute address for a resource instance under a
// given module path.
type AbsResourceInstance struct {
	targetable
	Module   ModuleInstance
	Resource ResourceInstance
}

// ResourceInstance returns the address of a particular resource instance within the receiver.
func (m ModuleInstance) ResourceInstance(mode ResourceMode, typeName string, name string, key InstanceKey) AbsResourceInstance {
	return AbsResourceInstance{
		Module: m,
		Resource: ResourceInstance{
			Resource: Resource{
				Mode: mode,
				Type: typeName,
				Name: name,
			},
			Key: key,
		},
	}
}

// ContainingResource returns the address of the resource that contains the
// receving resource instance. In other words, it discards the key portion
// of the address to produce an AbsResource value.
func (r AbsResourceInstance) ContainingResource() AbsResource {
	return AbsResource{
		Module:   r.Module,
		Resource: r.Resource.ContainingResource(),
	}
}

// TargetContains implements Targetable by returning true if the given other
// address is equal to the receiver.
func (r AbsResourceInstance) TargetContains(other Targetable) bool {
	switch to := other.(type) {

	// while we currently don't start with an AbsResourceInstance as a target
	// address, check all resource types for consistency.
	case AbsResourceInstance:
		// We'll use our stringification as a cheat-ish way to test for equality.
		return to.String() == r.String()
	case ConfigResource:
		return to.String() == r.String()
	case AbsResource:
		return to.String() == r.String()

	default:
		return false

	}
}

func (r AbsResourceInstance) AddrType() TargetableAddrType {
	return AbsResourceInstanceAddrType
}

func (r AbsResourceInstance) String() string {
	if len(r.Module) == 0 {
		return r.Resource.String()
	}
	return fmt.Sprintf("%s.%s", r.Module.String(), r.Resource.String())
}

// AffectedAbsResource returns the AbsResource for the instance.
func (r AbsResourceInstance) AffectedAbsResource() AbsResource {
	return AbsResource{
		Module:   r.Module,
		Resource: r.Resource.Resource,
	}
}

func (r AbsResourceInstance) Equal(o AbsResourceInstance) bool {
	return r.Module.Equal(o.Module) && r.Resource.Equal(o.Resource)
}

// Less returns true if the receiver should sort before the given other value
// in a sorted list of addresses.
func (r AbsResourceInstance) Less(o AbsResourceInstance) bool {
	switch {

	case len(r.Module) != len(o.Module):
		return len(r.Module) < len(o.Module)

	case r.Module.String() != o.Module.String():
		return r.Module.Less(o.Module)

	case r.Resource.Resource.Mode != o.Resource.Resource.Mode:
		return r.Resource.Resource.Mode == DataResourceMode

	case r.Resource.Resource.Type != o.Resource.Resource.Type:
		return r.Resource.Resource.Type < o.Resource.Resource.Type

	case r.Resource.Resource.Name != o.Resource.Resource.Name:
		return r.Resource.Resource.Name < o.Resource.Resource.Name

	case r.Resource.Key != o.Resource.Key:
		return InstanceKeyLess(r.Resource.Key, o.Resource.Key)

	default:
		return false

	}
}

type absResourceInstanceKey string

func (r AbsResourceInstance) UniqueKey() UniqueKey {
	return absResourceInstanceKey(r.String())
}

func (r absResourceInstanceKey) uniqueKeySigil() {}

func (r AbsResourceInstance) absMoveableSigil() {
	// AbsResourceInstance is moveable
}

// ConfigResource is an address for a resource within a configuration.
type ConfigResource struct {
	targetable
	Module   Module
	Resource Resource
}

// Resource returns the address of a particular resource within the module.
func (m Module) Resource(mode ResourceMode, typeName string, name string) ConfigResource {
	return ConfigResource{
		Module: m,
		Resource: Resource{
			Mode: mode,
			Type: typeName,
			Name: name,
		},
	}
}

// Absolute produces the address for the receiver within a specific module instance.
func (r ConfigResource) Absolute(module ModuleInstance) AbsResource {
	return AbsResource{
		Module:   module,
		Resource: r.Resource,
	}
}

// TargetContains implements Targetable by returning true if the given other
// address is either equal to the receiver or is an instance of the
// receiver.
func (r ConfigResource) TargetContains(other Targetable) bool {
	switch to := other.(type) {
	case ConfigResource:
		// We'll use our stringification as a cheat-ish way to test for equality.
		return to.String() == r.String()
	case AbsResource:
		return r.TargetContains(to.Config())
	case AbsResourceInstance:
		return r.TargetContains(to.ContainingResource())
	default:
		return false
	}
}

func (r ConfigResource) AddrType() TargetableAddrType {
	return ConfigResourceAddrType
}

func (r ConfigResource) String() string {
	if len(r.Module) == 0 {
		return r.Resource.String()
	}
	return fmt.Sprintf("%s.%s", r.Module.String(), r.Resource.String())
}

func (r ConfigResource) Equal(o ConfigResource) bool {
	return r.Module.Equal(o.Module) && r.Resource.Equal(o.Resource)
}

func (r ConfigResource) configMoveableSigil() {
	// AbsResource is moveable
}

// ResourceMode defines which lifecycle applies to a given resource. Each
// resource lifecycle has a slightly different address format.
type ResourceMode rune

//go:generate go run golang.org/x/tools/cmd/stringer -type ResourceMode

const (
	// InvalidResourceMode is the zero value of ResourceMode and is not
	// a valid resource mode.
	InvalidResourceMode ResourceMode = 0

	// ManagedResourceMode indicates a managed resource, as defined by
	// "resource" blocks in configuration.
	ManagedResourceMode ResourceMode = 'M'

	// DataResourceMode indicates a data resource, as defined by
	// "data" blocks in configuration.
	DataResourceMode ResourceMode = 'D'
)
