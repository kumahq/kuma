package metadata

import . "github.com/kumahq/kuma/pkg/core/xds/origin"

const (
	// OriginGateway is the origin for resources produced by the gateway generator
	OriginGateway               Origin = "gateway"
	PluginName                         = "gateway"
	ProfileGatewayProxy                = "gateway-proxy"
	UnresolvedBackendServiceTag        = "kuma.io/unresolved-backend"
)
