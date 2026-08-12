package v1alpha1

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"

	"github.com/kumahq/kuma/v3/pkg/core/kri"
	"github.com/kumahq/kuma/v3/pkg/core/naming"
	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	core_rules "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/outbound"
	policies_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshcircuitbreaker/api/v1alpha1"
	plugin_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshcircuitbreaker/plugin/xds"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
)

var _ core_plugins.PolicyPlugin = &plugin{}

type plugin struct{}

func (p plugin) Order() int { return api.MeshCircuitBreakerResourceTypeDescriptor.Order }

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) Apply(
	rs *core_xds.ResourceSet,
	ctx xds_context.Context,
	proxy *core_xds.Proxy,
) error {
	applyTrackRemaining(policies_xds.GatherAllClusters(rs))

	clusters := policies_xds.GatherClusters(rs)

	policies, ok := proxy.Policies.Dynamic[api.MeshCircuitBreakerType]
	if !ok {
		return nil
	}

	if err := applyToInbounds(policies.FromRules, clusters.Inbound, proxy.Dataplane); err != nil {
		return err
	}

	if err := applyToRealResources(ctx.Mesh, rs, policies.ToRules.ResourceRules); err != nil {
		return err
	}

	return nil
}

func applyTrackRemaining(clusters []*envoy_cluster.Cluster) {
	for _, cluster := range clusters {
		plugin_xds.EnsureTrackRemaining(cluster)
	}
}

func applyToInbounds(
	fromRules core_rules.FromRules,
	inboundClusters map[string]*envoy_cluster.Cluster,
	dataplane *core_mesh.DataplaneResource,
) error {
	return policies_xds.ForEachInbound[api.Conf](dataplane, fromRules, func(m policies_xds.InboundMatch[api.Conf]) error {
		clusterName := naming.MustContextualInboundName(dataplane, m.Interface.InboundName)
		cluster, ok := inboundClusters[clusterName]
		if !ok {
			return nil
		}

		return plugin_xds.NewConfigurer(m.Conf).ConfigureCluster(cluster)
	})
}

func applyToRealResource(
	meshCtx xds_context.MeshContext,
	rules outbound.ResourceRules,
	uri kri.Identifier,
	resourcesByType core_xds.ResourcesByType,
) error {
	conf := rules.Compute(uri, meshCtx.Resources)
	if conf == nil {
		return nil
	}

	for typ, resources := range resourcesByType {
		if typ == envoy_resource.ClusterType {
			err := configureClusters(resources, conf.Conf[0].(api.Conf))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func applyToRealResources(
	meshCtx xds_context.MeshContext,
	rs *core_xds.ResourceSet,
	rules outbound.ResourceRules,
	filters ...func(*core_xds.Resource) bool,
) error {
	for uri, resType := range rs.IndexByOrigin(filters...) {
		if err := applyToRealResource(meshCtx, rules, uri, resType); err != nil {
			return err
		}
	}
	return nil
}

func configureClusters(resources []*core_xds.Resource, conf api.Conf) error {
	for _, resource := range resources {
		configurer := plugin_xds.Configurer{
			Conf: conf,
		}
		err := configurer.ConfigureCluster(resource.Resource.(*envoy_cluster.Cluster))
		if err != nil {
			return err
		}
	}
	return nil
}
