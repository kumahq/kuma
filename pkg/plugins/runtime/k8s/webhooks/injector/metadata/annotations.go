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

	KumaTransparentProxyingInboundPortAnnotation  = "kuma.io/transparent-proxying-inbound-port"
	KumaTransparentProxyingOutboundPortAnnotation = "kuma.io/transparent-proxying-outbound-port"

	KumaGatewayAnnotation = "kuma.io/gateway"
	KumaGatewayEnabled    = "enabled"

	KumaIngressAnnotation = "kuma.io/ingress"
	KumaIngressEnabled    = "enabled"

	KumaDirectAccess    = "kuma.io/direct-access-services"
	KumaDirectAccessAll = "*"

	KumaMetricsPrometheusPort = "prometheus.metrics.kuma.io/port"
	KumaMetricsPrometheusPath = "prometheus.metrics.kuma.io/path"

	KumaTrafficExcludeInboundPorts  = "traffic.kuma.io/exclude-inbound-ports"
	KumaTrafficExcludeOutboundPorts = "traffic.kuma.io/exclude-outbound-ports"

	CNCFNetworkAnnotation = "k8s.v1.cni.cncf.io/networks"
	KumaCNI               = "kuma-cni"
)
