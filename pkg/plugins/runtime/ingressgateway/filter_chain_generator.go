package ingressgateway

import (
	"time"

	envoy_config_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_listeners_v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
)

// TODO(jpeach) It's a lot to ask operators to tune these defaults,
// and we probably would never do that. However, it would be convenient
// to be able to update them for performance testing and benchmarking,
// so at some point we should consider making these settings available,
// perhaps on the Gateway or on the Dataplane.

// Concurrency defaults.
const DefaultConcurrentStreams = 100

// Window size defaults.
const (
	DefaultInitialStreamWindowSize     = 64 * 1024
	DefaultInitialConnectionWindowSize = 1024 * 1024
)

// Timeout defaults.
const (
	DefaultRequestHeadersTimeout = 500 * time.Millisecond
	DefaultStreamIdleTimeout     = 10 * time.Second
	DefaultIdleTimeout           = 10 * time.Minute
)

type HTTPFilterChainGenerator struct{}

func (g *HTTPFilterChainGenerator) Generate(
	xdsCtx xds_context.Context, info GatewayListenerInfo,
) (
	*core_xds.ResourceSet, []*envoy_listeners.FilterChainBuilder, error,
) {
	log.V(1).Info("generating filter chain", "protocol", "HTTP")

	// HTTP listeners get a single filter chain for all hostnames. So
	// if there's already a filter chain, we have nothing to do.
	return nil, []*envoy_listeners.FilterChainBuilder{newHTTPFilterChain(xdsCtx, info)}, nil
}

// HTTPSFilterChainGenerator generates a filter chain for an HTTPS listener.
type HTTPSFilterChainGenerator struct{}

func (g *HTTPSFilterChainGenerator) Generate(
	xdsCtx xds_context.Context, info GatewayListenerInfo,
) (
	*core_xds.ResourceSet, []*envoy_listeners.FilterChainBuilder, error,
) {
	builder := newHTTPFilterChain(xdsCtx, info)
	builder.Configure(
		envoy_listeners.ServerSideMTLS(xdsCtx.Mesh.Resource, info.Proxy.SecretsTracker),
	)

	return nil, []*envoy_listeners.FilterChainBuilder{builder}, nil
}

func newHTTPFilterChain(xdsCtx xds_context.Context, info GatewayListenerInfo) *envoy_listeners.FilterChainBuilder {
	// A Gateway is a single service across all listeners.
	service := info.Proxy.Dataplane.Spec.GetIdentifyingService()

	builder := envoy_listeners.NewFilterChainBuilder(info.Proxy.APIVersion, envoy_common.AnonymousResource).Configure(
		// Note that even for HTTPS cases, we don't enable client certificate
		// forwarding. This is because this particular configurer will enable
		// forwarding for the client certificate URI, which is OK for SPIFFE-
		// oriented mesh use cases, but unlikely to be appropriate for a
		// general-purpose gateway.
		envoy_listeners.HttpConnectionManager(service, false),
		envoy_listeners.ServerHeader("Koyeb Ingress Gateway"),
		// Use dynamic routes because we are going to update them often. Whenever a static route
		// is updated, the listener is reloaded, which resets all inbound connections.
		// We want to keep those connections live because they could be long-lived (e.g. websockets)

		envoy_listeners.HttpDynamicRoute(info.Listener.ResourceName),
		// TODO(nicoche)
		// envoy_listeners.MaxConnectAttempts(&defaultRetryPolicy),
		// envoy_listeners.LocalReplyConfig(
		//	mapper503To502,
		//	// If X-KOYEB-ROUTE does not fit to an existing cluster, display
		//	// a custom HTML page and a 503 error code
		//	igwFallbackNoClusterHeader,
		// ),
	)

	// Add edge proxy recommendations.
	builder.Configure(
		envoy_listeners.EnablePathNormalization(),
		envoy_listeners.AddFilterChainConfigurer(
			envoy_listeners_v3.HttpConnectionManagerMustConfigureFunc(func(hcm *envoy_hcm.HttpConnectionManager) {
				hcm.UseRemoteAddress = util_proto.Bool(true)

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
				hcm.UpgradeConfigs = append(hcm.UpgradeConfigs, &envoy_hcm.HttpConnectionManager_UpgradeConfig{
					UpgradeType: "websocket",
					Enabled:     util_proto.Bool(true),
				})
			}),
		),
	)

	// Tracing and logging have to be configured after the HttpConnectionManager is enabled.
	builder.Configure(
		// Force the ratelimit filter to always be present. This
		// is a no-op unless we later add a per-route configuration.
		envoy_listeners.RateLimit([]*core_mesh.RateLimitResource{nil}),
		envoy_listeners.DefaultCompressorFilter(),
		// In mesh proxies, the access log is configured on the outbound
		// listener, which is why we index the Logs slice by destination
		// service name.  A Gateway listener by definition forwards traffic
		// to multiple destinations, so rather than making up some arbitrary
		// rules about which destination service we should accept here, we
		// match the log policy for the generic pass through service. This
		// will be the only policy available for a Dataplane with no outbounds.
		envoy_listeners.HttpAccessLog(
			xdsCtx.Mesh.Resource.Meta.GetName(),
			envoy.TrafficDirectionInbound,
			service,                // Source service is the gateway service.
			mesh_proto.MatchAllTag, // Destination service could be anywhere, depending on the routes.
			xdsCtx.Mesh.GetLoggingBackend(info.Proxy.Policies.TrafficLogs[core_mesh.PassThroughService]),
			info.Proxy,
		),
	)
	builder.AddConfigurer(&envoy_listeners_v3.HTTPRouterStartChildSpanRouter{})

	// TODO(jpeach) if proxy protocol is enabled, add the proxy protocol listener filter.

	return builder
}
