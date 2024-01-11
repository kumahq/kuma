package xds

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
