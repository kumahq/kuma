package model

import (
	"fmt"
	"hash/fnv"
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"
	"k8s.io/kube-openapi/pkg/validation/spec"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/pkg/util/pointer"
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
	// KDSDisabledFlag is a flag that indicates that this resource type is not sent using KDS.
	KDSDisabledFlag = KDSFlagType(0)

	// ZoneToGlobalFlag is a flag that indicates that this resource type is sent from Zone CP to Global CP.
	ZoneToGlobalFlag = KDSFlagType(1)

	// GlobalToAllZonesFlag is a flag that indicates that this resource type is sent from Global CP to all zones.
	GlobalToAllZonesFlag = KDSFlagType(1 << 2)

	// GlobalToAllButOriginalZoneFlag is a flag that indicates that this resource type is sent from Global CP to
	// all zones except the zone where the resource was originally created. Today the only resource that has this
	// flag is ZoneIngress.
	GlobalToAllButOriginalZoneFlag = KDSFlagType(1 << 3)
)

const (
	// GlobalToZoneSelector is selector for all flags that indicate resource sync from Global to Zone.
	// Can't be used as KDS flag for resource type.
	GlobalToZoneSelector = GlobalToAllZonesFlag | GlobalToAllButOriginalZoneFlag

	// AllowedOnGlobalSelector is selector for all flags that indicate resource can be created on Global.
	AllowedOnGlobalSelector = GlobalToAllZonesFlag

	// AllowedOnZoneSelector is selector for all flags that indicate resource can be created on Zone.
	AllowedOnZoneSelector = ZoneToGlobalFlag | GlobalToAllButOriginalZoneFlag
)

// Has return whether this flag has all the passed flags on.
func (kt KDSFlagType) Has(flag KDSFlagType) bool {
	return kt&flag != 0
}

type ResourceSpec interface{}

type ResourceStatus interface{}

type Resource interface {
	GetMeta() ResourceMeta
	SetMeta(ResourceMeta)
	GetSpec() ResourceSpec
	SetSpec(ResourceSpec) error
	GetStatus() ResourceStatus
	SetStatus(ResourceStatus) error
	Descriptor() ResourceTypeDescriptor
}

type ResourceHasher interface {
	Hash() []byte
}

func Hash(resource Resource) []byte {
	if r, ok := resource.(ResourceHasher); ok {
		return r.Hash()
	}
	return HashMeta(resource)
}

