package generator

import (
	"github.com/kumahq/kuma/pkg/core"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/validators"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
)

const OriginInboundGateway = "inbound-gateway"

type InboundGatewayGenerator struct {
	// TrustedDownstream specifies whether we trust special headers from the downstream client.
	TrustedDownstream bool
	// ProxyProtocol specifies whether to ass the proxy protocol filter.
	ProxyProtocol bool
}

var _ ResourceGenerator = &InboundGatewayGenerator{}

func (g InboundGatewayGenerator) Generate(ctx xds_context.Context, proxy *core_xds.Proxy) (*core_xds.ResourceSet, error) {
	log := core.Log.WithName(OriginInboundGateway + "-generator")

	resources := core_xds.NewResourceSet()

	// A Gateway is a single service across all listeners.
	service := proxy.Dataplane.Spec.GetIdentifyingService()

	for i, inbound := range proxy.Dataplane.Spec.GetNetworking().GetInbound() {
		// We need an inbound interface spec, but gateways don't
		// have an inbound service, so clear the workload fields.
		endpoint := proxy.Dataplane.Spec.GetNetworking().ToInboundInterface(inbound)
		endpoint.WorkloadIP = ""
		endpoint.WorkloadPort = 0

		addr := endpoint.DataplaneIP
		port := endpoint.DataplanePort
		name := envoy_names.GetInboundListenerName(addr, port)

		log.Info("generating listener",
			"name", name,
			"addr", addr,
			"port", port,
		)

		filters := envoy_listeners.NewFilterChainBuilder(proxy.APIVersion)
		filters.Configure(
			envoy_listeners.HttpConnectionManager(name, true),
			envoy_listeners.HttpDynamicRoute(name),
			envoy_listeners.FaultInjection(proxy.Policies.FaultInjections[endpoint]),
			envoy_listeners.RateLimit(proxy.Policies.RateLimits[endpoint]),
			envoy_listeners.Tracing(proxy.Policies.TracingBackend, service),
		)

		// TODO(jpeach) enable RDS as the route source on the HttpConnectionManager.

		if g.TrustedDownstream {
			// TODO(jpeach) enable HCM header forwarding options
			// TODO(jpeach) enable request ID forwarding
		}

		listener := envoy_listeners.NewListenerBuilder(proxy.APIVersion)
		listener.Configure(
			envoy_listeners.InboundListener(name, addr, port, core_xds.SocketAddressProtocolTCP),
			// Limit default buffering for edge connections.
			envoy_listeners.ConnectionBufferLimit(32*1024),
			// Roughly balance incoming connections.
			envoy_listeners.EnableReusePort(true),
			envoy_listeners.TLSInspector(),
			envoy_listeners.FilterChain(filters),
		)

		// If MTLS is enabled, we should also accept and verify MTLS client sessions from mesh services.
		if ctx.Mesh.Resource.MTLSEnabled() {
			// TODO(jpeach) Add envoy_listeners.NetworkRBAC()
			// TODO(jpeach) Add envoy_listeners.ServerSideMTLS()
		}

		if g.ProxyProtocol {
			// TODO(jpeach) enable proxy protocol filter
		}

		listenerResource, err := listener.Build()
		if err != nil {
			return nil, errors.Wrapf(err, "%s: could not generate listener %q",
				validators.RootedAt("dataplane").Field("networking").Field("inbound").Index(i),
				name)
		}

		// TODO(jpeach) Add a layered runtime setting with the connection limit for this
		// listener name.

		resources.Add(&core_xds.Resource{
			Name:     name,
			Resource: listenerResource,
			Origin:   OriginInboundGateway,
		})
	}

	return resources, nil
}
