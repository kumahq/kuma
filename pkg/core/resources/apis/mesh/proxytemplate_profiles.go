package mesh

const (
	ProfileDefaultProxy = "default-proxy"
)

// AvailableProfiles is populated by generator.RegisterProfile for every
// registered ProxyTemplate profile (default-proxy, ingress-proxy, egress-proxy).
var AvailableProfiles = map[string]struct{}{}
