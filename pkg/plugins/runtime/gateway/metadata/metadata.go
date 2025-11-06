package metadata

import . "github.com/kumahq/kuma/v2/pkg/core/xds/origin"

const (
	// OriginGateway is the origin for resources produced by the gateway generator
	OriginGateway Origin = "gateway"
)

const (
	PluginName                  = "gateway"
	ProfileGatewayProxy         = "gateway-proxy"
	UnresolvedBackendServiceTag = "kuma.io/unresolved-backend"
)
