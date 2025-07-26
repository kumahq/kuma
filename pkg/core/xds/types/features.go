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

// FeatureTCPAccessLogViaNamedPipe indicates that the DP implements TCP accesslog
// across a named pipe. Sotw DP versions may use structured data across GRPC.
const FeatureTCPAccessLogViaNamedPipe string = "feature-tcp-accesslog-via-named-pipe"

// FeatureEmbeddedDNS indicates that the DP runs with the embedded DNS instead of the buddy coreDNS
const FeatureEmbeddedDNS string = "feature-embedded-dns"

// FeatureDeltaGRPC indicates that the Envoy sidecar uses incremental xDS for configuration exchange.
// https://www.envoyproxy.io/docs/envoy/latest/api-docs/xds_protocol#xds-protocol-delta
const FeatureDeltaGRPC string = "feature-delta-grpc"

const FeatureTransparentProxyInDataplaneMetadata string = "feature-transparent-proxy-in-dataplane-metadata"

// FeatureBindOutbounds indicates that the DP runs with outbound listeners bound to 127.0.0.0/8 range addresses
const FeatureBindOutbounds string = "feature-bind-outbounds"

// FeatureUnifiedResourceNaming indicates that the proxy (data plane, zone ingress, or zone egress)
// uses the unified naming format for Envoy resources and stats. This includes KRI-based naming for
// distinct Kuma resources, contextual naming for proxy-scoped resources like inbounds and transparent
// proxy passthrough, and system format for internal Kuma resources that users typically
// don't need to care about unless debugging Kuma.
const FeatureUnifiedResourceNaming string = "feature-unified-resource-naming"

// FeatureReadinessUnixSocket indicates the readiness probe of kuma-sidecar is responded from the kuma-dp process via Unix socket.
// TODO: remove in 2.15 or higher, see: https://github.com/kumahq/kuma/issues/14039
const FeatureReadinessUnixSocket = "feature-readiness-unix-socket"
