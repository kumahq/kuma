package metadata

const (
	// KumaMeshLabel defines a Pod label to associate objects
	// with a particular Mesh.
	// Label value must be the name of a Mesh resource.
	KumaMeshLabel = "kuma.io/mesh"

	// KumaZoneProxyTypeLabel marks mesh-scoped zone proxy resources.
	// On Services it drives zone proxy listener generation on the Dataplane of a
	// matching pod, and on Pods it acts as an injector hint for mesh-scoped zone
	// proxies.
	// Allowed values: ingress or egress.
	KumaZoneProxyTypeLabel = "k8s.kuma.io/zone-proxy-type"
)
