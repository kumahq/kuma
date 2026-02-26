package dpapi

const PATH = "/meshaccesslog"

// MeshAccessLogDpConfig is the configuration sent from CP to DP via dynconf for MeshAccessLog.
type MeshAccessLogDpConfig struct {
	Backends []OtelBackendConfig `json:"backends"`
}

type OtelBackendConfig struct {
	// SocketPath is the Unix socket kuma-dp listens on.
	SocketPath string `json:"socketPath"`
	// Endpoint is the host:port of the real OTel collector.
	Endpoint string `json:"endpoint"`
	// UseHTTP controls whether kuma-dp forwards via HTTP instead of gRPC.
	UseHTTP bool `json:"useHTTP"`
	// Path is the base path for HTTP forwarding (HTTP only).
	Path string `json:"path,omitempty"`
}
