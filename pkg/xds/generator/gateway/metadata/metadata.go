package metadata

import . "github.com/kumahq/kuma/v3/pkg/core/xds/origin"

const (
	// OriginGateway is the origin for resources produced by the gateway generator
	OriginGateway Origin = "gateway"
)

const (
	// PluginName is the key used to store gateway listener info on
	// proxy.RuntimeExtensions.
	PluginName                  = "gateway"
	UnresolvedBackendServiceTag = "kuma.io/unresolved-backend"
)
