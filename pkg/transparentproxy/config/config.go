package config

type TransparentProxyConfig struct {
	DryRun               bool
	RedirectPortOutBound string
	RedirectPortInBound  string
	ExcludeInboundPorts  string
	ExcludeOutboundPorts string
	UID                  string
	GID                  string
}
