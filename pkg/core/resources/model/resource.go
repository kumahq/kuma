package model

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"
	"k8s.io/kube-openapi/pkg/validation/spec"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
)

const (
	DefaultMesh = "default"
	// NoMesh defines a marker that resource is not bound to a Mesh.
	// Resources not bound to a mesh (ScopeGlobal) should have an empty string in Mesh field.
	NoMesh = ""
)

// ResourceNameExtensionsUnsupported is a convenience constant
// that is meant to make source code more readable.
var ResourceNameExtensionsUnsupported = ResourceNameExtensions(nil)

func WithMesh(mesh string, name string) ResourceKey {
	return ResourceKey{Mesh: mesh, Name: name}
}

func WithoutMesh(name string) ResourceKey {
	return ResourceKey{Mesh: NoMesh, Name: name}
}

type ResourceKey struct {
	Mesh string
	Name string
}

type ResourceScope string

const (
	ScopeMesh   = "Mesh"
	ScopeGlobal = "Global"
)

type KDSFlagType uint32

const (
	KDSDisabled      = KDSFlagType(0)
	ConsumedByZone   = KDSFlagType(1)
	ConsumedByGlobal = KDSFlagType(1 << 2)
	ProvidedByZone   = KDSFlagType(1 << 3)
	ProvidedByGlobal = KDSFlagType(1 << 4)
	FromZoneToGlobal = ConsumedByGlobal | ProvidedByZone
	FromGlobalToZone = ProvidedByGlobal | ConsumedByZone
)

// Has return whether this flag has all the passed flags on.
func (kt KDSFlagType) Has(flag KDSFlagType) bool {
	return kt&flag != 0
}

type ResourceSpec interface{}

type Resource interface {
	GetMeta() ResourceMeta
	SetMeta(ResourceMeta)
	GetSpec() ResourceSpec
	SetSpec(ResourceSpec) error
	Descriptor() ResourceTypeDescriptor
}

type ResourceValidator interface {
	Validate() error
}

func Validate(resource Resource) error {
	if rv, ok := resource.(ResourceValidator); ok {
		return rv.Validate()
	}
	return nil
}

type ResourceTypeDescriptor struct {
	// Name identifier of this resourceType this maps to the k8s entity and universal name.
	Name ResourceType
	// Resource a created element of this type
	Resource Resource
	// ResourceList a create list container of this type
	ResourceList ResourceList
	// ReadOnly if this type will be created, modified and deleted by the system.
	ReadOnly bool
	// AdminOnly if this type requires users to be admin to access.
	AdminOnly bool
	// Scope whether this resource is Global or Mesh scoped.
	Scope ResourceScope
	// KDSFlags a set of flags that defines how this entity is sent using KDS (if unset KDS is disabled).
	KDSFlags KDSFlagType
	// WsPath the path to access on the REST api.
	WsPath string
	// KumactlArg the name of the cmdline argument when doing `get` or `delete`.
	KumactlArg string
	// KumactlListArg the name of the cmdline argument when doing `list`.
	KumactlListArg string
	// AllowToInspect if it's required to generate Inspect API endpoint for this type
	AllowToInspect bool
	// IsPolicy if this type is a policy (Dataplanes, Insights, Ingresses are not policies as they describe either metadata or workload, Retries are policies).
	IsPolicy bool
	// DisplayName the name of the policy showed as plural to be displayed in the UI and maybe CLI
	SingularDisplayName string
	// PluralDisplayName the name of the policy showed as plural to be displayed in the UI and maybe CLI
	PluralDisplayName string
	// IsExperimental indicates if a policy is in experimental state (might not be production ready).
	IsExperimental bool
	// IsPluginOriginated indicates if a policy is implemented as a plugin
	IsPluginOriginated bool
	// IsTargetRefBased indicates if a policy uses targetRef or not
	IsTargetRefBased bool
	// HasToTargetRef indicates that the policy can be applied to outbound traffic
	HasToTargetRef bool
	// HasFromTargetRef indicates that the policy can be applied to outbound traffic
	HasFromTargetRef bool
	// Schema contains an unmarshalled OpenAPI schema of the resource
	Schema *spec.Schema
}

func (d ResourceTypeDescriptor) NewObject() Resource {
	specType := reflect.TypeOf(d.Resource.GetSpec()).Elem()
	newSpec := reflect.New(specType).Interface().(ResourceSpec)

	resType := reflect.TypeOf(d.Resource).Elem()
	resource := reflect.New(resType).Interface().(Resource)

	if err := resource.SetSpec(newSpec); err != nil {
		panic(errors.Wrap(err, "could not set spec on the new resource"))
	}

	return resource
}

func (d ResourceTypeDescriptor) NewList() ResourceList {
	listType := reflect.TypeOf(d.ResourceList).Elem()
	return reflect.New(listType).Interface().(ResourceList)
}

type TypeFilter interface {
	Apply(descriptor ResourceTypeDescriptor) bool
}

type TypeFilterFn func(descriptor ResourceTypeDescriptor) bool

func (f TypeFilterFn) Apply(descriptor ResourceTypeDescriptor) bool {
	return f(descriptor)
}

func HasKDSFlag(flagType KDSFlagType) TypeFilter {
	return TypeFilterFn(func(descriptor ResourceTypeDescriptor) bool {
		return descriptor.KDSFlags.Has(flagType)
	})
}

func HasKdsEnabled() TypeFilter {
	return TypeFilterFn(func(descriptor ResourceTypeDescriptor) bool {
		return descriptor.KDSFlags != 0
	})
}

