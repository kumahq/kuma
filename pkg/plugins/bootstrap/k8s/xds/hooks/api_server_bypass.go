package hooks

import (
	"github.com/pkg/errors"

	unified_naming "github.com/kumahq/kuma/pkg/core/naming/unified-naming"
	"github.com/kumahq/kuma/pkg/core/system_names"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/bootstrap/k8s/xds/hooks/metadata"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	xds_hooks "github.com/kumahq/kuma/pkg/xds/hooks"
)

const apiServerBypassHookResourcesName = "plugins:bootstrap:k8s:hooks:apiServerBypass" // #nosec G101 -- no idea why gosec things this is a secret

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

	getNameOrDefault := system_names.GetNameOrDefault(unified_naming.Enabled(proxy.Metadata, ctx.Mesh.Resource))

	name := getNameOrDefault(
		system_names.MustBeSystemName("kube_api_server_bypass"),
		apiServerBypassHookResourcesName,
	)

	listener, err := envoy_listeners.NewOutboundListenerBuilder(proxy.APIVersion, h.Address, h.Port, core_xds.SocketAddressProtocolTCP).
		WithOverwriteName(name).
		Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).
			Configure(envoy_listeners.TcpProxyDeprecated(name, envoy_common.NewCluster(envoy_common.WithService(name)))))).
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