func HashMeta(r Resource) []byte {
	meta := r.GetMeta()
	hasher := fnv.New128a()
	_, _ = hasher.Write([]byte(r.Descriptor().Name))
	_, _ = hasher.Write([]byte(meta.GetMesh()))
	_, _ = hasher.Write([]byte(meta.GetName()))
	_, _ = hasher.Write([]byte(meta.GetVersion()))
	return hasher.Sum(nil)
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

func Deprecations(resource Resource) []string {
	if v, ok := interface{}(resource).(interface{ Deprecations() []string }); ok {
		return v.Deprecations()
	}
	return nil
}

type OverviewResource interface {
	SetOverviewSpec(resource Resource, insight Resource) error
}

type ResourceWithInsights interface {
	NewInsightList() ResourceList
	NewOverviewList() ResourceList
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
	// HasStatus indicates that the policy has a status field
	HasStatus bool
	// Schema contains an unmarshalled OpenAPI schema of the resource
	Schema *spec.Schema
	// Insight contains the insight type attached to this resourceType
	Insight Resource
	// Overview contains the overview type attached to this resourceType
	Overview Resource
	// DumpForGlobal whether resources of this type should be dumped when exporting a zone to migrate to global
	DumpForGlobal bool
	// AllowedOnSystemNamespaceOnly whether this resource type can be created only in the system namespace
	AllowedOnSystemNamespaceOnly bool
	// IsReferenceableInTo whether this resource type can be used in spec.to[].targetRef
	IsReferenceableInTo bool
}

func newObject(baseResource Resource) Resource {
	specType := reflect.TypeOf(baseResource.GetSpec()).Elem()
	newSpec := reflect.New(specType).Interface().(ResourceSpec)

	resType := reflect.TypeOf(baseResource).Elem()
	resource := reflect.New(resType).Interface().(Resource)

	if err := resource.SetSpec(newSpec); err != nil {
		panic(errors.Wrap(err, "could not set spec on the new resource"))
	}

	if baseResource.Descriptor().HasStatus {
		statusType := reflect.TypeOf(baseResource.GetStatus()).Elem()
		newStatus := reflect.New(statusType).Interface().(ResourceSpec)
		if err := resource.SetStatus(newStatus); err != nil {
			panic(errors.Wrap(err, "could not set spec on the new resource"))
		}
	}

	return resource
}

func (d ResourceTypeDescriptor) NewObject() Resource {
	return newObject(d.Resource)
}

func (d ResourceTypeDescriptor) NewList() ResourceList {
	listType := reflect.TypeOf(d.ResourceList).Elem()
	return reflect.New(listType).Interface().(ResourceList)
}

func (d ResourceTypeDescriptor) HasInsights() bool {
	return d.Insight != nil
}

func (d ResourceTypeDescriptor) NewInsight() Resource {
	if !d.HasInsights() {
		panic("No insight type precondition broken")
	}
	return newObject(d.Insight)
}

func (d ResourceTypeDescriptor) NewInsightList() ResourceList {
	if !d.HasInsights() {
		panic("No insight type precondition broken")
	}
	return d.Insight.Descriptor().NewList()
}

func (d ResourceTypeDescriptor) NewOverview() Resource {
	if !d.HasInsights() {
		panic("No insight type precondition broken")
	}
	return newObject(d.Overview)
}

func (d ResourceTypeDescriptor) NewOverviewList() ResourceList {
	if !d.HasInsights() {
		panic("No insight type precondition broken")
	}
	return d.Overview.Descriptor().NewList()
}

func (d ResourceTypeDescriptor) IsInsight() bool {
	return strings.HasSuffix(string(d.Name), "Insight")
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
		return descriptor.KDSFlags != KDSDisabledFlag
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

func IsInsight() TypeFilter {
	return TypeFilterFn(func(descriptor ResourceTypeDescriptor) bool {
		return descriptor.IsInsight()
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

const (
	// K8sNamespaceComponent identifies the namespace component of a resource name on Kubernetes.
	// The value is considered a part of user-facing Kuma API and should not be changed lightly.
	// The value has a format of a Kubernetes label name.
	K8sNamespaceComponent = "k8s.kuma.io/namespace"

	// K8sNameComponent identifies the name component of a resource name on Kubernetes.
	// The value is considered a part of user-facing Kuma API and should not be changed lightly.
	// The value has a format of a Kubernetes label name.
	K8sNameComponent = "k8s.kuma.io/name"
)

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
	GetLabels() map[string]string
}

// IsReferenced check if `refMeta` references with `refName` the entity `resourceMeta`
// This is required because in multi-zone policies may have names different from the name they are defined with.
func IsReferenced(refMeta ResourceMeta, refName string, resourceMeta ResourceMeta) bool {
	if refMeta.GetMesh() != resourceMeta.GetMesh() {
		return false
	}
	return refName == GetDisplayName(resourceMeta)
}

func IsLocallyOriginated(mode config_core.CpMode, labels map[string]string) bool {
	switch mode {
	case config_core.Global:
		origin, ok := resourceOrigin(labels)
		return !ok || origin == mesh_proto.GlobalResourceOrigin
	case config_core.Zone:
		origin, ok := resourceOrigin(labels)
		return !ok || origin == mesh_proto.ZoneResourceOrigin
	default:
		return true
	}
}

func GetDisplayName(rm ResourceMeta) string {
	// prefer display name as it's more predictable, because
	// * Kubernetes expects sorting to be by just a name. Considering suffix with namespace breaks this
	// * When policies are synced to Zone, hash suffix also breaks sorting
	if labels := rm.GetLabels(); labels != nil && labels[mesh_proto.DisplayName] != "" {
		return labels[mesh_proto.DisplayName]
	}
	return rm.GetName()
}

func ResourceOrigin(rm ResourceMeta) (mesh_proto.ResourceOrigin, bool) {
	if rm == nil {
		return "", false
	}
	return resourceOrigin(rm.GetLabels())
}

func resourceOrigin(labels map[string]string) (mesh_proto.ResourceOrigin, bool) {
	if labels != nil && labels[mesh_proto.ResourceOriginLabel] != "" {
		return mesh_proto.ResourceOrigin(labels[mesh_proto.ResourceOriginLabel]), true
	}
	return "", false
}

// Namespace type allows to avoid carrying both 'namespace' and 'systemNamespace' around the code base
// and depend on this type instead
type Namespace struct {
	value  string
	system bool
}

var UnsetNamespace = Namespace{}

func NewNamespace(value string, system bool) Namespace {
	return Namespace{
		value:  value,
		system: system,
	}
}

func GetNamespace(rm ResourceMeta, systemNamespace string) Namespace {
	if ns, ok := rm.GetNameExtensions()[K8sNamespaceComponent]; ok && ns != "" {
		return Namespace{
			value:  ns,
			system: ns == systemNamespace,
		}
	}
	return UnsetNamespace
}

func ComputeLabels(
	rd ResourceTypeDescriptor,
	spec ResourceSpec,
	existingLabels map[string]string,
	ns Namespace,
	mesh string,
	mode config_core.CpMode,
	isK8s bool,
	localZone string,
) (map[string]string, error) {
	labels := map[string]string{}
	if len(existingLabels) > 0 {
		labels = existingLabels
	}

	setIfNotExist := func(k, v string) {
		if _, ok := labels[k]; !ok {
			labels[k] = v
		}
	}

	getMeshOrDefault := func() string {
		if mesh != "" {
			return mesh
		}
		return DefaultMesh
	}

	if rd.Scope == ScopeMesh {
		setIfNotExist(metadata.KumaMeshLabel, getMeshOrDefault())
	}

	if mode == config_core.Zone {
		// If resource can't be created on Zone (like Mesh), there is no point in adding
		// 'kuma.io/zone', 'kuma.io/origin' and 'kuma.io/env' labels even if the zone is non-federated
		if rd.KDSFlags.Has(AllowedOnZoneSelector) {
			setIfNotExist(mesh_proto.ResourceOriginLabel, string(mesh_proto.ZoneResourceOrigin))
			if labels[mesh_proto.ResourceOriginLabel] != string(mesh_proto.GlobalResourceOrigin) {
				setIfNotExist(mesh_proto.ZoneTag, localZone)
				env := mesh_proto.UniversalEnvironment
				if isK8s {
					env = mesh_proto.KubernetesEnvironment
				}
				setIfNotExist(mesh_proto.EnvTag, env)
			}
		}
	}

	if ns.value != "" && isK8s && IsLocallyOriginated(mode, labels) {
		setIfNotExist(mesh_proto.KubeNamespaceTag, ns.value)
	}

	if ns.value != "" && rd.IsPolicy && rd.IsPluginOriginated && IsLocallyOriginated(mode, labels) {
		role, err := ComputePolicyRole(spec.(Policy), ns)
		if err != nil {
			return nil, err
		}
		labels[mesh_proto.PolicyRoleLabel] = string(role)
	}

	return labels, nil
}

func ComputePolicyRole(p Policy, ns Namespace) (mesh_proto.PolicyRole, error) {
	if ns.system || ns == UnsetNamespace {
		// on Universal the value is always empty
		return mesh_proto.SystemPolicyRole, nil
	}

	hasTo := false
	if pwtl, ok := p.(PolicyWithToList); ok && len(pwtl.GetToList()) > 0 {
		hasTo = true
	}

	hasFrom := false
	if pwfl, ok := p.(PolicyWithFromList); ok && len(pwfl.GetFromList()) > 0 {
		hasFrom = true
	}

	if hasFrom && hasTo {
		return "", errors.New("it's not allowed to mix 'to' and 'from' arrays in the same policy")
	}

	if hasFrom || !(hasTo || hasFrom) {
		// if there is 'from' or neither (single item)
		return mesh_proto.WorkloadOwnerPolicyRole, nil
	}

	isProducerItem := func(tr common_api.TargetRef) bool {
		return tr.Kind == common_api.MeshService && tr.Name != "" && (tr.Namespace == "" || tr.Namespace == ns.value)
	}

	producerItems := 0
	for _, item := range p.(PolicyWithToList).GetToList() {
		if isProducerItem(item.GetTargetRef()) {
			producerItems++
		}
	}

	switch {
	case producerItems == len(p.(PolicyWithToList).GetToList()):
		return mesh_proto.ProducerPolicyRole, nil
	case producerItems == 0:
		return mesh_proto.ConsumerPolicyRole, nil
	default:
		return "", errors.New("it's not allowed to mix producer and consumer items in the same policy")
	}
}

func PolicyRole(rm ResourceMeta) mesh_proto.PolicyRole {
	if rm == nil || rm.GetLabels() == nil || rm.GetLabels()[mesh_proto.PolicyRoleLabel] == "" {
		return mesh_proto.SystemPolicyRole
	}
	return mesh_proto.PolicyRole(rm.GetLabels()[mesh_proto.PolicyRoleLabel])
}

// ZoneOfResource returns zone from which the resource was synced to Global CP
// There is no information in the resource itself whether the resource is synced or created on the CP.
// Therefore, it's a caller responsibility to make use it only on synced resources.
func ZoneOfResource(res Resource) string {
	if labels := res.GetMeta().GetLabels(); labels != nil && labels[mesh_proto.ZoneTag] != "" {
		return labels[mesh_proto.ZoneTag]
	}
	parts := strings.Split(res.GetMeta().GetName(), ".")
	return parts[0]
}

func IsShadowedResource(r Resource) bool {
	if labels := r.GetMeta().GetLabels(); labels != nil && labels[mesh_proto.EffectLabel] == "shadow" {
		return true
	}
	return false
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

func ResourceListByMesh(rl ResourceList) (map[string]ResourceList, error) {
	res := map[string]ResourceList{}
	for _, r := range rl.GetItems() {
		mrl, ok := res[r.GetMeta().GetMesh()]
		if !ok {
			mrl = r.Descriptor().NewList()
			res[r.GetMeta().GetMesh()] = mrl
		}
		if err := mrl.AddItem(r); err != nil {
			return nil, err
		}
	}
	return res, nil
}

func ResourceListHash(rl ResourceList) []byte {
	hasher := fnv.New128()
	for _, entity := range rl.GetItems() {
		_, _ = hasher.Write(Hash(entity))
	}
	return hasher.Sum(nil)
}

type ResourceList interface {
	GetItemType() ResourceType
	GetItems() []Resource
	NewItem() Resource
	AddItem(Resource) error
	GetPagination() *Pagination
	SetPagination(pagination Pagination)
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

type PolicyItem interface {
	GetTargetRef() common_api.TargetRef
	GetDefault() interface{}
}

type TransformDefaultAfterMerge interface {
	Transform()
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

func IndexByKey[T Resource](resources []T) map[ResourceKey]T {
	indexedResources := make(map[ResourceKey]T)
	for _, resource := range resources {
		key := MetaToResourceKey(resource.GetMeta())
		indexedResources[key] = resource
	}
	return indexedResources
}

// Resource can implement defaulter to provide static default fields.
// Kubernetes Webhook and Resource Manager will make sure that Default() is called before Create/Update
type Defaulter interface {
	Resource
	Default() error
}

type ResourceIdentifier struct {
	Name      string
	Mesh      string
	Namespace string
	Zone      string
}

func NewResourceIdentifier(r Resource) ResourceIdentifier {
	return ResourceIdentifier{
		Name:      GetDisplayName(r.GetMeta()),
		Mesh:      r.GetMeta().GetMesh(),
		Namespace: r.GetMeta().GetLabels()[mesh_proto.KubeNamespaceTag],
		Zone:      r.GetMeta().GetLabels()[mesh_proto.ZoneTag],
	}
}

func TargetRefToResourceIdentifier(meta ResourceMeta, tr common_api.TargetRef) ResourceIdentifier {
	switch tr.Kind {
	case common_api.Mesh:
		return ResourceIdentifier{
			Name: meta.GetMesh(),
		}
	default:
		var namespace string
		if tr.Namespace != "" {
			namespace = tr.Namespace
		} else {
			namespace = meta.GetLabels()[mesh_proto.KubeNamespaceTag]
		}
		return ResourceIdentifier{
			Mesh:      meta.GetMesh(),
			Zone:      meta.GetLabels()[mesh_proto.ZoneTag],
			Namespace: namespace,
			Name:      tr.Name,
		}
	}
}

func ResourceToBackendRef(r Resource, resType ResourceType, port uint32) common_api.BackendRef {
	id := NewResourceIdentifier(r)
	return common_api.BackendRef{
		TargetRef: common_api.TargetRef{
			Kind:      common_api.TargetRefKind(resType),
			Name:      id.Name,
			Namespace: id.Namespace,
		},
		Port: pointer.To(port),
	}
}

type LabelResourceIdentifierResolver func(ResourceType, map[string]string) *ResourceIdentifier

func ResolveBackendRef(meta ResourceMeta, br common_api.BackendRef, resolver LabelResourceIdentifierResolver) *ResolvedBackendRef {
	switch {
	case br.Kind == common_api.MeshService && br.ReferencesRealObject():
	case br.Kind == common_api.MeshExternalService:
	case br.Kind == common_api.MeshMultiZoneService:
	default:
		return &ResolvedBackendRef{Ref: pointer.To(LegacyBackendRef(br))}
	}

	rr := RealResourceBackendRef{
		Resource: &TypedResourceIdentifier{
			ResourceIdentifier: TargetRefToResourceIdentifier(meta, br.TargetRef),
			ResourceType:       ResourceType(br.Kind),
		},
		Weight: pointer.DerefOr(br.Weight, 1),
	}

	if len(br.Labels) > 0 {
		ri := resolver(ResourceType(br.Kind), br.Labels)
		if ri == nil {
			return nil
		}
		rr.Resource.ResourceIdentifier = *ri
	}

	if br.Port != nil {
		rr.Resource.SectionName = fmt.Sprintf("%d", *br.Port)
	}

	return &ResolvedBackendRef{Ref: &rr}
}

func (r ResourceIdentifier) String() string {
	var pairs []string
	if r.Mesh != "" {
		pairs = append(pairs, fmt.Sprintf("mesh/%s", r.Mesh))
	}
	if r.Zone != "" {
		pairs = append(pairs, fmt.Sprintf("zone/%s", r.Zone))
	}
	if r.Namespace != "" {
		pairs = append(pairs, fmt.Sprintf("namespace/%s", r.Namespace))
	}
	if r.Name != "" {
		pairs = append(pairs, fmt.Sprintf("name/%s", r.Name))
	}
	return strings.Join(pairs, ":")
}

type TypedResourceIdentifier struct {
	ResourceIdentifier

	ResourceType ResourceType
	SectionName  string
}

type IsResolvedBackendRef interface {
	isResolvedBackendRef()
}

type ResolvedBackendRef struct {
	// Ref is either LegacyBackendRef or RealResourceBackendRef
	Ref IsResolvedBackendRef
}

func NewResolvedBackendRef(r IsResolvedBackendRef) *ResolvedBackendRef {
	return &ResolvedBackendRef{Ref: r}
}

func (rbr *ResolvedBackendRef) ReferencesRealResource() bool {
	if rbr == nil {
		return false
	}
	if rbr.Ref == nil {
		return false
	}
	_, ok := rbr.Ref.(*RealResourceBackendRef)
	return ok
}

func (rbr *ResolvedBackendRef) ResourceOrNil() *TypedResourceIdentifier {
	if rr := rbr.RealResourceBackendRef(); rr != nil {
		return rr.Resource
	}
	return nil
}

func (rbr *ResolvedBackendRef) LegacyBackendRef() *LegacyBackendRef {
	if lbr, ok := rbr.Ref.(*LegacyBackendRef); ok {
		return lbr
	}
	return nil
}

func (rbr *ResolvedBackendRef) RealResourceBackendRef() *RealResourceBackendRef {
	if rr, ok := rbr.Ref.(*RealResourceBackendRef); ok {
		return rr
	}
	return nil
}

type LegacyBackendRef common_api.BackendRef

func (lbr *LegacyBackendRef) isResolvedBackendRef() {}

type RealResourceBackendRef struct {
	Resource *TypedResourceIdentifier
	Weight   uint
}

func (rbr *RealResourceBackendRef) isResolvedBackendRef() {}

type NewTypedResourceIdentifierFunc func(id *TypedResourceIdentifier)

func WithSectionName(sectionName string) NewTypedResourceIdentifierFunc {
	return func(id *TypedResourceIdentifier) {
		id.SectionName = sectionName
	}
}

func NewTypedResourceIdentifier(r Resource, opts ...NewTypedResourceIdentifierFunc) TypedResourceIdentifier {
	tri := TypedResourceIdentifier{
		ResourceType:       r.Descriptor().Name,
		ResourceIdentifier: NewResourceIdentifier(r),
	}
	for _, opt := range opts {
		opt(&tri)
	}
	return tri
}

func (ri TypedResourceIdentifier) MarshalText() ([]byte, error) {
	return []byte(ri.String()), nil
}

func (ri TypedResourceIdentifier) String() string {
	var pairs []string
	if ri.ResourceType != "" {
		pairs = append(pairs, strings.ToLower(string(ri.ResourceType)))
	}
	pairs = append(pairs, ri.ResourceIdentifier.String())
	if ri.SectionName != "" {
		pairs = append(pairs, fmt.Sprintf("section/%s", ri.SectionName))
	}
	return strings.Join(pairs, ":")
}
