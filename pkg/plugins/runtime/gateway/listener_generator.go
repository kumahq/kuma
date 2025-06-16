package gateway

import (
	"github.com/kumahq/kuma/pkg/core/kri"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
)

// TODO(jpeach) It's a lot to ask operators to tune these defaults,
// and we probably would never do that. However, it would be convenient
// to be able to update them for performance testing and benchmarking,
// so at some point we should consider making these settings available,
// perhaps on the Gateway or on the Dataplane.

// Buffer defaults.
const DefaultConnectionBuffer = 32 * 1024

type RuntimeResoureLimitListener struct {
	Name            string
	ConnectionLimit uint32
}

func GenerateListener(info GatewayListenerInfo) (*envoy_listeners.ListenerBuilder, *RuntimeResoureLimitListener) {
	// TODO(jpeach) what we really need to do here is to
	// generate a HTTP filter chain for each
	// host on the same HTTPConnectionManager. Each HTTP filter
	// chain should be wrapped in a matcher that selects it for
	// only the host's domain name. This will give us consistent
	// per-host HTTP filter chains for both HTTP and HTTPS
	// listeners.
	//
	// https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/advanced/matching/matching_api
	// A new listener gets a new filter chain.
	port := info.Listener.Port
	protocol := info.Listener.Protocol
	address := info.Proxy.Dataplane.Spec.GetNetworking().Address

	log.V(1).Info("generating listener",
		"address", address,
		"port", port,
		"protocol", protocol,
	)

	name :=  kri.From(info.Gateway, "").String()
	statName := ""
	if !info.Proxy.Metadata.HasFeature(xds_types.FeatureKRIStats) {
		statName = name
	}

	var limits *RuntimeResoureLimitListener
	if resources := info.Listener.Resources; resources != nil {
		if resources.ConnectionLimit > 0 {
			limits = &RuntimeResoureLimitListener{
				Name:            name,
				ConnectionLimit: resources.ConnectionLimit,
			}
		}
	}

	// TODO(jpeach) if proxy protocol is enabled, add the proxy protocol listener filter.
	return envoy_listeners.NewInboundListenerBuilder(
		info.Proxy.APIVersion,
		address,
		port,
		core_xds.SocketAddressProtocolTCP,
	).
		WithOverwriteName(name).
		Configure(
			// Limit default buffering for edge connections.
			envoy_listeners.ConnectionBufferLimit(DefaultConnectionBuffer),
			// Roughly balance incoming connections.
			envoy_listeners.EnableReusePort(true),
			// Always sniff for TLS.
			envoy_listeners.TLSInspector(),
			envoy_listeners.StatPrefix(statName),
	
		), limits
}
