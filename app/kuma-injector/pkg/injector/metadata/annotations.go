package metadata

// Annotations that can be used by the end users.
const (
	// KumaMeshAnnotation defines an annotation that can be put on Pods
	// in order to associate them with a particular Mesh.
	// Annotation value must be a name of a Mesh resource.
	KumaMeshAnnotation = "kuma.io/mesh"
)

// Annotations that are being automatically set by the Kuma Sidecar Injector.
const (
	KumaSidecarInjectedAnnotation = "kuma.io/sidecar-injected"
	KumaSidecarInjected           = "true"

	KumaTransparentProxyingAnnotation = "kuma.io/transparent-proxying"
	KumaTransparentProxyingEnabled    = "enabled"

	KumaTransparentProxyingPortAnnotation = "kuma.io/transparent-proxying-port"
)
