package v1alpha1

import (
	"context"
	"slices"
	"strings"

	"github.com/pkg/errors"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtcproute/api/v1alpha1"
	plugin_gateway "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/route"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
)

func generateGatewayListeners(
	ctx xds_context.Context,
	info plugin_gateway.GatewayListenerInfo,
) (*core_xds.ResourceSet, *plugin_gateway.RuntimeResoureLimitListener, error) {
	resources := core_xds.NewResourceSet()

	listenerBuilder, limit := plugin_gateway.GenerateListener(info)

	generator := &plugin_gateway.TCPFilterChainGenerator{}
	res, filterChainBuilders, err := generator.Generate(ctx, info)
	if err != nil {
		return nil, limit, err
	}
	resources.AddSet(res)

	for _, filterChainBuilder := range filterChainBuilders {
		listenerBuilder.Configure(envoy_listeners.FilterChain(filterChainBuilder))
	}

	res, err = plugin_gateway.BuildResourceSet(listenerBuilder)
	if err != nil {
		return nil, limit, errors.Wrapf(err, "failed to build listener resource")
	}
	resources.AddSet(res)

	return resources, limit, nil
}

func generateGatewayClusters(
	ctx context.Context,
	xdsCtx xds_context.Context,
	info plugin_gateway.GatewayListenerInfo,
) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()

	gen := plugin_gateway.ClusterGenerator{Zone: xdsCtx.ControlPlane.Zone}
	for _, listenerHostname := range info.ListenerHostnames {
		for _, hostInfo := range listenerHostname.HostInfos {
			clusterRes, err := gen.GenerateClusters(ctx, xdsCtx, info, hostInfo.Entries(), hostInfo.Host.Tags)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to generate clusters for dataplane %q", info.Proxy.Id)
			}
			resources.AddSet(clusterRes)
		}
	}

	return resources, nil
}

func generateEnvoyRouteEntries(host plugin_gateway.GatewayHost, toRules rules.Rules) []route.Entry {
	var entries []route.Entry

	for _, rule := range toRules {
		var names []string
		for _, orig := range rule.Origin {
			names = append(names, orig.GetName())
		}
		slices.Sort(names)
		entries = append(entries, makeTcpRouteEntry(strings.Join(names, "_"), rule.Conf.(api.Rule)))
	}

	return plugin_gateway.HandlePrefixMatchesAndPopulatePolicies(host, nil, nil, entries)
}

func makeTcpRouteEntry(name string, rule api.Rule) route.Entry {
	entry := route.Entry{
		Route: name,
	}

	for _, b := range rule.Default.BackendRefs {
		dest, ok := tags.FromTargetRef(b.TargetRef)
		if !ok {
			// This should be caught by validation
			continue
		}
		target := route.Destination{
			Destination:   dest,
			Weight:        uint32(*b.Weight),
			Policies:      nil,
			RouteProtocol: core_mesh.ProtocolTCP,
		}

		entry.Action.Forward = append(entry.Action.Forward, target)
	}

	return entry
}
