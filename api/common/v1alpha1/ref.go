// +kubebuilder:object:generate=true
package v1alpha1

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	util_maps "github.com/kumahq/kuma/pkg/util/maps"
)

type TargetRefKind string

var (
	Mesh                 TargetRefKind = "Mesh"
	MeshSubset           TargetRefKind = "MeshSubset"
	MeshGateway          TargetRefKind = "MeshGateway"
	MeshService          TargetRefKind = "MeshService"
	MeshExternalService  TargetRefKind = "MeshExternalService"
	MeshMultiZoneService TargetRefKind = "MeshMultiZoneService"
	MeshServiceSubset    TargetRefKind = "MeshServiceSubset"
	MeshHTTPRoute        TargetRefKind = "MeshHTTPRoute"
)

var order = map[TargetRefKind]int{
	Mesh:                 1,
	MeshSubset:           2,
	MeshGateway:          3,
	MeshService:          4,
	MeshExternalService:  5,
	MeshMultiZoneService: 6,
	MeshServiceSubset:    7,
	MeshHTTPRoute:        8,
}

// +kubebuilder:validation:Enum=Sidecar;Gateway
type TargetRefProxyType string

var (
	Sidecar     TargetRefProxyType = "Sidecar"
	Gateway     TargetRefProxyType = "Gateway"
	ZoneIngress TargetRefProxyType = "ZoneIngress"
	ZoneEgress  TargetRefProxyType = "ZoneEgress"
)

func (k TargetRefKind) Compare(o TargetRefKind) int {
	return order[k] - order[o]
}

func (k TargetRefKind) IsRealResource() bool {
	switch k {
	case MeshSubset, MeshServiceSubset:
		return false
	default:
		return true
	}
}

// These are the kinds that can be used in Kuma policies before support for
// actual resources (e.g., MeshExternalService, MeshMultiZoneService, and MeshService) was introduced.
func (k TargetRefKind) IsOldKind() bool {
	switch k {
	case Mesh, MeshSubset, MeshServiceSubset, MeshService, MeshGateway, MeshHTTPRoute:
		return true
	default:
		return false
	}
}

func AllTargetRefKinds() []TargetRefKind {
	keys := util_maps.AllKeys(order)
	sort.Sort(TargetRefKindSlice(keys))
	return keys
}

type TargetRefKindSlice []TargetRefKind

func (x TargetRefKindSlice) Len() int           { return len(x) }
func (x TargetRefKindSlice) Less(i, j int) bool { return string(x[i]) < string(x[j]) }
func (x TargetRefKindSlice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

// TargetRef defines structure that allows attaching policy to various objects
type TargetRef struct {
	// This is needed to not sync policies with empty topLevelTarget ref to old zones that does not support it
	// This can be removed in 2.11.x
	UsesSyntacticSugar bool `json:"-"`

	// Kind of the referenced resource
	// +kubebuilder:validation:Enum=Mesh;MeshSubset;MeshGateway;MeshService;MeshExternalService;MeshMultiZoneService;MeshServiceSubset;MeshHTTPRoute
	Kind TargetRefKind `json:"kind,omitempty"`
	// Name of the referenced resource. Can only be used with kinds: `MeshService`,
	// `MeshServiceSubset` and `MeshGatewayRoute`
	Name string `json:"name,omitempty"`
	// Tags used to select a subset of proxies by tags. Can only be used with kinds
	// `MeshSubset` and `MeshServiceSubset`
	Tags map[string]string `json:"tags,omitempty"`
	// Mesh is reserved for future use to identify cross mesh resources.
	Mesh string `json:"mesh,omitempty"`
	// ProxyTypes specifies the data plane types that are subject to the policy. When not specified,
	// all data plane types are targeted by the policy.
	// +kubebuilder:validation:MinItems=1
	ProxyTypes []TargetRefProxyType `json:"proxyTypes,omitempty"`
	// Namespace specifies the namespace of target resource. If empty only resources in policy namespace
	// will be targeted.
	Namespace string `json:"namespace,omitempty"`
	// Labels are used to select group of MeshServices that match labels. Either Labels or
	// Name and Namespace can be used.
	Labels map[string]string `json:"labels,omitempty"`
	// SectionName is used to target specific section of resource.
	// For example, you can target port from MeshService.ports[] by its name. Only traffic to this port will be affected.
	SectionName string `json:"sectionName,omitempty"`
}

func IncludesGateways(ref TargetRef) bool {
	isGateway := ref.Kind == MeshGateway
	isMeshKind := ref.Kind == Mesh || ref.Kind == MeshSubset
	isGatewayInProxyTypes := len(ref.ProxyTypes) == 0 || slices.Contains(ref.ProxyTypes, Gateway)
	isGatewayCompatible := isMeshKind && isGatewayInProxyTypes
	isMeshHTTPRoute := ref.Kind == MeshHTTPRoute

	return isGateway || isGatewayCompatible || isMeshHTTPRoute
}

// BackendRef defines where to forward traffic.
type BackendRef struct {
	TargetRef `json:","`
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=1
	Weight *uint `json:"weight,omitempty"`
	// Port is only supported when this ref refers to a real MeshService object
	Port *uint32 `json:"port,omitempty"`
}

func (b BackendRef) ReferencesRealObject() bool {
	switch b.Kind {
	case MeshService:
		return b.Port != nil
	case MeshServiceSubset:
		return false
	// empty targetRef should not be treated as real object
	case "":
		return false
	default:
		return true
	}
}

// MatchesHash is used to hash route matches to determine the origin resource
// for a ref
type MatchesHash string

type BackendRefHash string

// Hash returns a hash of the BackendRef
func (in BackendRef) Hash() BackendRefHash {
	keys := util_maps.SortedKeys(in.Tags)
	orderedTags := make([]string, 0, len(keys))
	for _, k := range keys {
		orderedTags = append(orderedTags, fmt.Sprintf("%s=%s", k, in.Tags[k]))
	}

	keys = util_maps.SortedKeys(in.Labels)
	orderedLabels := make([]string, 0, len(in.Labels))
	for _, k := range keys {
		orderedLabels = append(orderedLabels, fmt.Sprintf("%s=%s", k, in.Labels[k]))
	}

	name := in.Name
	if in.Port != nil {
		name = fmt.Sprintf("%s_svc_%d", in.Name, *in.Port)
	}
	return BackendRefHash(fmt.Sprintf("%s/%s/%s/%s/%s", in.Kind, name, strings.Join(orderedTags, "/"), strings.Join(orderedLabels, "/"), in.Mesh))
}
