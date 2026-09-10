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

const FeatureTransparentProxyInDataplaneMetadata string = "feature-transparent-proxy-in-dataplane-metadata"

// FeatureBindOutbounds indicates that the DP runs with outbound listeners bound to 127.0.0.0/8 range addresses
const FeatureBindOutbounds string = "feature-bind-outbounds"

// FeatureSpire indicates whether the sidecar has mounted a volume that includes the socket for the Spire agent to retrieve its identity.
// Currently supported only on Kubernetes.
const FeatureSpire string = "feature-spire"

// FeatureOtelViaKumaDp indicates that kuma-dp can act as a gRPC proxy for OTel
// traces and access logs. When present, the CP routes the OTel cluster to a Unix
// socket instead of connecting directly to the collector.
const FeatureOtelViaKumaDp = "feature-otel-via-kuma-dp"
