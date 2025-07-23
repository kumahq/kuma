package v1alpha1

import (
	"time"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/pkg/errors"
	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/kri"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	"github.com/kumahq/kuma/pkg/core/resources/apis/core/destinationname"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	bldrs_clusters "github.com/kumahq/kuma/pkg/envoy/builders/cluster"
	. "github.com/kumahq/kuma/pkg/envoy/builders/common"
	bldrs_endpoint "github.com/kumahq/kuma/pkg/envoy/builders/endpoint"
	bldrs_listener "github.com/kumahq/kuma/pkg/envoy/builders/listener"
	bldrs_route "github.com/kumahq/kuma/pkg/envoy/builders/route"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	rules_outbound "github.com/kumahq/kuma/pkg/plugins/policies/core/rules/outbound"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/subsetutils"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancingstrategy/api/v1alpha1"
	gateway_plugin "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/metadata"
	util_maps "github.com/kumahq/kuma/pkg/util/maps"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_slices "github.com/kumahq/kuma/pkg/util/slices"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
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

	if err := p.configureGateway(
		ctx.Mesh,
		proxy,
		policies.GatewayRules,
		listeners.Gateway,
		clusters.Gateway,
		routes.Gateway,
		rs,
		ctx.Mesh.Resource.ZoneEgressEnabled(),
	); err != nil {
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
	if proxy.Dataplane.Spec.IsBuiltinGateway() {
		return nil
	}
	serviceConfs := map[string]api.Conf{}

	for _, outbound := range proxy.Outbounds.Filter(xds_types.NonBackendRefFilter) {
		oface := proxy.Dataplane.Spec.Networking.ToOutboundInterface(outbound.LegacyOutbound)
		serviceName := outbound.LegacyOutbound.GetService()

		computed := toRules.Rules.Compute(subsetutils.MeshServiceElement(serviceName))
		if computed == nil {
			continue
		}

		conf := computed.Conf.(api.Conf)

		if listener, ok := listeners.Outbound[oface]; ok {
			if err := NewModifier(listener).Configure(listenerConfigurer(rules_outbound.AsResourceContext(conf))).Modify(); err != nil {
				return err
			}
		}

		serviceConfs[serviceName] = conf
	}

	clusterModifier := func(cluster *envoy_cluster.Cluster, conf api.Conf) error {
		return NewModifier(cluster).
			Configure(clusterConfigurer(conf)).
			Configure(If(cluster.LoadAssignment != nil, staticCLAConfigurer(conf, proxy.Dataplane.Spec.TagSet(), proxy.Zone, meshCtx.Resource.ZoneEgressEnabled(), generator.OriginOutbound))).
			Modify()
	}

	// when VIPs are enabled 2 listeners are pointing to the same cluster, that's why
	// we configure clusters in a separate loop to avoid configuring the same cluster twice
	for serviceName, conf := range serviceConfs {
		if cluster, ok := clusters.Outbound[serviceName]; ok {
			if err := clusterModifier(cluster, conf); err != nil {
				return err
			}
		}
		for _, cluster := range clusters.OutboundSplit[serviceName] {
			if err := clusterModifier(cluster, conf); err != nil {
				return err
			}
		}
		for _, cla := range endpoints[serviceName] {
			if err := NewModifier(cla).Configure(claConfigurer(conf, proxy.Dataplane.Spec.TagSet(), proxy.Zone, meshCtx.Resource.ZoneEgressEnabled(), generator.OriginOutbound)).Modify(); err != nil {
				return err
			}
		}
	}

	rctx := rules_outbound.RootContext[api.Conf](meshCtx.Resource, toRules.ResourceRules)
	for _, r := range util_slices.Filter(rs.List(), core_xds.HasAssociatedServiceResource) {
		svcCtx := rctx.
			WithID(kri.NoSectionName(*r.ResourceOrigin)).
			WithID(*r.ResourceOrigin)
		if err := p.applyToRealResource(svcCtx, r, proxy); err != nil {
			return err
		}
	}

	return nil
}

func (p plugin) applyToRealResource(rctx *rules_outbound.ResourceContext[api.Conf], r *core_xds.Resource, proxy *core_xds.Proxy) error {
	switch envoyResource := r.Resource.(type) {
	case *envoy_listener.Listener:
		return NewModifier(envoyResource).
			Configure(listenerConfigurer(rctx)).
			Modify()
	case *envoy_cluster.Cluster:
		return NewModifier(envoyResource).
			Configure(clusterConfigurer(rctx.Conf())).
			Configure(If(envoyResource.LoadAssignment != nil, staticCLAConfigurer(rctx.Conf(), proxy.Dataplane.Spec.TagSet(), proxy.Zone, false, generator.OriginOutbound))).
			Modify()
	case *envoy_endpoint.ClusterLoadAssignment:
		return NewModifier(envoyResource).
			Configure(claConfigurer(rctx.Conf(), proxy.Dataplane.Spec.TagSet(), proxy.Zone, false, generator.OriginOutbound)).
			Modify()
	}
	return nil
}

func staticCLAConfigurer(conf api.Conf, tags mesh_proto.MultiValueTagSet, localZone string, egressEnabled bool, origin string) Configurer[envoy_cluster.Cluster] {
	return func(c *envoy_cluster.Cluster) error {
		return NewModifier(c.LoadAssignment).
			Configure(claConfigurer(conf, tags, localZone, egressEnabled, origin)).
			Modify()
	}
}

func listenerConfigurer(rctx *rules_outbound.ResourceContext[api.Conf]) Configurer[envoy_listener.Listener] {
	return bldrs_listener.FilterChains(filterChainConfigurer(rctx))
}

func filterChainConfigurer(rctx *rules_outbound.ResourceContext[api.Conf]) Configurer[envoy_listener.FilterChain] {
	return bldrs_listener.RoutesOnFilterChain(func(route *envoy_route.Route) error {
		var routeCtx *rules_outbound.ResourceContext[api.Conf]
		if routeID, err := kri.FromString(route.Name); err == nil {
			routeCtx = rctx.WithID(routeID)
		} else {
			routeCtx = rctx
		}
		return NewModifier(route).Configure(routeConfigurer(routeCtx)).Modify()
	})
}

func routeConfigurer(rctx *rules_outbound.ResourceContext[api.Conf]) Configurer[envoy_route.Route] {
	return IfNotNil(getHashPolicies(rctx.Conf()), func(hashPolicies []api.HashPolicy) Configurer[envoy_route.Route] {
		return bldrs_route.HashPolicies(util_slices.Map(hashPolicies, hashPolicy))
	})
}

func clusterConfigurer(conf api.Conf) Configurer[envoy_cluster.Cluster] {
	return func(cluster *envoy_cluster.Cluster) error {
		return NewModifier(cluster).
			Configure(If(shouldUseLocalityWeightedLb(conf), bldrs_clusters.LocalityWeightedLbConfigurer())).
			Configure(IfNotNil(conf.LoadBalancer, loadBalancerConfigurer)).
			Modify()
	}
}

func claConfigurer(conf api.Conf, tags mesh_proto.MultiValueTagSet, localZone string, egressEnabled bool, origin string) Configurer[envoy_endpoint.ClusterLoadAssignment] {
	return func(cla *envoy_endpoint.ClusterLoadAssignment) error {
		atLeastOneLocalityGroup := conf.LocalityAwareness != nil && (conf.LocalityAwareness.LocalZone != nil || conf.LocalityAwareness.CrossZone != nil)
		isLocalityAware := conf.LocalityAwareness == nil || !pointer.Deref(conf.LocalityAwareness.Disabled)
		return NewModifier(cla).
			Configure(bldrs_endpoint.NonLocalPriority(isLocalityAware, localZone)).
			Configure(If(atLeastOneLocalityGroup, bldrs_endpoint.Endpoints(NewEndpoints(cla.Endpoints, tags, pointer.To(conf), localZone, egressEnabled, origin)))).
			Configure(If(atLeastOneLocalityGroup, bldrs_endpoint.OverprovisioningFactor(overprovisioningFactor(conf)))).
			Modify()
	}
}

func loadBalancerConfigurer(lb api.LoadBalancer) Configurer[envoy_cluster.Cluster] {
	return func(cluster *envoy_cluster.Cluster) error {
		modifier := NewModifier(cluster)
		switch lb.Type {
		case api.RoundRobinType:
			modifier.
				Configure(bldrs_clusters.LbPolicy(envoy_cluster.Cluster_ROUND_ROBIN))
		case api.LeastRequestType:
			modifier.
				Configure(bldrs_clusters.LbPolicy(envoy_cluster.Cluster_LEAST_REQUEST)).
				Configure(IfNotNil(lb.LeastRequest, func(lr api.LeastRequest) Configurer[envoy_cluster.Cluster] {
					return bldrs_clusters.LeastRequestLbConfig(bldrs_clusters.NewLeastRequestConfig().
						Configure(IfNotNil(lr.ActiveRequestBias, bldrs_clusters.ActiveRequestBias)).
						Configure(IfNotNil(lr.ChoiceCount, bldrs_clusters.ChoiceCount)),
					)
				}))
		case api.RandomType:
			modifier.
				Configure(bldrs_clusters.LbPolicy(envoy_cluster.Cluster_RANDOM))
		case api.RingHashType:
			modifier.
				Configure(bldrs_clusters.LbPolicy(envoy_cluster.Cluster_RING_HASH)).
				Configure(IfNotNil(lb.RingHash, func(rh api.RingHash) Configurer[envoy_cluster.Cluster] {
					return bldrs_clusters.RingHashLbConfig(bldrs_clusters.NewRingHashConfig().
						Configure(IfNotNil(rh.MinRingSize, bldrs_clusters.MinRingSize)).
						Configure(IfNotNil(rh.MaxRingSize, bldrs_clusters.MaxRingSize)).
						Configure(IfNotNil(rh.HashFunction, func(hf api.HashFunctionType) Configurer[envoy_cluster.Cluster_RingHashLbConfig] {
							return bldrs_clusters.HashFunction(convertHashFunction(hf))
						})))
				}))
		case api.MaglevType:
			modifier.
				Configure(bldrs_clusters.LbPolicy(envoy_cluster.Cluster_MAGLEV)).
				Configure(IfNotNil(lb.Maglev, func(m api.Maglev) Configurer[envoy_cluster.Cluster] {
					return bldrs_clusters.MaglevLbConfig(bldrs_clusters.NewMaglevConfig().
						Configure(IfNotNil(m.TableSize, bldrs_clusters.TableSize)),
					)
				}))
		}
		return modifier.Modify()
	}
}

func convertHashFunction(hf api.HashFunctionType) envoy_cluster.Cluster_RingHashLbConfig_HashFunction {
	switch hf {
	case api.MurmurHash2Type:
		return envoy_cluster.Cluster_RingHashLbConfig_MURMUR_HASH_2
	case api.XXHashType:
		return envoy_cluster.Cluster_RingHashLbConfig_XX_HASH
	default:
		return envoy_cluster.Cluster_RingHashLbConfig_XX_HASH
	}
}

func hashPolicy(conf api.HashPolicy) *Builder[envoy_route.RouteAction_HashPolicy] {
	return bldrs_route.HashPolicy().
		Configure(IfNotNil(conf.Terminal, bldrs_route.Terminal)).
		Configure(
			If(conf.Type == api.HeaderType,
				IfNotNil(conf.Header, func(h api.Header) Configurer[envoy_route.RouteAction_HashPolicy] {
					return bldrs_route.HeaderPolicySpecifier(h.Name)
				}))).
		Configure(If(conf.Type == api.CookieType,
			IfNotNil(conf.Cookie, func(cookie api.Cookie) Configurer[envoy_route.RouteAction_HashPolicy] {
				return bldrs_route.CookiePolicySpecifier(cookie.Name, pointer.Deref(cookie.Path), getDurationOrNil(cookie.TTL))
			}))).
		Configure(If(conf.Type == api.ConnectionType,
			IfNotNil(conf.Connection, func(conn api.Connection) Configurer[envoy_route.RouteAction_HashPolicy] {
				return bldrs_route.ConnectionTypePolicySpecifier(pointer.Deref(conn.SourceIP))
			}))).
		Configure(If(conf.Type == api.QueryParameterType,
			IfNotNil(conf.QueryParameter, func(qp api.QueryParameter) Configurer[envoy_route.RouteAction_HashPolicy] {
				return bldrs_route.QueryPolicySpecifier(qp.Name)
			}))).
		Configure(If(conf.Type == api.FilterStateType,
			IfNotNil(conf.FilterState, func(fs api.FilterState) Configurer[envoy_route.RouteAction_HashPolicy] {
				return bldrs_route.FilterStatePolicySpecifier(fs.Key)
			})))
}

func getDurationOrNil(d *k8s.Duration) *time.Duration {
	if d == nil {
		return nil
	}
	return &d.Duration
}

func getHashPolicies(conf api.Conf) *[]api.HashPolicy {
	if conf.HashPolicies != nil {
		return conf.HashPolicies
	}

	if conf.LoadBalancer == nil {
		return nil
	}

	switch conf.LoadBalancer.Type {
	case api.RingHashType:
		if conf.LoadBalancer.RingHash == nil {
			return nil
		}
		return conf.LoadBalancer.RingHash.HashPolicies
	case api.MaglevType:
		if conf.LoadBalancer.Maglev == nil {
			return nil
		}
		return conf.LoadBalancer.Maglev.HashPolicies
	default:
		return nil
	}
}

func overprovisioningFactor(conf api.Conf) uint32 {
	if conf.LocalityAwareness == nil || conf.LocalityAwareness.CrossZone == nil || conf.LocalityAwareness.CrossZone.FailoverThreshold == nil {
		return defaultOverprovisioningFactor
	}
	val, err := common_api.NewDecimalFromIntOrString(conf.LocalityAwareness.CrossZone.FailoverThreshold.Percentage)
	if err != nil || val.IsZero() {
		return defaultOverprovisioningFactor
	}
	return uint32(100/val.InexactFloat64()) * 100
}

func (p plugin) configureGateway(
	meshCtx xds_context.MeshContext,
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
	resourcesByOrigin := rs.IndexByOrigin(core_xds.NonMeshExternalService)

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

		rctx := rules_outbound.RootContext[api.Conf](meshCtx.Resource, rules.ResourceRules)

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
					if localityConf := core_rules.ComputeConf[api.Conf](rules.Rules, subsetutils.MeshServiceElement(serviceName)); localityConf != nil {
						perServiceConfiguration[serviceName] = localityConf

						err := NewModifier(cluster).
							Configure(clusterConfigurer(*localityConf)).
							Configure(If(cluster.LoadAssignment != nil, staticCLAConfigurer(rctx.Conf(), proxy.Dataplane.Spec.TagSet(), proxy.Zone, egressEnabled, metadata.OriginGateway))).
							Modify()
						if err != nil {
							return err
						}

						for _, cla := range endpoints[serviceName] {
							if err := NewModifier(cla).Configure(claConfigurer(*localityConf, proxy.Dataplane.Spec.TagSet(), proxy.Zone, egressEnabled, metadata.OriginGateway)).Modify(); err != nil {
								return err
							}
						}
					}

					if dest.BackendRef == nil {
						continue
					}
					if realRef := dest.BackendRef.ResourceOrNil(); realRef != nil {
						svcCtx := rctx.
							WithID(kri.NoSectionName(*realRef)).
							WithID(*realRef)
						for _, rs := range resourcesByOrigin[*realRef] {
							for _, r := range rs {
								if err := p.applyToRealResource(svcCtx, r, proxy); err != nil {
									return err
								}
							}
						}
					}
				}
			}
		}
		for _, configuration := range perServiceConfiguration {
			if err := p.configureRDS(listener, gatewayRoutes, configuration); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p plugin) configureEgress(rs *core_xds.ResourceSet, proxy *core_xds.Proxy) error {
	indexed := rs.IndexByOrigin()
	endpoints := policies_xds.GatherEgressEndpoints(rs)
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
			for _, cla := range endpoints[clusterName] {
				if err := NewModifier(cla).Configure(claConfigurer(conf, mesh_proto.MultiValueTagSet{}, proxy.Zone, true, egress.OriginEgress)).Modify(); err != nil {
					return err
				}
			}
		}

		meshExternalServices := meshResources.ListOrEmpty(meshexternalservice_api.MeshExternalServiceType)
		for _, mes := range meshExternalServices.GetItems() {
			meshExtSvc := mes.(*meshexternalservice_api.MeshExternalServiceResource)
			destinationName := destinationname.MustResolve(false, meshExtSvc, meshExtSvc.Spec.Match)
			policies, ok := meshResources.Dynamic[destinationName]
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
							if err := NewModifier(cluster.Resource.(*envoy_cluster.Cluster)).Configure(clusterConfigurer(conf.Conf[0].(api.Conf))).Modify(); err != nil {
								return err
							}
						}
					}
				}

				for _, fc := range listeners.Egress.FilterChains {
					if fc.Name == destinationName {
						if err := NewModifier(fc).Configure(filterChainConfigurer(rules_outbound.AsResourceContext(conf.Conf[0].(api.Conf)))).Modify(); err != nil {
							return err
						}
					}
				}
			}
		}
	}

	return nil
}

