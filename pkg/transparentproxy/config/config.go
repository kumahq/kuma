package config

type TransparentProxyConfig struct {
	DryRun               bool
	RedirectPortOutBound string
	RedirectInBound      bool
	RedirectPortInBound  string
	ExcludeInboundPorts  string
	ExcludeOutboundPorts string
	UID                  string
	GID                  string
}
