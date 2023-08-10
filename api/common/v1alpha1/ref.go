// +kubebuilder:object:generate=true
package v1alpha1

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/exp/maps"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
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

func (k TargetRefKind) Less(o TargetRefKind) bool {
	return order[k] < order[o]
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
}

type TargetRefHash string

// Hash returns a hash of the TargetRef
func (in TargetRef) Hash() TargetRefHash {
	keys := maps.Keys(in.Tags)
	sort.Strings(keys)
	orderedTags := make([]string, len(keys))
	for _, k := range keys {
		orderedTags = append(orderedTags, fmt.Sprintf("%s=%s", k, in.Tags[k]))
	}
	return TargetRefHash(fmt.Sprintf("%s/%s/%s/%s", in.Kind, in.Name, strings.Join(orderedTags, "/"), in.Mesh))
}

func (in TargetRef) EnvoyTags() (map[string]string, bool) {
	var service string
	tags := map[string]string{}

	switch in.Kind {
	case MeshService:
		service = in.Name
	case MeshServiceSubset:
		service = in.Name
		tags = in.Tags
	case Mesh:
		service = mesh_proto.MatchAllTag
	case MeshSubset:
		service = mesh_proto.MatchAllTag
		tags = in.Tags
	default:
		return nil, false
	}

	tags[mesh_proto.ServiceTag] = service
	return tags, true
}

// BackendRef defines where to forward traffic.
type BackendRef struct {
	TargetRef `json:","`
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=1
	Weight *uint `json:"weight,omitempty"`
}
