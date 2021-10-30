package gateway

import (
	"time"

	envoy_config_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
)

// TODO(jpeach) It's a lot to ask operators to tune these defaults,
// and we probably would never do that. However, it would be convenient
// to be able to update them for performance testing and benchmarking,
// so at some point we should consider making these settings available,
// perhaps on the Gateway or on the Dataplane.

// Buffer defaults.
const DefaultConnectionBuffer = 32 * 1024

// Concurrency defaults.
const DefaultConcurrentStreams = 100

// Window size defaults.
const DefaultInitialStreamWindowSize = 64 * 1024
const DefaultInitialConnectionWindowSize = 1024 * 1024

// Timeout defaults.
const DefaultRequestHeadersTimeout = 500 * time.Millisecond
const DefaultStreamIdleTimeout = 5 * time.Second
const DefaultIdleTimeout = 5 * time.Minute

// ListenerGenerator generates Kuma gateway listeners.
type ListenerGenerator struct{}

func (*ListenerGenerator) SupportsProtocol(p mesh_proto.Gateway_Listener_Protocol) bool {
	switch p {
	case mesh_proto.Gateway_Listener_UDP,
		mesh_proto.Gateway_Listener_TCP,
		mesh_proto.Gateway_Listener_TLS,
		mesh_proto.Gateway_Listener_HTTP,
		mesh_proto.Gateway_Listener_HTTPS:
		return true
	default:
		return false
	}
}

func (*ListenerGenerator) GenerateHost(ctx xds_context.Context, info *GatewayResourceInfo) (*core_xds.ResourceSet, error) {
	// TODO(jpeach) what we really need to do here is build the
	// listener once, then generate a HTTP filter chain for each
	// host on the same HTTPConnectionManager. Each HTTP filter
	// chain should be wrapped in a matcher that selects it for
	// only the host's domain name. This will give us consistent
	// per-host HTTP filter chains for both HTTP and HTTPS
	// listeners.
	//
	// https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/advanced/matching/matching_api
	if info.Resources.Listener != nil {
		return nil, nil
	}

	// A Gateway is a single service across all listeners.
	service := info.Dataplane.Spec.GetIdentifyingService()

	port := info.Listener.Port
	protocol := info.Listener.Protocol
	address := info.Dataplane.Spec.GetNetworking().Address

	log.V(1).Info("generating listener",
		"address", address,
		"port", port,
		"protocol", protocol,
	)

	switch protocol {
	case mesh_proto.Gateway_Listener_UDP,
		mesh_proto.Gateway_Listener_TCP,
		mesh_proto.Gateway_Listener_TLS,
		mesh_proto.Gateway_Listener_HTTPS:
		return nil, errors.Errorf("unsupported protocol %q", protocol)
	}

	filters := envoy_listeners.NewFilterChainBuilder(info.Proxy.APIVersion)

	switch protocol {
	case mesh_proto.Gateway_Listener_HTTP,
		mesh_proto.Gateway_Listener_HTTPS:
		filters.Configure(
			// Note that even for HTTPS cases, we don't enable client certificate
			// forwarding. This is because this particular configurer will enable
			// forwarding for the client certificate URI, which is OK for SPIFFE-
			// oriented mesh use cases, but unlikely to be appropriate for a
			// general-purpose gateway.
			envoy_listeners.HttpConnectionManager(service, false),
			envoy_listeners.ServerHeader("Kuma Gateway"),
			envoy_listeners.HttpDynamicRoute(info.Listener.ResourceName),
		)
	}

	// Add edge proxy recommendations.
	filters.Configure(
		envoy_listeners.EnablePathNormalization(),
		envoy_listeners.StripHostPort(),
		envoy_listeners.AddFilterChainConfigurer(
			v3.HttpConnectionManagerMustConfigureFunc(func(hcm *envoy_hcm.HttpConnectionManager) {
				hcm.RequestHeadersTimeout = util_proto.Duration(DefaultRequestHeadersTimeout)
				hcm.StreamIdleTimeout = util_proto.Duration(DefaultStreamIdleTimeout)

				hcm.CommonHttpProtocolOptions = &envoy_config_core.HttpProtocolOptions{
					IdleTimeout:                  util_proto.Duration(DefaultIdleTimeout),
					HeadersWithUnderscoresAction: envoy_config_core.HttpProtocolOptions_REJECT_REQUEST,
				}

				hcm.Http2ProtocolOptions = &envoy_config_core.Http2ProtocolOptions{
					MaxConcurrentStreams:        util_proto.UInt32(DefaultConcurrentStreams),
					InitialStreamWindowSize:     util_proto.UInt32(DefaultInitialStreamWindowSize),
					InitialConnectionWindowSize: util_proto.UInt32(DefaultInitialConnectionWindowSize),
					AllowConnect:                true,
				}
			}),
		),
	)

	// Tracing and logging have to be configured after the HttpConnectionManager is enabled.
	filters.Configure(
		envoy_listeners.Tracing(info.Proxy.Policies.TracingBackend, service),
		// TODO(jpeach) Logging policy doesn't work at all. The logging backend is
		// selected by matching against outbound service names, and gateway dataplanes
		// don't have any of those.
		envoy_listeners.HttpAccessLog(
			ctx.Mesh.Resource.Meta.GetName(),
			envoy.TrafficDirectionInbound,
			service, // Source service is the gateway service.
			"*",     // Destination service could be anywhere, depending on the routes.
			info.Proxy.Policies.Logs[service],
			info.Proxy,
		),
	)

	// TODO(jpeach) add compressor filter.
	// TODO(jpeach) add decompressor filter.
	// TODO(jpeach) add grpc_web filter.
	// TODO(jpeach) add grpc_stats filter.

	info.Resources.Listener = envoy_listeners.NewListenerBuilder(info.Proxy.APIVersion).
		Configure(
			envoy_listeners.InboundListener(
				envoy_names.GetGatewayListenerName(info.Gateway.Meta.GetName(), protocol.String(), port),
				address, port, core_xds.SocketAddressProtocolTCP),
			// Limit default buffering for edge connections.
			envoy_listeners.ConnectionBufferLimit(DefaultConnectionBuffer),
			// Roughly balance incoming connections.
			envoy_listeners.EnableReusePort(true),
			// Always sniff for TLS.
			envoy_listeners.TLSInspector(),
		)

	// TODO(jpeach) if proxy protocol is enabled, add the proxy protocol listener filter.

	// Now, for each of the virtual hosts this port, configure the
	// TLS transport sockets and matching.
	switch protocol {
	case mesh_proto.Gateway_Listener_HTTPS:
		// TODO(jpeach) add a SNI listener to match the hostname
		// and apply the right set of dynamic HTTP routes.
	case mesh_proto.Gateway_Listener_TLS:
		// TODO(jpeach) add a SNI listener to match the hostname
		// and apply the right set of dynamic TCP or TLS routes.
	}

	info.Resources.Listener.Configure(
		envoy_listeners.FilterChain(filters),
	)

	return nil, nil
}
