package metadata

// Annotations that can be used by the end users.
const (
	// KonvoyMeshAnnotation defines an annotation that can be put on Pods
	// in order to associate them with a particular Mesh.
	// Annotation value must be a name of a Mesh resource.
	KonvoyMeshAnnotation = "kuma.io/mesh"
)

// Annotations that are being automatically set by the Konvoy Sidecar Injector.
const (
	KonvoySidecarInjectedAnnotation = "kuma.io/sidecar-injected"
	KonvoySidecarInjected           = "true"

	KonvoyTransparentProxyingAnnotation = "kuma.io/transparent-proxying"
	KonvoyTransparentProxyingEnabled    = "enabled"

	KonvoyTransparentProxyingPortAnnotation = "kuma.io/transparent-proxying-port"
)
