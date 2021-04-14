package config

type TransparentProxyConfig struct {
	DryRun                bool
	RedirectPortOutBound  string
	RedirectInBound       bool
	RedirectPortInBound   string
	RedirectPortInBoundV6 string
	ExcludeInboundPorts   string
	ExcludeOutboundPorts  string
	UID                   string
	GID                   string
	RedirectDNS           bool
	AgentDNSListenerPort  string
}
