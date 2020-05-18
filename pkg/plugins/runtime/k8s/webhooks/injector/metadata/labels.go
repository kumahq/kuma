package metadata

// Labels that can be used by end users.
const (
	// KumaSidecarInjectionLabel defines a Namespace label that allows to enable and disable sidecar-injection
	// for all pods in that Namespace unless that behaviour is overridden in Pod
	KumaSidecarInjectionLabel = "kuma.io/sidecar-injection"

	// KumaMeshLabel defines a Namespace label that associates every Pod in that Namespace with a particular Mesh
	// unless that behaviour is overridden in Pod. Annotation value must be the name of a Mesh resource.
	KumaMeshLabel = "kuma.io/mesh"
)
