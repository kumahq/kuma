package metadata

// Annotations that can be used by the end users.
const (
	// KumaMeshAnnotation defines a Pod annotation that
	// associates a given Pod with a particular Mesh.
	// Annotation value must be the name of a Mesh resource.
	KumaMeshAnnotation = "kuma.io/mesh"

	// KumaSidecarInjectionAnnotation defines a Pod annotation that
	// gives users a chance to opt out of side-car injection
	// into a given Pod by setting its value to KumaSidecarInjectionDisabled.
	KumaSidecarInjectionAnnotation = "kuma.io/sidecar-injection"
	// KumaSidecarInjectionDisabled defines a value of KumaSidecarInjectionAnnotation
	// that will prevent Kuma from injecting a side-car into that Pod.
	KumaSidecarInjectionDisabled = "disabled"
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

	KumaIngressAnnotation = "kuma.io/ingress"
	KumaIngressEnabled    = "enabled"

	KumaDirectAccess    = "kuma.io/direct-access-services"
	KumaDirectAccessAll = "*"

	KumaMetricsPrometheusPort = "prometheus.metrics.kuma.io/port"
	KumaMetricsPrometheusPath = "prometheus.metrics.kuma.io/path"

	CNCFNetworkAnnotation = "k8s.v1.cni.cncf.io/networks"
	KumaCNI               = "kuma-cni"
)
