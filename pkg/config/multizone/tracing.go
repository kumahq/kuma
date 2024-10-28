package multizone

type KDSServerTracing struct {
	// Defines whether tracing is enabled for all gRPC methods
	// of GlobalKDSServer and KDSSyncService or completely disabled
	Enabled bool `json:"enabled,omitempty" envconfig:"kuma_multizone_global_kds_tracing_enabled"`
}
