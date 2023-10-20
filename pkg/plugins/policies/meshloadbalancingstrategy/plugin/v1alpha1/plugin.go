package v1alpha1

import (
	"context"
	"fmt"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancingstrategy/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancingstrategy/plugin/xds"
	gateway_plugin "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	"github.com/kumahq/kuma/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_endpoints "github.com/kumahq/kuma/pkg/xds/envoy/endpoints"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
	envoy_metadata "github.com/kumahq/kuma/pkg/xds/envoy/metadata/v3"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
)

var _ core_plugins.EgressPolicyPlugin = &plugin{}

type plugin struct{}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshLoadBalancingStrategyType, dataplane, resources)
}

func (p plugin) EgressMatchedPolicies(es *core_mesh.ExternalServiceResource, resources xds_context.Resources) (core_xds.TypedMatchingPolicies, error) {
	return matchers.EgressMatchedPolicies(api.MeshLoadBalancingStrategyType, es, resources)
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
	endpoints := policies_xds.GatherEndpoints(rs)
	routes := policies_xds.GatherRoutes(rs)
	endpointsCopy := policies_xds.GatherEndpointsCopy(rs)

	if err := p.configureGateway(ctx, proxy, policies.ToRules, listeners.Gateway, clusters.Gateway, routes.Gateway, endpoints, rs); err != nil {
		return err
	}

	return p.configureDPP(proxy, policies.ToRules, listeners, clusters, endpointsCopy, ctx, rs)
}

func (p plugin) configureDPP(
	proxy *core_xds.Proxy,
	toRules core_rules.ToRules,
	listeners policies_xds.Listeners,
	clusters policies_xds.Clusters,
	endpoints policies_xds.EndpointMap,
	ctx xds_context.Context,
	rs *core_xds.ResourceSet,
) error {
	serviceConfs := map[string]api.Conf{}

	for _, outbound := range proxy.Dataplane.Spec.Networking.GetOutbound() {
		oface := proxy.Dataplane.Spec.Networking.ToOutboundInterface(outbound)
		serviceName := outbound.GetService()

		computed := toRules.Rules.Compute(core_rules.MeshService(serviceName))
		if computed == nil {
			continue
		}

		conf := computed.Conf.(api.Conf)

		if listener, ok := listeners.Outbound[oface]; ok {
			if err := p.configureListener(listener, nil, conf.LoadBalancer); err != nil {
				return err
			}
		}

		serviceConfs[serviceName] = conf
	}

	// when VIPs are enabled 2 listeners are pointing to the same cluster, that's why
	// we configure clusters in a separate loop to avoid configuring the same cluster twice
	for serviceName, conf := range serviceConfs {
		if cluster, ok := clusters.Outbound[serviceName]; ok {
			if err := p.configureCluster(cluster, &conf); err != nil {
				return err
			}
		}
		for _, cluster := range clusters.OutboundSplit[serviceName] {
			if err := p.configureCluster(cluster, &conf); err != nil {
				return err
			}
		}
		if err:= configureEndpoints(proxy, endpoints, serviceName, conf, ctx, rs); err != nil {
			return err
		}
	}

	return nil
}

type LocalityLbGroups struct {
	Key string
	Value string
	Weight uint32
}

type CrossZoneLocalityLbGroups struct {
	Disabled bool
	AllZones bool
	Zones map[string]bool
	ExceptZones map[string]bool
	Priority uint32
}

