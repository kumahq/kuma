package generator

import (
	"context"
	"net/url"

	"github.com/pkg/errors"

	model "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/envoy/names"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
	envoy_virtual_hosts "github.com/kumahq/kuma/pkg/xds/envoy/virtualhosts"
)

const (
	// OriginProbes is a marker to indicate by which ProxyGenerator resources were generated.
	OriginProbe            = "probe"
	listenerName           = "probe:listener"
	routeConfigurationName = "probe:route_configuration"
)

type ProbeProxyGenerator struct{}

func (g ProbeProxyGenerator) Generate(ctx context.Context, _ *model.ResourceSet, xdsCtx xds_context.Context, proxy *model.Proxy) (*model.ResourceSet, error) {
	// if app probe proxy is enabled for this DP, Virtual Probes are not needed
	appProbeProxyEnabled := proxy.Metadata.GetAppProbeProxyEnabled()
	if appProbeProxyEnabled {
		return nil, nil
	}

	probes := proxy.Dataplane.Spec.Probes
	if probes == nil {
		return nil, nil
	}

	virtualHostBuilder := envoy_virtual_hosts.NewVirtualHostBuilder(proxy.APIVersion, "probe")

	portSet := map[uint32]bool{}
	for _, inbound := range proxy.Dataplane.Spec.Networking.Inbound {
		portSet[proxy.Dataplane.Spec.Networking.ToInboundInterface(inbound).WorkloadPort] = true
	}
	for _, endpoint := range probes.Endpoints {
		matchURL, err := url.Parse(endpoint.Path)
		if err != nil {
			return nil, err
		}
		newURL, err := url.Parse(endpoint.InboundPath)
		if err != nil {
			return nil, err
		}
		if portSet[endpoint.InboundPort] {
			virtualHostBuilder.Configure(
				envoy_virtual_hosts.Route(matchURL.Path, newURL.Path, names.GetLocalClusterName(endpoint.InboundPort), true))
		} else {
			// On Kubernetes we are overriding probes for every container, but there is no guarantee that given
			// probe will have an equivalent in inbound interface (ex. sidecar that is not selected by any service).
			// In this situation there is no local cluster therefore we are sending redirect to a real destination.
			// System responsible for using virtual probes needs to support redirect (kubelet on K8S supports it).
			virtualHostBuilder.Configure(
				envoy_virtual_hosts.Redirect(matchURL.Path, newURL.Path, true, endpoint.InboundPort))
		}
	}

	probeListener, err := envoy_listeners.NewInboundListenerBuilder(proxy.APIVersion, proxy.Dataplane.Spec.GetNetworking().GetAddress(), probes.Port, model.SocketAddressProtocolTCP).
		WithOverwriteName(listenerName).
		Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).
			Configure(envoy_listeners.HttpConnectionManager(listenerName, false)).
			Configure(envoy_listeners.HttpStaticRoute(envoy_routes.NewRouteConfigurationBuilder(proxy.APIVersion, routeConfigurationName).
				Configure(envoy_routes.VirtualHost(virtualHostBuilder)))))).
		Configure(envoy_listeners.TransparentProxying(proxy.Dataplane.Spec.Networking.GetTransparentProxying())).
		Build()
	if err != nil {
		return nil, errors.Wrapf(err, "could not generate listener %s", listenerName)
	}

	resources := model.NewResourceSet()
	resources.Add(&model.Resource{
		Name:     listenerName,
		Resource: probeListener,
		Origin:   OriginProbe,
	})

	return resources, nil
}
