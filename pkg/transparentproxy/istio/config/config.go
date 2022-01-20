package config

type TransparentProxyConfig struct {
	DryRun                    bool
	Verbose                   bool
	RedirectPortOutBound      string
	RedirectInBound           bool
	RedirectPortInBound       string
	RedirectPortInBoundV6     string
	ExcludeInboundPorts       string
	ExcludeOutboundPorts      string
	UID                       string
	GID                       string
	RedirectDNS               bool
	RedirectAllDNSTraffic     bool
	AgentDNSListenerPort      string
	DNSUpstreamTargetChain    string
	SkipDNSConntrackZoneSplit bool
}
