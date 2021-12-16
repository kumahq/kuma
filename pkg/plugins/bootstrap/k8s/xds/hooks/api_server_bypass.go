package hooks

import (
	"github.com/pkg/errors"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	xds_hooks "github.com/kumahq/kuma/pkg/xds/hooks"
)

const (
	OriginApiServerBypass            = "apiServerBypass"
	apiServerBypassHookResourcesName = "plugins:bootstrap:k8s:hooks:apiServerBypass"
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

func (h ApiServerBypass) Modify(resources *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	if proxy.Dataplane == nil {
		return nil
	}
	if ctx.Mesh.Resource.Spec.IsPassthrough() {
		return nil
	}

	listener, err := envoy_listeners.NewListenerBuilder(proxy.APIVersion).
		Configure(envoy_listeners.OutboundListener(apiServerBypassHookResourcesName, h.Address, h.Port, core_xds.SocketAddressProtocolTCP)).
		Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder(proxy.APIVersion).
			Configure(envoy_listeners.TcpProxy(apiServerBypassHookResourcesName, envoy_common.NewCluster(envoy_common.WithService(apiServerBypassHookResourcesName)))))).
		Configure(envoy_listeners.NoBindToPort()).
		Configure(envoy_listeners.OriginalDstForwarder()).
		Build()
	if err != nil {
		return errors.Wrapf(err, "could not generate listener: %s", apiServerBypassHookResourcesName)
	}

	cluster, err := envoy_clusters.NewClusterBuilder(proxy.APIVersion).
		Configure(envoy_clusters.PassThroughCluster(apiServerBypassHookResourcesName)).
		Build()
	if err != nil {
		return errors.Wrapf(err, "could not generate cluster: %s", apiServerBypassHookResourcesName)
	}

	resources.Add(&core_xds.Resource{
		Name:     listener.GetName(),
		Origin:   OriginApiServerBypass,
		Resource: listener,
	})

	resources.Add(&core_xds.Resource{
		Name:     cluster.GetName(),
		Origin:   OriginApiServerBypass,
		Resource: cluster,
	})

	return nil
}
