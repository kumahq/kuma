package metadata

// Annotations that can be used by the end users.
const (
	// KonvoyMeshAnnotation defines an annotation that can be put on Pods
	// in order to associate them with a particular Mesh.
	// Annotation value must be a name of a Mesh resource.
	KonvoyMeshAnnotation = "getkonvoy.io/mesh"
)

// Annotations that are being automatically set by the Konvoy Sidecar Injector.
const (
	KonvoySidecarInjectedAnnotation = "getkonvoy.io/sidecar-injected"
	KonvoySidecarInjected           = "true"

	KonvoyTransparentProxyingAnnotation = "getkonvoy.io/transparent-proxying"
	KonvoyTransparentProxyingEnabled    = "enabled"

	KonvoyTransparentProxyingPortAnnotation = "getkonvoy.io/transparent-proxying-port"
)
