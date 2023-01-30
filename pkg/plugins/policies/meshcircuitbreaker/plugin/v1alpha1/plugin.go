package v1alpha1

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"

	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/matchers"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshcircuitbreaker/api/v1alpha1"
	plugin_xds "github.com/kumahq/kuma/pkg/plugins/policies/meshcircuitbreaker/plugin/xds"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
)

var _ core_plugins.PolicyPlugin = &plugin{}

type plugin struct {
}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(
	dataplane *core_mesh.DataplaneResource,
	resources xds_context.Resources,
) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshCircuitBreakerType, dataplane, resources)
}

func (p plugin) Apply(
	rs *core_xds.ResourceSet,
	ctx xds_context.Context,
	proxy *core_xds.Proxy,
) error {
	policies, ok := proxy.Policies.Dynamic[api.MeshCircuitBreakerType]
	if !ok {
		return nil
	}

	clusters := policies_xds.GatherClusters(rs)

	if err := applyToInbounds(policies.FromRules, clusters.Inbound, proxy.Dataplane); err != nil {
		return err
	}

	if err := applyToOutbounds(policies.ToRules, clusters.Outbound, clusters.OutboundSplit, proxy.Dataplane); err != nil {
		return err
	}

	return nil
}

func applyToInbounds(
	fromRules core_xds.FromRules,
	inboundClusters map[string]*envoy_cluster.Cluster,
	dataplane *core_mesh.DataplaneResource,
) error {
	for _, inbound := range dataplane.Spec.Networking.GetInbound() {
		iface := dataplane.Spec.Networking.ToInboundInterface(inbound)

		listenerKey := core_xds.InboundListener{
			Address: iface.DataplaneIP,
			Port:    iface.DataplanePort,
		}

		rules, ok := fromRules.Rules[listenerKey]
		if !ok {
			continue
		}

		cluster, ok := inboundClusters[envoy_names.GetLocalClusterName(iface.DataplanePort)]
		if !ok {
			continue
		}

		if err := configure(rules, core_xds.MeshSubset(), cluster); err != nil {
			return err
		}
	}

	return nil
}

func applyToOutbounds(
	rules core_xds.ToRules,
	outboundClusters map[string]*envoy_cluster.Cluster,
	outboundSplitClusters map[string][]*envoy_cluster.Cluster,
	dataplane *core_mesh.DataplaneResource,
) error {
	targetedClusters := policies_xds.GatherTargetedClusters(dataplane.Spec.Networking.GetOutbound(), outboundSplitClusters, outboundClusters)

	for cluster, serviceName := range *targetedClusters {
		if err := configure(rules.Rules, core_xds.MeshService(serviceName), cluster); err != nil {
			return err
		}
	}

	return nil
}

func configure(
	rules core_xds.Rules,
	subset core_xds.Subset,
	cluster *envoy_cluster.Cluster,
) error {
	if computed := rules.Compute(subset); computed != nil {
		return plugin_xds.NewConfigurer(computed.Conf.(api.Conf)).ConfigureCluster(cluster)
	}

	return nil
}
