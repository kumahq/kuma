package v1alpha1

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/pkg/errors"
	"golang.org/x/exp/maps"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancingstrategy/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancingstrategy/plugin/xds"
	gateway_plugin "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/metadata"
	"github.com/kumahq/kuma/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	"github.com/kumahq/kuma/pkg/xds/generator"
	"github.com/kumahq/kuma/pkg/xds/generator/egress"
)

var _ core_plugins.EgressPolicyPlugin = &plugin{}

type plugin struct{}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources, opts ...core_plugins.MatchedPoliciesOption) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshLoadBalancingStrategyType, dataplane, resources, opts...)
}

func (p plugin) EgressMatchedPolicies(tags map[string]string, resources xds_context.Resources, opts ...core_plugins.MatchedPoliciesOption) (core_xds.TypedMatchingPolicies, error) {
	return matchers.EgressMatchedPolicies(api.MeshLoadBalancingStrategyType, tags, resources, opts...)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	if proxy.ZoneEgressProxy != nil {
		return p.configureEgress(rs, proxy)
	}

	policies, ok := proxy.Policies.Dynamic[api.MeshLoadBalancingStrategyType]
	if !ok {
		return nil
	}

	listeners := policies_xds.GatherListeners(rs)
	clusters := policies_xds.GatherClusters(rs)
	endpoints := policies_xds.GatherOutboundEndpoints(rs)
	routes := policies_xds.GatherRoutes(rs)

	if err := p.configureGateway(proxy, policies.GatewayRules, listeners.Gateway, clusters.Gateway, routes.Gateway, rs, ctx.Mesh.Resource.ZoneEgressEnabled()); err != nil {
		return err
	}

	return p.configureDPP(proxy, policies.ToRules, listeners, clusters, endpoints, rs, ctx.Mesh.Resource.ZoneEgressEnabled())
}

func (p plugin) configureDPP(
	proxy *core_xds.Proxy,
	toRules core_rules.ToRules,
	listeners policies_xds.Listeners,
	clusters policies_xds.Clusters,
	endpoints policies_xds.EndpointMap,
	rs *core_xds.ResourceSet,
	egressEnabled bool,
) error {
	serviceConfs := map[string]api.Conf{}

	for _, outbound := range proxy.Dataplane.Spec.Networking.GetOutbounds(mesh_proto.NonBackendRefFilter) {
		oface := proxy.Dataplane.Spec.Networking.ToOutboundInterface(outbound)
		serviceName := outbound.GetService()

		computed := toRules.Rules.Compute(core_rules.MeshServiceElement(serviceName))
		if computed == nil {
			continue
		}

		conf := computed.Conf.(api.Conf)

		if listener, ok := listeners.Outbound[oface]; ok {
			if err := p.configureListener(listener, nil, &conf); err != nil {
				return err
			}
		}

		serviceConfs[serviceName] = conf
	}

	// when VIPs are enabled 2 listeners are pointing to the same cluster, that's why
	// we configure clusters in a separate loop to avoid configuring the same cluster twice
	for serviceName, conf := range serviceConfs {
		if cluster, ok := clusters.Outbound[serviceName]; ok {
			if err := p.configureCluster(cluster, conf); err != nil {
				return err
			}
			if err := configureEndpoints(proxy.Dataplane.Spec.TagSet(), cluster, endpoints[serviceName], serviceName, conf, rs, proxy.Zone, proxy.APIVersion, egressEnabled, generator.OriginOutbound); err != nil {
				return errors.Wrapf(err, "failed to configure ClusterLoadAssignment for %s", serviceName)
			}
		}
		for _, cluster := range clusters.OutboundSplit[serviceName] {
			if err := p.configureCluster(cluster, conf); err != nil {
				return err
			}
			if err := configureEndpoints(proxy.Dataplane.Spec.TagSet(), cluster, endpoints[serviceName], cluster.Name, conf, rs, proxy.Zone, proxy.APIVersion, egressEnabled, generator.OriginOutbound); err != nil {
				return errors.Wrapf(err, "failed to configure ClusterLoadAssignment for %s", cluster.Name)
			}
		}
	}

	return nil
}

func configureEndpoints(
	tags mesh_proto.MultiValueTagSet,
	cluster *envoy_cluster.Cluster,
	endpoints []*envoy_endpoint.ClusterLoadAssignment,
	serviceName string,
	conf api.Conf,
	rs *core_xds.ResourceSet,
	localZone string,
	apiVersion core_xds.APIVersion,
	egressEnabled bool,
	origin string,
) error {
	if cluster.LoadAssignment != nil {
		if err := ConfigureStaticEndpointsLocalityAware(tags, endpoints, cluster, conf, serviceName, localZone, apiVersion, egressEnabled, origin); err != nil {
			return err
		}
	} else {
		if err := ConfigureEndpointsLocalityAware(tags, endpoints, conf, rs, serviceName, localZone, apiVersion, egressEnabled, origin); err != nil {
			return err
		}
	}

	if conf.LocalityAwareness == nil || !pointer.Deref(conf.LocalityAwareness.Disabled) {
		for _, cla := range endpoints {
			for _, localityLbEndpoints := range cla.Endpoints {
				if localityLbEndpoints.Locality != nil && localityLbEndpoints.Locality.Zone != localZone {
					localityLbEndpoints.Priority = 1
				}
			}
		}
	}
	return nil
}