func configureEndpoints(
	proxy *core_xds.Proxy,
	endpoints policies_xds.EndpointMap,
	serviceName string,
	conf api.Conf,
	ctx xds_context.Context,
	rs *core_xds.ResourceSet,
) error {
	var localZone string
	dataplane := proxy.Dataplane
	if inbounds := dataplane.Spec.GetNetworking().GetInbound(); len(inbounds) != 0 {
		localZone = inbounds[0].GetTags()[mesh_proto.ZoneTag]
	}

	// matchingTags := map[string]string{}

	localZonePriority := []LocalityLbGroups{}

	// each endpoint iterate
	crossZonePriority := []CrossZoneLocalityLbGroups{}
	if conf.LocalityAwareness != nil && conf.LocalityAwareness.LocalZone != nil {
		tagsSet := dataplane.Spec.TagSet()
		for _, tag := range conf.LocalityAwareness.LocalZone.AffinityTags {
			values := tagsSet.Values(tag.Key)
			if len(tagsSet.Values(tag.Key)) != 0 {
				localZonePriority = append(localZonePriority, LocalityLbGroups{
					Key: tag.Key,
					Value: values[0],
					Weight: uint32(tag.Weight.IntVal),
					// Zone: zone,
				})
			}
		}
		if conf.LocalityAwareness.CrossZone != nil && len(conf.LocalityAwareness.CrossZone.Failover) > 0 {
			for priority, rule := range conf.LocalityAwareness.CrossZone.Failover{
				lb := CrossZoneLocalityLbGroups{}
				if rule.From != nil {
					doesRuleApply := false
					for _, zone := range rule.From.Zones{
						if zone == localZone {
							doesRuleApply = true
						}
					}
					if !doesRuleApply {
						continue
					}
				}
				switch rule.To.Type{
				case api.Any:
					lb.Disabled = false
					lb.AllZones = true
					lb.Priority = uint32(priority + 1)
				case api.AnyExcept:
					lb.Disabled = false
					lb.AllZones = false
					exceptZones := map[string]bool{}
					for _, zone := range rule.To.Zones {
						exceptZones[zone] = true
					}
					lb.ExceptZones = exceptZones
					lb.Priority = uint32(priority + 1)
				case api.Only:
					lb.Disabled = false
					lb.AllZones = false
					onlyZones := map[string]bool{}
					for _, zone := range rule.To.Zones {
						onlyZones[zone] = true
					}
					lb.Zones = onlyZones
					lb.Priority = uint32(priority + 1)
				default: 
					lb.Disabled = true
				}
				crossZonePriority = append(crossZonePriority, lb)
			}	
		}
		
	}

	
	// if is nil keep in the zone
	if conf.LocalityAwareness != nil {
		endpointsList := []core_xds.Endpoint{}
		for _, endpoint := range endpoints[serviceName]{
			for _, localityLbEndpoint := range endpoint.Endpoints {
				for _, lbEndpoint := range localityLbEndpoint.LbEndpoints{
					ed := core_xds.Endpoint{}
					ed.Weight = lbEndpoint.LoadBalancingWeight.GetValue()
					// check if has first tag, if yes set value as subzone and weight
					// if no go to another
					// if has no tag set default weight 1
					// if is cross zone set 
					tags := envoy_metadata.ExtractLbTags(lbEndpoint.Metadata)
					//iterate over local and if empty put all in the same locality
					skipEndpoint := false
					if localityLbEndpoint.Locality == nil || localityLbEndpoint.Locality.Zone == localZone {
						for _, localRule := range localZonePriority {
							val, ok := tags[localRule.Key]
							if ok && val == localRule.Value {
								ed.Locality = &core_xds.Locality{
									Zone: localZone,
									SubZone: fmt.Sprintf("%s=%s", localRule.Key, val),
									Weight: localRule.Weight,
									Priority: 0,
								}
								break
							}
							
						}
						if ed.Locality == nil {
							ed.Locality = &core_xds.Locality{
								Zone: localZone,
								Weight: 1,
								Priority: 0,
							}
						}
					} else {
						for _, zoneRule := range crossZonePriority {
							if zoneRule.Disabled {
								skipEndpoint = true
								break
							}
							if zoneRule.AllZones {
								ed.Locality = &core_xds.Locality{
									Zone: localityLbEndpoint.Locality.Zone,
									Priority: zoneRule.Priority,
								}
								break
							}
							if len(zoneRule.ExceptZones) > 0{
								_, ok := zoneRule.ExceptZones[tags[mesh_proto.ZoneTag]]
								if ok {
									continue
								} else {
									ed.Locality = &core_xds.Locality{
										Zone: localityLbEndpoint.Locality.Zone,
										Priority: zoneRule.Priority,
									}
									break
								}
							}
							if len(zoneRule.Zones) > 0{
								_, ok := zoneRule.Zones[tags[mesh_proto.ZoneTag]]
								if ok {
									ed.Locality = &core_xds.Locality{
										Zone: localityLbEndpoint.Locality.Zone,
										Priority: zoneRule.Priority,
									}
									break
								}
							}
						}
						if ed.Locality == nil {
							skipEndpoint = true
						}
					}
					if skipEndpoint {
						continue
					}
					
					//iterate over crosszone
					
					ed.Tags = tags
					address := lbEndpoint.GetEndpoint().GetAddress()
					if address.GetSocketAddress() != nil {
						ed.Target = address.GetSocketAddress().GetAddress()
						ed.Port = address.GetSocketAddress().GetPortValue()
					}
					if address.GetPipe() != nil {
						ed.UnixDomainPath = address.GetPipe().GetPath()
					}
					if localityLbEndpoint.Locality != nil && localityLbEndpoint.Locality.Zone != localZone && ed.Locality == nil{
						ed.Locality = &core_xds.Locality{
							Zone: localityLbEndpoint.Locality.Zone,
							Priority: 1,
						}
					}
					endpointsList = append(endpointsList, ed)
				}
			}
		}
		cla, err := envoy_endpoints.CreateClusterLoadAssignment(serviceName, endpointsList, proxy.APIVersion)
		if err != nil {
			return err
		}
		core.Log.Info("test print of endpoints list", "list", endpointsList)
		rs.Add(&core_xds.Resource{
			Name:     serviceName,
			Resource: cla,
		})

		core.Log.Info("MLBS resources set", "rs", rs)
	}

	// ctx.ControlPlane.CLACache.GetCLA
	if conf.LocalityAwareness == nil || !pointer.Deref(conf.LocalityAwareness.Disabled) {
		for _, cla := range endpoints[serviceName] {
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
	ctx xds_context.Context,
	proxy *core_xds.Proxy,
	rules core_rules.ToRules,
	gatewayListeners map[core_rules.InboundListener]*envoy_listener.Listener,
	gatewayClusters map[string]*envoy_cluster.Cluster,
	gatewayRoutes map[string]*envoy_route.RouteConfiguration,
	endpoints policies_xds.EndpointMap,
	rs *core_xds.ResourceSet,
) error {
	if !proxy.Dataplane.Spec.IsBuiltinGateway() {
		return nil
	}

	gatewayListenerInfos, err := gateway_plugin.GatewayListenerInfoFromProxy(context.TODO(), ctx.Mesh, proxy, ctx.ControlPlane.Zone)
	if err != nil {
		return err
	}

	conf := core_rules.ComputeConf[api.Conf](rules.Rules, core_rules.MeshSubset())
	if conf == nil {
		return nil
	}

	for _, listenerInfo := range gatewayListenerInfos {
		listener, ok := gatewayListeners[core_rules.InboundListener{
			Address: proxy.Dataplane.Spec.GetNetworking().GetAddress(),
			Port:    listenerInfo.Listener.Port,
		}]
		if !ok {
			continue
		}

		if err := p.configureListener(listener, gatewayRoutes, conf.LoadBalancer); err != nil {
			return err
		}

		for _, hostInfo := range listenerInfo.HostInfos {
			destinations := gateway_plugin.RouteDestinationsMutable(hostInfo.Entries)
			for _, dest := range destinations {
				clusterName, err := dest.Destination.DestinationClusterName(hostInfo.Host.Tags)
				if err != nil {
					continue
				}
				cluster, ok := gatewayClusters[clusterName]
				if !ok {
					continue
				}

				if err := p.configureCluster(cluster, conf); err != nil {
					return err
				}

				serviceName := dest.Destination[mesh_proto.ServiceTag]
				if err := configureEndpoints(proxy, endpoints, serviceName, *conf, ctx, rs); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (p plugin) configureEgress(rs *core_xds.ResourceSet, proxy *core_xds.Proxy) error {
	localityAwareExternalServices := map[string][]core_xds.ServiceName{}

	for _, mr := range proxy.ZoneEgressProxy.MeshResourcesList {
		for es, dynamic := range mr.Dynamic {
			policies, ok := dynamic[api.MeshLoadBalancingStrategyType]
			if !ok {
				continue
			}

			if !p.isLocalityAware(policies.FromRules) {
				continue
			}

			meshName := mr.Mesh.GetMeta().GetName()
			localityAwareExternalServices[meshName] = append(localityAwareExternalServices[meshName], es)
		}
	}

	endpoints := policies_xds.GatherEgressEndpoints(rs)
	zone := proxy.ZoneEgressProxy.ZoneEgressResource.Spec.GetZone()

	for meshName, externalServices := range localityAwareExternalServices {
		for _, es := range externalServices {
			clusterName := envoy_names.GetMeshClusterName(meshName, es)
			cla, ok := endpoints[clusterName]
			if !ok {
				continue
			}
			for _, localityLbEndpoints := range cla.Endpoints {
				if localityLbEndpoints.Locality != nil && localityLbEndpoints.Locality.Zone != zone {
					localityLbEndpoints.Priority = 1
				}
			}
		}
	}

	return nil
}

// Zone egress is a single point for multiple clients. At this moment we don't support different
// configurations based on the client, that's why locality awareness is enabled if at least one
// client requires it to be enabled.
func (p plugin) isLocalityAware(fr core_rules.FromRules) bool {
	for _, rules := range fr.Rules {
		for _, r := range rules {
			conf := r.Conf.(api.Conf)
			if conf.LocalityAwareness == nil || !pointer.Deref(conf.LocalityAwareness.Disabled) {
				return true
			}
		}
	}
	return false
}

func (p plugin) configureListener(
	l *envoy_listener.Listener,
	routes map[string]*envoy_route.RouteConfiguration,
	lbConf *api.LoadBalancer,
) error {
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
}

func (p plugin) configureCluster(c *envoy_cluster.Cluster, config *api.Conf) error {
	if config == nil {
		return nil
	}
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

func shouldUseLocalityWeightedLb(config *api.Conf) bool {
	return config.LocalityAwareness != nil && config.LocalityAwareness.LocalZone != nil && len(config.LocalityAwareness.LocalZone.AffinityTags) > 0
}