package config

import (
	"io"
)

type TransparentProxyConfig struct {
	DryRun                         bool
	Verbose                        bool
	RedirectPortOutBound           string
	RedirectInBound                bool
	RedirectPortInBound            string
	RedirectPortInBoundV6          string
	ExcludeInboundPorts            string
	ExcludeOutboundPorts           string
	ExcludeOutboundTCPPortsForUIDs []string
	ExcludeOutboundUDPPortsForUIDs []string
	UID                            string
	GID                            string
	RedirectDNS                    bool
	RedirectAllDNSTraffic          bool
	AgentDNSListenerPort           string
	DNSUpstreamTargetChain         string
	SkipDNSConntrackZoneSplit      bool
	ExperimentalEngine             bool
	EbpfEnabled                    bool
	EbpfInstanceIP                 string
	EbpfBPFFSPath                  string
	EbpfProgramsSourcePath         string
	Stdout                         io.Writer
	Stderr                         io.Writer
}
