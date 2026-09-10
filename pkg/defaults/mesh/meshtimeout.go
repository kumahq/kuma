package mesh

import (
	policies_defaults "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/defaults"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
)

// DefaultInboundTimeout returns timeouts for the inbound side. This resource is not created
// in the store. It's used directly in InboundProxyGenerator. In the future, it could be replaced
// with a new InboundTimeout policy. The main idea around these values is to have them either
// bigger than outbound side timeouts or disabled.
var DefaultInboundTimeout = func() envoy_common.Timeouts {
	const factor = 2

	return envoy_common.Timeouts{
		Connect:        factor * policies_defaults.DefaultConnectTimeout,
		TcpIdle:        factor * policies_defaults.DefaultIdleTimeout,
		HttpIdle:       factor * policies_defaults.DefaultIdleTimeout,
		HttpStreamIdle: factor * policies_defaults.DefaultStreamIdleTimeout,
	}
}
