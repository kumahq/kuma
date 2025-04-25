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

const FeatureGRPC string = "feature-grpc"

const FeatureTransparentProxyInDataplaneMetadata string = "feature-transparent-proxy-in-dataplane-metadata"