func (p plugin) configureGateway(
	proxy *core_xds.Proxy,
	rules core_rules.GatewayRules,
	gatewayListeners map[core_rules.InboundListener]*envoy_listener.Listener,
	gatewayClusters map[string]*envoy_cluster.Cluster,
	gatewayRoutes map[string]*envoy_route.RouteConfiguration,
	rs *core_xds.ResourceSet,
	egressEnabled bool,
) error {
	gatewayListenerInfos := gateway_plugin.ExtractGatewayListeners(proxy)
	if len(gatewayListenerInfos) == 0 {
		return nil
	}

	endpoints := policies_xds.GatherGatewayEndpoints(rs)

	for _, listenerInfo := range gatewayListenerInfos {
		inboundListener := core_rules.InboundListener{
			Address: proxy.Dataplane.Spec.GetNetworking().GetAddress(),
			Port:    listenerInfo.Listener.Port,
		}

		listener, ok := gatewayListeners[inboundListener]
		if !ok {
			continue
		}

		rules, ok := rules.ToRules.ByListener[inboundListener]
		if !ok {
			continue
		}

		perServiceConfiguration := map[string]*api.Conf{}
		for _, listenerHostnames := range listenerInfo.ListenerHostnames {
			for _, hostInfo := range listenerHostnames.HostInfos {
				destinations := gateway_plugin.RouteDestinationsMutable(hostInfo.Entries())
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
					localityConf := core_rules.ComputeConf[api.Conf](rules, core_rules.MeshServiceElement(serviceName))
					if localityConf == nil {
						continue
					}
					perServiceConfiguration[serviceName] = localityConf

					if err := p.configureCluster(cluster, *localityConf); err != nil {
						return err
					}

					if err := configureEndpoints(proxy.Dataplane.Spec.TagSet(), cluster, endpoints[serviceName], clusterName, *localityConf, rs, proxy.Zone, proxy.APIVersion, egressEnabled, metadata.OriginGateway); err != nil {
						return err
					}
				}
			}
		}
		for _, configuration := range perServiceConfiguration {
			if err := p.configureListener(listener, gatewayRoutes, configuration); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p plugin) configureEgress(rs *core_xds.ResourceSet, proxy *core_xds.Proxy) error {
	endpoints := policies_xds.GatherEgressEndpoints(rs)
	clusters := policies_xds.GatherClusters(rs)

	for _, mr := range proxy.ZoneEgressProxy.MeshResourcesList {
		for serviceName, dynamic := range mr.Dynamic {
			meshName := mr.Mesh.GetMeta().GetName()
			policies, ok := dynamic[api.MeshLoadBalancingStrategyType]
			if !ok {
				continue
			}

			rule := p.computeFrom(policies.FromRules)
			if rule == nil {
				continue
			}
			conf := rule.Conf.(api.Conf)

			clusterName := envoy_names.GetMeshClusterName(meshName, serviceName)
			err := configureEndpoints(mesh_proto.MultiValueTagSet{}, clusters.Egress[clusterName], endpoints[clusterName], clusterName, conf, rs, proxy.Zone, proxy.APIVersion, true, egress.OriginEgress)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Zone egress is a single point for multiple clients. At this moment we don't support different
// configurations based on the client. That's why we are computing rules for MeshSubset
func (p plugin) computeFrom(fr core_rules.FromRules) *core_rules.Rule {
	rules := maps.Values(fr.Rules)
	if len(rules) == 0 {
		return nil
	}
	return rules[0].Compute(core_rules.MeshElement())
}

func (p plugin) configureListener(
	l *envoy_listener.Listener,
	routes map[string]*envoy_route.RouteConfiguration,
	conf *api.Conf,
) error {
	if conf == nil || conf.LoadBalancer == nil {
		return nil
	}

	var hashPolicy *[]api.HashPolicy

	switch conf.LoadBalancer.Type {
	case api.RingHashType:
		if conf.LoadBalancer.RingHash == nil {
			return nil
		}
		hashPolicy = conf.LoadBalancer.RingHash.HashPolicies
	case api.MaglevType:
		if conf.LoadBalancer.Maglev == nil {
			return nil
		}
		hashPolicy = conf.LoadBalancer.Maglev.HashPolicies
	default:
		return nil
	}

	if l.FilterChains == nil {
		return errors.New("expected at least one filter chain")
	}

	for _, chain := range l.FilterChains {
		err := v3.UpdateHTTPConnectionManager(chain, func(hcm *envoy_hcm.HttpConnectionManager) error {
			var routeConfig *envoy_route.RouteConfiguration
			switch r := hcm.RouteSpecifier.(type) {
			case *envoy_hcm.HttpConnectionManager_RouteConfig:
				routeConfig = r.RouteConfig
			case *envoy_hcm.HttpConnectionManager_Rds:
				routeConfig = routes[r.Rds.RouteConfigName]
			default:
				return errors.Errorf("unexpected RouteSpecifer %T", r)
			}

			hpc := &xds.HashPolicyConfigurer{HashPolicies: *hashPolicy}
			for _, vh := range routeConfig.VirtualHosts {
				for _, route := range vh.Routes {
					if err := hpc.Configure(route); err != nil {
						return err
					}
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (p plugin) configureCluster(c *envoy_cluster.Cluster, config api.Conf) error {
	if shouldUseLocalityWeightedLb(config) {
		if err := (&xds.LocalityWeightedLbConfigurer{}).Configure(c); err != nil {
			return err
		}
	}
	if config.LoadBalancer != nil {
		if err := (&xds.LoadBalancerConfigurer{LoadBalancer: *config.LoadBalancer}).Configure(c); err != nil {
			return err
		}
	}
	return nil
}

func shouldUseLocalityWeightedLb(config api.Conf) bool {
	return config.LocalityAwareness != nil && config.LocalityAwareness.LocalZone != nil && len(pointer.Deref(config.LocalityAwareness.LocalZone.AffinityTags)) > 0
}
