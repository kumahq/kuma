// +kubebuilder:object:generate=true
package v1alpha1

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"golang.org/x/exp/maps"
)

type TargetRefKind string

var (
	Mesh              TargetRefKind = "Mesh"
	MeshSubset        TargetRefKind = "MeshSubset"
	MeshGateway       TargetRefKind = "MeshGateway"
	MeshService       TargetRefKind = "MeshService"
	MeshServiceSubset TargetRefKind = "MeshServiceSubset"
	MeshHTTPRoute     TargetRefKind = "MeshHTTPRoute"
)

var order = map[TargetRefKind]int{
	Mesh:              1,
	MeshSubset:        2,
	MeshGateway:       3,
	MeshService:       4,
	MeshServiceSubset: 5,
	MeshHTTPRoute:     6,
}

// +kubebuilder:validation:Enum=Sidecar;Gateway
type TargetRefProxyType string

var (
	Sidecar TargetRefProxyType = "Sidecar"
	Gateway TargetRefProxyType = "Gateway"
)

func (k TargetRefKind) Less(o TargetRefKind) bool {
	return order[k] < order[o]
}

func AllTargetRefKinds() []TargetRefKind {
	return maps.Keys(order)
}

// TargetRef defines structure that allows attaching policy to various objects
type TargetRef struct {
	// Kind of the referenced resource
	// +kubebuilder:validation:Enum=Mesh;MeshSubset;MeshGateway;MeshService;MeshServiceSubset;MeshHTTPRoute
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

type BackendRefHash string

// Hash returns a hash of the BackendRef
func (in BackendRef) Hash() BackendRefHash {
	keys := maps.Keys(in.Tags)
	sort.Strings(keys)
	orderedTags := make([]string, len(keys))
	for _, k := range keys {
		orderedTags = append(orderedTags, fmt.Sprintf("%s=%s", k, in.Tags[k]))
	}
	name := in.Name
	if in.Port != nil {
		name = fmt.Sprintf("%s_svc_%d", in.Name, *in.Port)
	}
	return BackendRefHash(fmt.Sprintf("%s/%s/%s/%s", in.Kind, name, strings.Join(orderedTags, "/"), in.Mesh))
}
