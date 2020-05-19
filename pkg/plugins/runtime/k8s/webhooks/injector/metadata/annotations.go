package metadata

// Annotations that can be used by the end users.
const (
	// KumaMeshAnnotation defines a Pod annotation that
	// associates a given Pod with a particular Mesh.
	// Annotation value must be the name of a Mesh resource.
	KumaMeshAnnotation = "kuma.io/mesh"

	// KumaSidecarInjectionAnnotation defines a Pod/Namespace annotation that
	// gives users an ability to enable or disable sidecar-injection
	KumaSidecarInjectionAnnotation = "kuma.io/sidecar-injection"
	// KumaSidecarInjectionDisabled defines a value of KumaSidecarInjectionAnnotation
	// that will prevent Kuma from injecting a sidecar into that Pod or Namespace.
	KumaSidecarInjectionDisabled = "disabled"
	// KumaSidecarInjectionEnabled a value of KumaSidecarInjectionAnnotation
	// that will let Kuma to be injected as a sidecar into that Pod or Namespace.
	KumaSidecarInjectionEnabled = "enabled"
)

// Annotations that are being automatically set by the Kuma Sidecar Injector.
const (
	KumaSidecarInjectedAnnotation = "kuma.io/sidecar-injected"
	KumaSidecarInjected           = "true"

	KumaTransparentProxyingAnnotation = "kuma.io/transparent-proxying"
	KumaTransparentProxyingEnabled    = "enabled"

	KumaTransparentProxyingPortAnnotation = "kuma.io/transparent-proxying-port"

	KumaGatewayAnnotation = "kuma.io/gateway"
	KumaGatewayEnabled    = "enabled"

	KumaMetricsPrometheusPort = "prometheus.metrics.kuma.io/port"
	KumaMetricsPrometheusPath = "prometheus.metrics.kuma.io/path"

	CNCFNetworkAnnotation = "k8s.v1.cni.cncf.io/networks"
	KumaCNI               = "kuma-cni"
)
