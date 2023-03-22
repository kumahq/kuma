package v1alpha1

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/matchers"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancingstrategy/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancingstrategy/plugin/xds"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
)

var _ core_plugins.PolicyPlugin = &plugin{}

type plugin struct{}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshLoadBalancingStrategyType, dataplane, resources)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	policies, ok := proxy.Policies.Dynamic[api.MeshLoadBalancingStrategyType]
	if !ok {
		return nil
	}

	listeners := policies_xds.GatherListeners(rs)
	clusters := policies_xds.GatherClusters(rs)

	serviceConfs := map[string]api.Conf{}

	for _, outbound := range proxy.Dataplane.Spec.Networking.GetOutbound() {
		oface := proxy.Dataplane.Spec.Networking.ToOutboundInterface(outbound)
		serviceName := outbound.GetTagsIncludingLegacy()[mesh_proto.ServiceTag]

		computed := policies.ToRules.Rules.Compute(core_xds.MeshService(serviceName))
		if computed == nil {
			continue
		}

		conf := computed.Conf.(api.Conf)

		if listener, ok := listeners.Outbound[oface]; ok {
			if err := p.generateLDS(listener, conf.LoadBalancer); err != nil {
				return err
			}
		}

		serviceConfs[serviceName] = conf
	}

	// when VIPs are enabled 2 listeners are pointing to the same cluster, that's why
	// we configure clusters in a separate loop to avoid configuring the same cluster twice
	for serviceName, conf := range serviceConfs {
		if cluster, ok := clusters.Outbound[serviceName]; ok {
			if err := p.generateCDS(cluster, conf.LoadBalancer); err != nil {
				return err
			}
		}
		for _, cluster := range clusters.OutboundSplit[serviceName] {
			if err := p.generateCDS(cluster, conf.LoadBalancer); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p plugin) generateLDS(l *envoy_listener.Listener, lbConf *api.LoadBalancer) error {
	if lbConf == nil {
		return nil
	}

	var hashPolicy *[]api.HashPolicy

	switch lbConf.Type {
	case api.RingHashType:
		if lbConf.RingHash == nil {
			return nil
		}
		hashPolicy = lbConf.RingHash.HashPolicies
	case api.MaglevType:
		if lbConf.Maglev == nil {
			return nil
		}
		hashPolicy = lbConf.Maglev.HashPolicies
	default:
		return nil
	}

	if l.FilterChains == nil || len(l.FilterChains) != 1 {
		return errors.Errorf("expected exactly one filter chain, got %d", len(l.FilterChains))
	}

	return v3.UpdateHTTPConnectionManager(l.FilterChains[0], func(hcm *envoy_hcm.HttpConnectionManager) error {
		rc := hcm.RouteSpecifier.(*envoy_hcm.HttpConnectionManager_RouteConfig).RouteConfig
		hpc := &xds.HashPolicyConfigurer{HashPolicies: *hashPolicy}

		for _, vh := range rc.VirtualHosts {
			for _, route := range vh.Routes {
				if err := hpc.Configure(route); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (p plugin) generateCDS(c *envoy_cluster.Cluster, lbConf *api.LoadBalancer) error {
	if lbConf == nil {
		return nil
	}
	return (&xds.LoadBalancerConfigurer{LoadBalancer: *lbConf}).Configure(c)
}
