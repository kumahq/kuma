package metadata

import core_plugins "github.com/kumahq/kuma/pkg/core/plugins"

// OriginIngressGateway marks xDS resources generated by this plugin.
const (
	OriginIngressGateway                               = "ingress-gateway"
	PluginName                 core_plugins.PluginName = "ingress-gateway"
	ProfileIngressGatewayProxy                         = "ingress-gateway-proxy"
)
