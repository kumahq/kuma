package v1alpha1

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshcircuitbreaker/api/v1alpha1"
	plugin_xds "github.com/kumahq/kuma/pkg/plugins/policies/meshcircuitbreaker/plugin/xds"
	gateway "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
)

var _ core_plugins.EgressPolicyPlugin = &plugin{}

type plugin struct{}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(
	dataplane *core_mesh.DataplaneResource,
	resources xds_context.Resources,
	opts ...core_plugins.MatchedPoliciesOption,
) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshCircuitBreakerType, dataplane, resources, opts...)
}

func (p plugin) EgressMatchedPolicies(tags map[string]string, resources xds_context.Resources, opts ...core_plugins.MatchedPoliciesOption) (core_xds.TypedMatchingPolicies, error) {
	return matchers.EgressMatchedPolicies(api.MeshCircuitBreakerType, tags, resources, opts...)
}

func (p plugin) Apply(
	rs *core_xds.ResourceSet,
	ctx xds_context.Context,
	proxy *core_xds.Proxy,
) error {
	if proxy.ZoneEgressProxy != nil {
		return applyToEgressRealResources(rs, proxy)
	}
	policies, ok := proxy.Policies.Dynamic[api.MeshCircuitBreakerType]
	if !ok {
		return nil
	}

	clusters := policies_xds.GatherClusters(rs)

	if err := applyToInbounds(policies.FromRules, clusters.Inbound, proxy.Dataplane); err != nil {
		return err
	}

	if err := applyToOutbounds(policies.ToRules, clusters.Outbound, clusters.OutboundSplit, proxy.Outbounds); err != nil {
		return err
	}

	if err := applyToGateways(policies.GatewayRules, clusters.Gateway, proxy); err != nil {
		return err
	}

	if err := applyToRealResources(rs, policies.ToRules.ResourceRules, ctx.Mesh); err != nil {
		return err
	}

	return nil
}

func applyToInbounds(
	fromRules core_rules.FromRules,
	inboundClusters map[string]*envoy_cluster.Cluster,
	dataplane *core_mesh.DataplaneResource,
) error {
	for _, inbound := range dataplane.Spec.Networking.GetInbound() {
		iface := dataplane.Spec.Networking.ToInboundInterface(inbound)

		listenerKey := core_rules.InboundListener{
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

		if err := configure(rules, core_rules.MeshSubset(), cluster); err != nil {
			return err
		}
	}

	return nil
}

func applyToOutbounds(
	rules core_rules.ToRules,
	outboundClusters map[string]*envoy_cluster.Cluster,
	outboundSplitClusters map[string][]*envoy_cluster.Cluster,
	outbounds xds_types.Outbounds,
) error {
	targetedClusters := policies_xds.GatherTargetedClusters(
		outbounds,
		outboundSplitClusters,
		outboundClusters,
	)

	for cluster, serviceName := range targetedClusters {
		if err := configure(rules.Rules, core_rules.MeshService(serviceName), cluster); err != nil {
			return err
		}
	}

	return nil
}

func applyToGateways(
	gatewayRules core_rules.GatewayRules,
	gatewayClusters map[string]*envoy_cluster.Cluster,
	proxy *core_xds.Proxy,
) error {
	for _, listenerInfo := range gateway.ExtractGatewayListeners(proxy) {
		rules, ok := gatewayRules.ToRules.ByListener[core_rules.InboundListener{
			Address: proxy.Dataplane.Spec.GetNetworking().Address,
			Port:    listenerInfo.Listener.Port,
		}]
		if !ok {
			continue
		}
		for _, listenerHostnames := range listenerInfo.ListenerHostnames {
			for _, hostInfo := range listenerHostnames.HostInfos {
				destinations := gateway.RouteDestinationsMutable(hostInfo.Entries())
				for _, dest := range destinations {
					clusterName, err := dest.Destination.DestinationClusterName(hostInfo.Host.Tags)
					if err != nil {
						continue
					}
					cluster, ok := gatewayClusters[clusterName]
					if !ok {
						continue
					}

					serviceName := dest.Destination[mesh_proto.ServiceTag]

					if err := configure(
						rules,
						core_rules.MeshService(serviceName),
						cluster,
					); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func configure(
	rules core_rules.Rules,
	subset core_rules.Subset,
	cluster *envoy_cluster.Cluster,
) error {
	if computed := rules.Compute(subset); computed != nil {
		return plugin_xds.NewConfigurer(computed.Conf.(api.Conf)).ConfigureCluster(cluster)
	}

	return nil
}

func applyToEgressRealResources(rs *core_xds.ResourceSet, proxy *core_xds.Proxy) error {
	indexed := rs.IndexByOrigin()
	for _, meshResources := range proxy.ZoneEgressProxy.MeshResourcesList {
		meshExternalServices := meshResources.ListOrEmpty(meshexternalservice_api.MeshExternalServiceType)
		for _, mes := range meshExternalServices.GetItems() {
			meshExtSvc := mes.(*meshexternalservice_api.MeshExternalServiceResource)
			policies, ok := meshResources.Dynamic[meshExtSvc.DestinationName(uint32(meshExtSvc.Spec.Match.Port))]
			if !ok {
				continue
			}
			mhc, ok := policies[api.MeshCircuitBreakerType]
			if !ok {
				continue
			}
			for mesID, typedResources := range indexed {
				conf := mhc.ToRules.ResourceRules.Compute(mesID, meshResources)
				if conf == nil {
					continue
				}

				for typ, resources := range typedResources {
					switch typ {
					case envoy_resource.ClusterType:
						err := configureClusters(resources, conf.Conf[0].(api.Conf))
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}
	return nil
}

func applyToRealResources(rs *core_xds.ResourceSet, rules core_rules.ResourceRules, meshCtx xds_context.MeshContext) error {
	for uri, resType := range rs.IndexByOrigin() {
		conf := rules.Compute(uri, meshCtx.Resources)
		if conf == nil {
			continue
		}

		for typ, resources := range resType {
			switch typ {
			case envoy_resource.ClusterType:
				err := configureClusters(resources, conf.Conf[0].(api.Conf))
				if err != nil {
					return err
				}
			}
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
