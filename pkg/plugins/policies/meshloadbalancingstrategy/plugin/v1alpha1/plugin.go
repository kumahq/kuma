package v1alpha1

import (
	"strings"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/pkg/errors"
	"golang.org/x/exp/maps"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
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
	"github.com/kumahq/kuma/pkg/xds/envoy/tls"
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

	return p.configureDPP(
		proxy,
		policies.ToRules,
		listeners,
		clusters,
		endpoints,
		rs,
		ctx.Mesh,
	)
}

func (p plugin) configureDPP(
	proxy *core_xds.Proxy,
	toRules core_rules.ToRules,
	listeners policies_xds.Listeners,
	clusters policies_xds.Clusters,
	endpoints policies_xds.EndpointMap,
	rs *core_xds.ResourceSet,
	meshCtx xds_context.MeshContext,
) error {
	serviceConfs := map[string]api.Conf{}

	for _, outbound := range proxy.Outbounds.Filter(xds_types.NonBackendRefFilter) {
		oface := proxy.Dataplane.Spec.Networking.ToOutboundInterface(outbound.LegacyOutbound)
		serviceName := outbound.LegacyOutbound.GetService()

		computed := toRules.Rules.Compute(core_rules.MeshService(serviceName))
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
			if err := configureEndpoints(proxy.Dataplane.Spec.TagSet(), cluster, endpoints[serviceName], serviceName, conf, rs, proxy.Zone, proxy.APIVersion, meshCtx.Resource.ZoneEgressEnabled(), generator.OriginOutbound); err != nil {
				return errors.Wrapf(err, "failed to configure ClusterLoadAssignment for %s", serviceName)
			}
		}
		for _, cluster := range clusters.OutboundSplit[serviceName] {
			if err := p.configureCluster(cluster, conf); err != nil {
				return err
			}
			if err := configureEndpoints(proxy.Dataplane.Spec.TagSet(), cluster, endpoints[serviceName], cluster.Name, conf, rs, proxy.Zone, proxy.APIVersion, meshCtx.Resource.ZoneEgressEnabled(), generator.OriginOutbound); err != nil {
				return errors.Wrapf(err, "failed to configure ClusterLoadAssignment for %s", cluster.Name)
			}
		}
	}

	if err := p.applyToRealResources(proxy, endpoints, rs, toRules.ResourceRules, meshCtx); err != nil {
		return err
	}

	return nil
}

func (p plugin) applyToRealResources(
	proxy *core_xds.Proxy,
	endpoints policies_xds.EndpointMap,
	rs *core_xds.ResourceSet,
	rules core_rules.ResourceRules,
	meshCtx xds_context.MeshContext,
) error {
	for uri, resType := range rs.IndexByOrigin() {
		conf := rules.Compute(uri, meshCtx.Resources)
		if conf == nil {
			continue
		}
		apiConf := conf.Conf[0].(api.Conf)

		for typ, resources := range resType {
			switch typ {
			case envoy_resource.ListenerType:
				for _, resource := range resources {
					if resource.Origin != generator.OriginOutbound {
						continue
					}
					if err := p.configureListener(resource.Resource.(*envoy_listener.Listener), nil, &apiConf); err != nil {
						return err
					}
				}
			case envoy_resource.ClusterType:
				for _, resource := range resources {
					if resource.Origin != generator.OriginOutbound {
						continue
					}
					cluster := resource.Resource.(*envoy_cluster.Cluster)
					if err := p.configureCluster(cluster, apiConf); err != nil {
						return err
					}
					if err := configureEndpoints(proxy.Dataplane.Spec.TagSet(), cluster, endpoints[cluster.Name], cluster.Name, apiConf, rs, proxy.Zone, proxy.APIVersion, false, generator.OriginOutbound); err != nil {
						return errors.Wrapf(err, "failed to configure ClusterLoadAssignment for %s", cluster.Name)
					}
				}
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
	if cluster == nil {
		return nil
	}
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
					localityConf := core_rules.ComputeConf[api.Conf](rules, core_rules.MeshService(serviceName))
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
	indexed := rs.IndexByOrigin()
	endpoints := policies_xds.GatherEgressEndpoints(rs)
	clusters := policies_xds.GatherClusters(rs)
	listeners := policies_xds.GatherListeners(rs)
	if listeners.Egress == nil {
		return nil
	}
	for _, meshResources := range proxy.ZoneEgressProxy.MeshResourcesList {
		for serviceName, dynamic := range meshResources.Dynamic {
			meshName := meshResources.Mesh.GetMeta().GetName()
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

		meshExternalServices := meshResources.ListOrEmpty(meshexternalservice_api.MeshExternalServiceType)
		for _, mes := range meshExternalServices.GetItems() {
			meshExtSvc := mes.(*meshexternalservice_api.MeshExternalServiceResource)
			policies, ok := meshResources.Dynamic[meshExtSvc.DestinationName(uint32(meshExtSvc.Spec.Match.Port))]
			if !ok {
				continue
			}
			mlbs, ok := policies[api.MeshLoadBalancingStrategyType]
			if !ok {
				continue
			}
			for mesID, typedResources := range indexed {
				conf := mlbs.ToRules.ResourceRules.Compute(mesID, meshResources)
				if conf == nil {
					continue
				}

				for typ, resources := range typedResources {
					switch typ {
					case envoy_resource.ClusterType:
						for _, cluster := range resources {
							err := p.configureCluster(cluster.Resource.(*envoy_cluster.Cluster), conf.Conf[0].(api.Conf))
							if err != nil {
								return err
							}
						}
					}
				}
				err := p.configureEgressListener(listeners.Egress, conf.Conf[0].(api.Conf), mesID.Name, mesID.Mesh, uint32(meshExtSvc.Spec.Match.Port))
				if err != nil {
					return err
				}
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
	return rules[0].Compute(core_rules.MeshSubset())
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

func (p plugin) configureEgressListener(
	l *envoy_listener.Listener,
	conf api.Conf,
	name string,
	meshName string,
	port uint32,
) error {
	if conf.LoadBalancer == nil {
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

	sni := tls.SNIForResource(name, meshName, meshexternalservice_api.MeshExternalServiceType, port, nil)
	for _, chain := range l.FilterChains {
		matched := false
		for _, serverName := range chain.FilterChainMatch.ServerNames {
			if strings.Contains(serverName, sni) {
				matched = true
			}
		}
		if !matched {
			continue
		}
		err := v3.UpdateHTTPConnectionManager(chain, func(hcm *envoy_hcm.HttpConnectionManager) error {
			var routeConfig *envoy_route.RouteConfiguration
			switch r := hcm.RouteSpecifier.(type) {
			case *envoy_hcm.HttpConnectionManager_RouteConfig:
				routeConfig = r.RouteConfig
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
