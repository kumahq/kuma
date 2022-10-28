// +kubebuilder:object:generate=true
package v1alpha1

type TargetRefKind string

var (
	Mesh              TargetRefKind = "Mesh"
	MeshSubset        TargetRefKind = "MeshSubset"
	MeshService       TargetRefKind = "MeshService"
	MeshServiceSubset TargetRefKind = "MeshServiceSubset"
	MeshGatewayRoute  TargetRefKind = "MeshGatewayRoute"
)

var order = map[TargetRefKind]int{
	Mesh:              1,
	MeshSubset:        2,
	MeshService:       3,
	MeshServiceSubset: 4,
	MeshGatewayRoute:  5,
}

func (k TargetRefKind) Less(o TargetRefKind) bool {
	return order[k] < order[o]
}

// TargetRef defines structure that allows attaching policy to various objects
type TargetRef struct {
	// Kind of the referenced resource
	// +kubebuilder:validation:Enum=Mesh;MeshSubset;MeshService;MeshServiceSubset;MeshGatewayRoute
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