// Zone egress is a single point for multiple clients. At this moment we don't support different
// configurations based on the client. That's why we are computing rules for MeshSubset
func (p plugin) computeFrom(fr core_rules.FromRules) *core_rules.Rule {
	rules := util_maps.AllValues(fr.Rules)
	if len(rules) == 0 {
		return nil
	}
	return rules[0].Compute(subsetutils.MeshElement())
}

func (p plugin) configureRDS(
	l *envoy_listener.Listener,
	routes map[string]*envoy_route.RouteConfiguration,
	conf *api.Conf,
) error {
	if conf == nil || conf.LoadBalancer == nil {
		return nil
	}

	routeConfigs := []string{}
	for _, chain := range l.FilterChains {
		for _, filter := range chain.Filters {
			if filter.Name != wellknown.HTTPConnectionManager {
				continue
			}
			var hcm *envoy_hcm.HttpConnectionManager
			if msg, err := filter.GetTypedConfig().UnmarshalNew(); err != nil {
				return err
			} else {
				hcm = msg.(*envoy_hcm.HttpConnectionManager)
			}
			rs, ok := hcm.RouteSpecifier.(*envoy_hcm.HttpConnectionManager_Rds)
			if !ok {
				return errors.Errorf("unexpected RouteSpecifer %T", hcm.RouteSpecifier)
			}
			routeConfigs = append(routeConfigs, rs.Rds.RouteConfigName)
		}
	}

	for _, rc := range routeConfigs {
		err := NewModifier(routes[rc]).
			Configure(bldrs_route.AllRoutes(routeConfigurer(rules_outbound.AsResourceContext(*conf)))).
			Modify()
		if err != nil {
			return err
		}
	}
	return nil
}

func shouldUseLocalityWeightedLb(config api.Conf) bool {
	return config.LocalityAwareness != nil && config.LocalityAwareness.LocalZone != nil && len(pointer.Deref(config.LocalityAwareness.LocalZone.AffinityTags)) > 0
}