func HasKumactlEnabled() TypeFilter {
	return TypeFilterFn(func(descriptor ResourceTypeDescriptor) bool {
		return descriptor.KumactlArg != ""
	})
}

func HasWsEnabled() TypeFilter {
	return TypeFilterFn(func(descriptor ResourceTypeDescriptor) bool {
		return descriptor.WsPath != ""
	})
}

func AllowedToInspect() TypeFilter {
	return TypeFilterFn(func(descriptor ResourceTypeDescriptor) bool {
		return descriptor.AllowToInspect
	})
}

func HasScope(scope ResourceScope) TypeFilter {
	return TypeFilterFn(func(descriptor ResourceTypeDescriptor) bool {
		return descriptor.Scope == scope
	})
}

func IsPolicy() TypeFilter {
	return TypeFilterFn(func(descriptor ResourceTypeDescriptor) bool {
		return descriptor.IsPolicy
	})
}

func Named(names ...ResourceType) TypeFilter {
	included := map[ResourceType]bool{}
	for _, n := range names {
		included[n] = true
	}

	return TypeFilterFn(func(descriptor ResourceTypeDescriptor) bool {
		return included[descriptor.Name]
	})
}

func Not(filter TypeFilter) TypeFilter {
	return TypeFilterFn(func(descriptor ResourceTypeDescriptor) bool {
		return !filter.Apply(descriptor)
	})
}

func Or(filters ...TypeFilter) TypeFilter {
	return TypeFilterFn(func(descriptor ResourceTypeDescriptor) bool {
		for _, filter := range filters {
			if filter.Apply(descriptor) {
				return true
			}
		}

		return false
	})
}

type ByMeta []Resource

func (a ByMeta) Len() int { return len(a) }

func (a ByMeta) Less(i, j int) bool {
	if a[i].GetMeta().GetMesh() == a[j].GetMeta().GetMesh() {
		return a[i].GetMeta().GetName() < a[j].GetMeta().GetName()
	}
	return a[i].GetMeta().GetMesh() < a[j].GetMeta().GetMesh()
}

func (a ByMeta) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

type ResourceType string

// ResourceNameExtensions represents an composite resource name in environments
// other than Universal.
//
// E.g., name of a Kubernetes resource consists of a namespace component
// and a name component that is local to that namespace.
//
// Technically, ResourceNameExtensions is a mapping between
// a component identifier and a component value, e.g.
//
//	"k8s.kuma.io/namespace" => "my-namespace"
//	"k8s.kuma.io/name"      => "my-policy"
//
// Component identifier must be considered a part of user-facing Kuma API.
// In other words, it is supposed to be visible to users and should not be changed lightly.
//
// Component identifier might have any value, however, it's preferable
// to choose one that is intuitive to users of that particular environment.
// E.g., on Kubernetes component identifiers should use a label name format,
// like in "k8s.kuma.io/namespace" and "k8s.kuma.io/name".
type ResourceNameExtensions map[string]string

type ResourceMeta interface {
	GetName() string
	GetNameExtensions() ResourceNameExtensions
	GetVersion() string
	GetMesh() string
	GetCreationTime() time.Time
	GetModificationTime() time.Time
}

func MetaToResourceKey(meta ResourceMeta) ResourceKey {
	if meta == nil {
		return ResourceKey{}
	}
	return ResourceKey{
		Mesh: meta.GetMesh(),
		Name: meta.GetName(),
	}
}

func ResourceListToResourceKeys(rl ResourceList) []ResourceKey {
	rkey := []ResourceKey{}
	for _, r := range rl.GetItems() {
		rkey = append(rkey, MetaToResourceKey(r.GetMeta()))
	}
	return rkey
}

type ResourceList interface {
	GetItemType() ResourceType
	GetItems() []Resource
	NewItem() Resource
	AddItem(Resource) error
	GetPagination() *Pagination
}

type Pagination struct {
	Total      uint32
	NextOffset string
}

func (p *Pagination) GetTotal() uint32 {
	return p.Total
}

func (p *Pagination) SetTotal(total uint32) {
	p.Total = total
}

func (p *Pagination) GetNextOffset() string {
	return p.NextOffset
}

func (p *Pagination) SetNextOffset(nextOffset string) {
	p.NextOffset = nextOffset
}

func ErrorInvalidItemType(expected, actual interface{}) error {
	return fmt.Errorf("Invalid argument type: expected=%q got=%q", reflect.TypeOf(expected), reflect.TypeOf(actual))
}

type ResourceWithAddress interface {
	Resource
	AdminAddress(defaultAdminPort uint32) string
}

// ZoneOfResource returns zone from which the resource was synced to Global CP
// There is no information in the resource itself whether the resource is synced or created on the CP.
// Therefore, it's a caller responsibility to make use it only on synced resources.
func ZoneOfResource(res Resource) string {
	parts := strings.Split(res.GetMeta().GetName(), ".")
	return parts[0]
}

type PolicyItem interface {
	GetTargetRef() common_api.TargetRef
	GetDefault() interface{}
}

type Policy interface {
	ResourceSpec
	GetTargetRef() common_api.TargetRef
}

type PolicyWithToList interface {
	Policy
	GetToList() []PolicyItem
}

type PolicyWithFromList interface {
	Policy
	GetFromList() []PolicyItem
}

type PolicyWithSingleItem interface {
	Policy
	GetPolicyItem() PolicyItem
}
