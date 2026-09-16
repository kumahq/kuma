package types

// Features is a set of features which a data plane has enabled.
type Features map[string]bool

// HasFeature returns true iff the feature string appears in the feature list.
func (f Features) HasFeature(feature string) bool {
	if f != nil {
		return f[feature]
	}
	return false
}

// FeatureBindOutbounds indicates that the DP runs with outbound listeners bound to 127.0.0.0/8 range addresses
const FeatureBindOutbounds string = "feature-bind-outbounds"

// FeatureSpire indicates whether the sidecar has mounted a volume that includes the socket for the Spire agent to retrieve its identity.
// Currently supported only on Kubernetes.
const FeatureSpire string = "feature-spire"

// Nothing here reads the three below. kuma-dp advertises them only so a 2.14
// control plane, which a proxy reaches whenever it reconnects mid-rollout,
// treats it the way 3.0 does. Envoy cannot change enable_reuse_port on an
// existing listener, so without the first one such a reconnect wedges the
// proxy. Remove in 3.1.
const (
	FeatureReusePort                           = "feature-reuse-port"
	FeatureStrictInboundPorts                  = "feature-strict-inbound-ports"
	FeatureTransparentProxyInDataplaneMetadata = "feature-transparent-proxy-in-dataplane-metadata"
)
