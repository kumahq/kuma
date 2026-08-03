package hooks

import (
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/v3/pkg/core/naming"
	"github.com/kumahq/kuma/v3/pkg/core/system_names"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/plugins/bootstrap/k8s/xds/hooks/metadata"
	plugins_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/v3/pkg/xds/envoy/clusters"
	envoy_listeners "github.com/kumahq/kuma/v3/pkg/xds/envoy/listeners"
	xds_hooks "github.com/kumahq/kuma/v3/pkg/xds/hooks"
)

type ApiServerBypass struct {
	Address string
	Port    uint32
}

var _ xds_hooks.ResourceSetHook = &ApiServerBypass{}

func NewApiServerBypass(address string, port uint32) ApiServerBypass {
	return ApiServerBypass{
		Address: address,
		Port:    port,
	}
}

func (h ApiServerBypass) Modify(resources *core_xds.ResourceSet, _ xds_context.Context, proxy *core_xds.Proxy) error {
	if proxy.Dataplane == nil {
		return nil
	}
	outboundPassThroughClusterName := naming.ContextualTransparentProxyName("outbound", 4)
	if _, ok := resources.Resources(envoy_resource.ClusterType)[outboundPassThroughClusterName]; ok {
		// default outbound passthrough is in effect for this proxy, so it can already reach the API Server
		return nil
	}

	name := system_names.MustBeSystemName("kube_api_server_bypass")

	listener, err := envoy_listeners.NewOutboundListenerBuilder(proxy.APIVersion, h.Address, h.Port, core_xds.SocketAddressProtocolTCP).
		WithOverwriteName(name).
		Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).
			Configure(envoy_listeners.TcpProxyDeprecated(name, plugins_xds.NewClusterBuilder().WithService(name).Build())))).
		Configure(envoy_listeners.NoBindToPort()).
		Configure(envoy_listeners.OriginalDstForwarder()).
		Build()
	if err != nil {
		return errors.Wrapf(err, "could not generate listener: %s", name)
	}

	cluster, err := envoy_clusters.NewClusterBuilder(proxy.APIVersion, name).
		Configure(envoy_clusters.PassThroughCluster()).
		Configure(envoy_clusters.DefaultTimeout()).
		Build()
	if err != nil {
		return errors.Wrapf(err, "could not generate cluster: %s", name)
	}

	resources.Add(&core_xds.Resource{
		Name:     listener.GetName(),
		Origin:   metadata.OriginAPIServerBypass,
		Resource: listener,
	})

	resources.Add(&core_xds.Resource{
		Name:     cluster.GetName(),
		Origin:   metadata.OriginAPIServerBypass,
		Resource: cluster,
	})

	return nil
}
