package gateway

import (
	"context"
	"fmt"
	"time"

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/wrapperspb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/policy"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

const DefaultConnectionBuffer = 32 * 1024
const DefaultRequestHeadersTimeoutMsec = 500 * time.Millisecond
const DefaultStreamIdleTimeoutMsec = 5 * time.Second
const DefaultConcurrentStreams = 100
const DefaultInitialStreamWindowSize = 64 * 1024
const DefaultInitialConnectionWindowSize = 1024 * 1024
const DefaultIdleTimeout = 5 * time.Minute

type gatewayPolicyAdaptor struct {
	*core_mesh.GatewayResource
}

func (g gatewayPolicyAdaptor) Selectors() []*mesh_proto.Selector {
	return g.Sources()
}

func findMatchingGateway(m manager.ReadOnlyResourceManager, dp *core_mesh.DataplaneResource) *core_mesh.GatewayResource {
	gatewayList := &core_mesh.GatewayResourceList{}

	if err := m.List(context.Background(), gatewayList, store.ListByMesh(dp.Meta.GetMesh())); err != nil {
		return nil
	}

	candidates := make([]policy.DataplanePolicy, len(gatewayList.Items))
	for i, gw := range gatewayList.Items {
		candidates[i] = gatewayPolicyAdaptor{gw}
	}

	if p := policy.SelectDataplanePolicy(dp, candidates); p != nil {
		return p.(gatewayPolicyAdaptor).GatewayResource
	}

	return nil
}

var _ policy.DataplanePolicy = gatewayPolicyAdaptor{}

// ListenerGenerator generates Kuma gateway listeners.
type ListenerGenerator struct {
	Resources manager.ReadOnlyResourceManager
}

var _ generator.ResourceGenerator = ListenerGenerator{}

func (l ListenerGenerator) Generate(_ xds_context.Context, proxy *core_xds.Proxy) (*core_xds.ResourceSet, error) {
	gw := findMatchingGateway(l.Resources, proxy.Dataplane)
	if gw == nil {
		log.V(1).Info("no matching gateway for dataplane",
			"name", proxy.Dataplane.Meta.GetName(),
			"mesh", proxy.Dataplane.Meta.GetMesh(),
		)

		return nil, nil
	}

	log.V(1).Info(fmt.Sprintf("matched gateway %q to dataplane %q",
		gw.Meta.GetName(), proxy.Dataplane.Meta.GetName()))

	// Multiple listener specifications can have the same port. If
	// they are compatible, then we can collapse those specifications
	// down to a single listener.
	collapsed := map[uint32][]*mesh_proto.Gateway_Listener{}
	for _, ep := range gw.Spec.GetConf().GetListeners() {
		collapsed[ep.GetPort()] = append(collapsed[ep.GetPort()], ep)
	}

	// A Gateway is a single service across all listeners.
	service := proxy.Dataplane.Spec.GetIdentifyingService()

	resources := core_xds.NewResourceSet()

	// Generate a listener resource for each port.
	for port, listeners := range collapsed {
		protocol := listeners[0].GetProtocol()
		address := proxy.Dataplane.Spec.GetNetworking().Address

		log.V(1).Info("generating listener",
			"address", address,
			"port", port,
			"protocol", protocol,
		)

		// TODO(jpeach) verify that the listeners are compatible.
		// TODO(jpeach) hoist the compatibility check and use it in Gateway validation.

		// This check forces all listeners on the port to have
		// the same protocol, which is unnecessarily strict. We
		// cal allow TLS and HTTPS on the same port, for example.
		for i := range listeners {
			if listeners[i].GetProtocol() != listeners[0].GetProtocol() {
				return nil, errors.Errorf("cannot collapse listener protocols %s and %s",
					listeners[i].GetProtocol(), listeners[0].GetProtocol(),
				)
			}
		}

		switch protocol {
		case mesh_proto.Gateway_Listener_UDP:
			fallthrough
		case mesh_proto.Gateway_Listener_TCP:
			fallthrough
		case mesh_proto.Gateway_Listener_TLS:
			fallthrough
		case mesh_proto.Gateway_Listener_HTTPS:
			return nil, errors.Errorf("unsupported protocol %q", protocol)
		}

		filters := envoy_listeners.NewFilterChainBuilder(proxy.APIVersion)

		// For HTTP protocols, add the connection manager and
		// (for now), send all traffic to the default route,
		// just to keep the xDS snapshot consistent.
		switch protocol {
		case mesh_proto.Gateway_Listener_HTTP:
			filters.Configure(
				envoy_listeners.HttpConnectionManager(service, false),
				envoy_listeners.HttpDynamicRoute(DefaultRouteName),
			)
		case mesh_proto.Gateway_Listener_HTTPS:
			filters.Configure(
				envoy_listeners.HttpConnectionManager(service, true),
				envoy_listeners.HttpDynamicRoute(DefaultRouteName),
			)
		}

		// Add edge proxy recommendations.
		filters.Configure(
			envoy_listeners.AddFilterChainConfigurer(
				v3.HttpConnectionManagerMustConfigureFunc(func(hcm *envoy_hcm.HttpConnectionManager) {
					hcm.ServerName = "Kuma Gateway"

					hcm.NormalizePath = wrapperspb.Bool(true)
					hcm.MergeSlashes = true

					// TODO(jpeach) set path_with_escaped_slashes_action when we upgrade to Envoy v1.19.

					hcm.RequestHeadersTimeout = util_proto.Duration(DefaultRequestHeadersTimeoutMsec)
					hcm.StreamIdleTimeout = util_proto.Duration(DefaultStreamIdleTimeoutMsec)

					hcm.CommonHttpProtocolOptions = &envoy_config_core_v3.HttpProtocolOptions{
						IdleTimeout:                  util_proto.Duration(DefaultIdleTimeout),
						HeadersWithUnderscoresAction: envoy_config_core_v3.HttpProtocolOptions_REJECT_REQUEST,
					}

					hcm.Http2ProtocolOptions = &envoy_config_core_v3.Http2ProtocolOptions{
						MaxConcurrentStreams:        wrapperspb.UInt32(DefaultConcurrentStreams),
						InitialStreamWindowSize:     wrapperspb.UInt32(DefaultInitialStreamWindowSize),
						InitialConnectionWindowSize: wrapperspb.UInt32(DefaultInitialConnectionWindowSize),
						AllowConnect:                true,
					}
				}),
			),
		)

		// Tracing and logging have to be configured after the HttpConnectionManager is enabled.
		filters.Configure(
			envoy_listeners.Tracing(proxy.Policies.TracingBackend, service),
			// XXX Logging policy doesn't work at all. The logging backend is selected by
			// matching against outbound service names, and gateway dataplanes don't have
			// any of those.
			envoy_listeners.HttpAccessLog(
				proxy.Dataplane.Meta.GetMesh(),
				envoy.TrafficDirectionInbound,
				service, // Source service is the gateway service.
				"*",     // Destination service could be anywhere, depending on the routes.
				proxy.Policies.Logs[service],
				proxy,
			),
		)

		// TODO(jpeach) add compressor filter.
		// TODO(jpeach) add decompressor filter.
		// TODO(jpeach) add grpc_web filter.
		// TODO(jpeach) add grpc_stats filter.

		listener := envoy_listeners.NewListenerBuilder(proxy.APIVersion)
		listener.Configure(
			envoy_listeners.InboundListener(
				envoy_names.GetGatewayListenerName(gw.Meta.GetName(), protocol.String(), port),
				address, port, core_xds.SocketAddressProtocolTCP),
			// Limit default buffering for edge connections.
			envoy_listeners.ConnectionBufferLimit(DefaultConnectionBuffer),
			// Roughly balance incoming connections.
			envoy_listeners.EnableReusePort(true),
			// Always sniff for TLS.
			envoy_listeners.TLSInspector(),
		)

		// TODO(jpeach) if proxy protocol is enabled, add the listener filter.

		// Now, for each of the collapsed listeners on this port, configure the
		for range listeners {
			switch protocol {
			case mesh_proto.Gateway_Listener_HTTPS:
				// TODO(jpeach) add a SNI listener to match the hostname
				// and apply the right set of dynamic HTTP routes.
			case mesh_proto.Gateway_Listener_TLS:
				// TODO(jpeach) add a SNI listener to match the hostname
				// and apply the right set of dynamic TCP or TLS routes.
			}
		}

		listener.Configure(
			envoy_listeners.FilterChain(filters),
		)

		resourceSet, err := BuildResourceSet(listener)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to generate listener for port %d", port)
		}

		resources.AddSet(resourceSet)
	}

	return resources, nil
}
