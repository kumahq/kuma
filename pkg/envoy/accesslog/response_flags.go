package accesslog

// List of supported response flags.
const (
	ResponseFlagDownstreamConnectionTermination = "DC"
	ResponseFlagFailedLocalHealthCheck          = "LH"
	ResponseFlagNoHealthyUpstream               = "UH"
	ResponseFlagUpstreamRequestTimeout          = "UT"
	ResponseFlagLocalReset                      = "LR"
	ResponseFlagUpstreamRemoteReset             = "UR"
	ResponseFlagUpstreamConnectionFailure       = "UF"
	ResponseFlagUpstreamConnectionTermination   = "UC"
	ResponseFlagUpstreamOverflow                = "UO"
	ResponseFlagUpstreamRetryLimitExceeded      = "URX"
	ResponseFlagNoRouteFound                    = "NR"
	ResponseFlagDelayInjected                   = "DI"
	ResponseFlagFaultInjected                   = "FI"
	ResponseFlagRateLimited                     = "RL"
	ResponseFlagUnauthorizedExternalService     = "UAEX"
	ResponseFlagRatelimitServiceError           = "RLSE"
	ResponseFlagStreamIdleTimeout               = "SI"
	ResponseFlagInvalidEnvoyRequestHeaders      = "IH"
	ResponseFlagDownstreamProtocolError         = "DPE"
)
