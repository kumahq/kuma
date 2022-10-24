// +kubebuilder:object:generate=true
package v1alpha1

type TargetRefKind string

var Mesh TargetRefKind = "Mesh"
var MeshSubset TargetRefKind = "MeshSubset"
var MeshService TargetRefKind = "MeshService"
var MeshServiceSubset TargetRefKind = "MeshServiceSubset"
var MeshGatewayRoute TargetRefKind = "MeshGatewayRoute"
var MeshHTTPRoute TargetRefKind = "MeshHTTPRoute"

var order = map[TargetRefKind]int{
	Mesh:              1,
	MeshSubset:        2,
	MeshService:       3,
	MeshServiceSubset: 4,
	MeshGatewayRoute:  5,
	MeshHTTPRoute:     6,
}

func (k TargetRefKind) Less(o TargetRefKind) bool {
	return order[k] < order[o]
}

// TargetRef defines structure that allows attaching policy to various objects
type TargetRef struct {
	// Kind of the referenced resource
	// +kubebuilder:validation:Enum=Mesh;MeshSubset;MeshService;MeshServiceSubset;MeshGatewayRoute;MeshHTTPRoute
	Kind TargetRefKind `json:"kind,omitempty"`
	// Name of the referenced resource
	Name string `json:"name,omitempty"`
	// Tags are used with MeshSubset and MeshServiceSubset to define a subset of
	// proxies
	Tags map[string]string `json:"tags,omitempty"`
	// Mesh is used with MeshService and MeshServiceSubset to identify the service
	// from another mesh. Could be useful when implementing policies with
	// cross-mesh support.
	Mesh string `json:"mesh,omitempty"`
}
