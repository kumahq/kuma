package metadata

const (
	// KumaMeshLabel defines a Pod label to associate objects
	// with a particular Mesh.
	// Label value must be the name of a Mesh resource.
	KumaMeshLabel = "kuma.io/mesh"

	// KumaMeshLabelDefault defines a default value for KumaMeshLabel
	KumaMeshLabelDefault = "default"
)
