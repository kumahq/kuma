package generator

import (
	"fmt"

	"github.com/pkg/errors"

	kuma_mesh "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/validators"
	model "github.com/Kong/kuma/pkg/core/xds"
	util_envoy "github.com/Kong/kuma/pkg/util/envoy"
	xds_context "github.com/Kong/kuma/pkg/xds/context"

	envoy_common "github.com/Kong/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/Kong/kuma/pkg/xds/envoy/clusters"
	envoy_endpoints "github.com/Kong/kuma/pkg/xds/envoy/endpoints"
	envoy_listeners "github.com/Kong/kuma/pkg/xds/envoy/listeners"
	envoy_names "github.com/Kong/kuma/pkg/xds/envoy/names"
	envoy_routes "github.com/Kong/kuma/pkg/xds/envoy/routes"
)

type TemplateProxyGenerator struct {
	ProxyTemplate *kuma_mesh.ProxyTemplate
}

func (g *TemplateProxyGenerator) Generate(ctx xds_context.Context, proxy *model.Proxy) ([]*model.Resource, error) {
	resources := make([]*model.Resource, 0, len(g.ProxyTemplate.GetConf().GetImports())+1)
	for i, name := range g.ProxyTemplate.GetConf().GetImports() {
		generator := &ProxyTemplateProfileSource{ProfileName: name}
		if rs, err := generator.Generate(ctx, proxy); err != nil {
			return nil, fmt.Errorf("imports[%d]{name=%q}: %s", i, name, err)
		} else {
			resources = append(resources, rs...)
		}
	}
	generator := &ProxyTemplateRawSource{Resources: g.ProxyTemplate.GetConf().GetResources()}
	if rs, err := generator.Generate(ctx, proxy); err != nil {
		return nil, fmt.Errorf("resources: %s", err)
	} else {
		resources = append(resources, rs...)
	}
	return resources, nil
}

type ProxyTemplateRawSource struct {
	Resources []*kuma_mesh.ProxyTemplateRawResource
}

func (s *ProxyTemplateRawSource) Generate(_ xds_context.Context, proxy *model.Proxy) ([]*model.Resource, error) {
	resources := make([]*model.Resource, 0, len(s.Resources))
	for i, r := range s.Resources {
		res, err := util_envoy.ResourceFromYaml(r.Resource)
		if err != nil {
			return nil, fmt.Errorf("raw.resources[%d]{name=%q}.resource: %s", i, r.Name, err)
		}

		resources = append(resources, &model.Resource{
			Name:     r.Name,
			Version:  r.Version,
			Resource: res,
		})
	}
	return resources, nil
}

var predefinedProfiles = make(map[string]ResourceGenerator)

func NewDefaultProxyProfile() ResourceGenerator {
	return CompositeResourceGenerator{PrometheusEndpointGenerator{}, TransparentProxyGenerator{}, InboundProxyGenerator{}, OutboundProxyGenerator{}}
}

func init() {
	predefinedProfiles[mesh_core.ProfileDefaultProxy] = NewDefaultProxyProfile()
}

type ProxyTemplateProfileSource struct {
	ProfileName string
}

func (s *ProxyTemplateProfileSource) Generate(ctx xds_context.Context, proxy *model.Proxy) ([]*model.Resource, error) {
	g, ok := predefinedProfiles[s.ProfileName]
	if !ok {
		return nil, fmt.Errorf("profile{name=%q}: unknown profile", s.ProfileName)
	}
	return g.Generate(ctx, proxy)
}

type InboundProxyGenerator struct {
}

func (g InboundProxyGenerator) Generate(ctx xds_context.Context, proxy *model.Proxy) ([]*model.Resource, error) {
	endpoints, err := proxy.Dataplane.Spec.Networking.GetInboundInterfaces()
	if err != nil {
		return nil, err
	}
	if len(endpoints) == 0 {
		return nil, nil
	}
	resources := &model.ResourceSet{}
	for i, endpoint := range endpoints {
		// generate CDS resource
		localClusterName := envoy_names.GetLocalClusterName(endpoint.WorkloadPort)
		resources.Add(&model.Resource{
			Name:     localClusterName,
			Version:  "",
			Resource: envoy_clusters.CreateLocalCluster(localClusterName, "127.0.0.1", endpoint.WorkloadPort),
		})

		// generate LDS resource
		iface := proxy.Dataplane.Spec.Networking.Inbound[i]
		service := iface.GetService()
		protocol := mesh_core.ParseProtocol(iface.GetProtocol())
		inboundListenerName := envoy_names.GetInboundListenerName(endpoint.DataplaneIP, endpoint.DataplanePort)
		filterChainBuilder := func() *envoy_listeners.FilterChainBuilder {
			filterChainBuilder := envoy_listeners.NewFilterChainBuilder()
			switch protocol {
			case mesh_core.ProtocolHTTP:
				// configuration for HTTP case
				filterChainBuilder.
					Configure(envoy_listeners.HttpConnectionManager(localClusterName)).
					Configure(envoy_listeners.FaultInjection(proxy.FaultInjections[endpoint])).
					Configure(envoy_listeners.Tracing(proxy.TracingBackend)).
					Configure(envoy_listeners.HttpInboundRoute(service, envoy_common.ClusterInfo{Name: localClusterName}))
			case mesh_core.ProtocolTCP:
				fallthrough
			default:
				// configuration for non-HTTP cases
				filterChainBuilder.Configure(envoy_listeners.TcpProxy(localClusterName, envoy_common.ClusterInfo{Name: localClusterName}))
			}
			return filterChainBuilder.
				Configure(envoy_listeners.ServerSideMTLS(ctx, proxy.Metadata)).
				Configure(envoy_listeners.NetworkRBAC(inboundListenerName, ctx.Mesh.Resource.MTLSEnabled(), proxy.TrafficPermissions[endpoint]))
		}()
		inboundListener, err := envoy_listeners.NewListenerBuilder().
			Configure(envoy_listeners.InboundListener(inboundListenerName, endpoint.DataplaneIP, endpoint.DataplanePort)).
			Configure(envoy_listeners.FilterChain(filterChainBuilder)).
			Configure(envoy_listeners.TransparentProxying(proxy.Dataplane.Spec.Networking.GetTransparentProxying())).
			Build()
		if err != nil {
			return nil, errors.Wrapf(err, "%s: could not generate listener %s", validators.RootedAt("dataplane").Field("networking").Field("inbound").Index(i), inboundListenerName)
		}
		resources.Add(&model.Resource{
			Name:     inboundListenerName,
			Version:  "",
			Resource: inboundListener,
		})
	}
	return resources.List(), nil
}

type OutboundProxyGenerator struct {
}

func (g OutboundProxyGenerator) Generate(ctx xds_context.Context, proxy *model.Proxy) ([]*model.Resource, error) {
	outbounds := proxy.Dataplane.Spec.Networking.GetOutbound()
	if len(outbounds) == 0 {
		return nil, nil
	}
	resources := &model.ResourceSet{}
	sourceService := proxy.Dataplane.Spec.GetIdentifyingService()
	meshName := ctx.Mesh.Resource.GetMeta().GetName()
	ofaces, err := proxy.Dataplane.Spec.Networking.GetOutboundInterfaces()
	if err != nil {
		return nil, err
	}
	for i, outbound := range outbounds {
		// pick a route
		route := proxy.TrafficRoutes[outbound.Service]
		if route == nil {
			return nil, errors.Errorf("%s{service=%q}: has no TrafficRoute", validators.RootedAt("dataplane").Field("networking").Field("outbound").Index(i), outbound.Service)
		}

		// determine the list of destination clusters
		clusters, err := g.determineClusters(ctx, proxy, route)
		if err != nil {
			return nil, err
		}

		// generate CDS and EDS resources
		edsResources, endpoints, err := g.generateEds(ctx, proxy, clusters)
		if err != nil {
			return nil, err
		}
		resources.Add(edsResources...)

		protocol := InferServiceProtocol(endpoints)

		// generate LDS resource
		outboundListenerName := envoy_names.GetOutboundListenerName(ofaces[i].DataplaneIP, ofaces[i].DataplanePort)
		outboundRouteName := envoy_names.GetOutboundRouteName(outbound.Service)
		destinationService := outbound.Service

		filterChainBuilder := func() *envoy_listeners.FilterChainBuilder {
			filterChainBuilder := envoy_listeners.NewFilterChainBuilder()
			switch protocol {
			case mesh_core.ProtocolHTTP:
				// configuration for HTTP case
				filterChainBuilder.
					Configure(envoy_listeners.HttpConnectionManager(outbound.Service)).
					Configure(envoy_listeners.Tracing(proxy.TracingBackend)).
					Configure(envoy_listeners.HttpAccessLog(meshName, sourceService, destinationService, proxy.Logs[outbound.Service], proxy)).
					Configure(envoy_listeners.HttpOutboundRoute(outboundRouteName))
			case mesh_core.ProtocolTCP:
				fallthrough
			default:
				// configuration for non-HTTP cases
				filterChainBuilder.
					Configure(envoy_listeners.TcpProxy(outbound.Service, clusters...)).
					Configure(envoy_listeners.NetworkAccessLog(meshName, sourceService, destinationService, proxy.Logs[outbound.Service], proxy))
			}
			return filterChainBuilder
		}()
		listener, err := envoy_listeners.NewListenerBuilder().
			Configure(envoy_listeners.OutboundListener(outboundListenerName, ofaces[i].DataplaneIP, ofaces[i].DataplanePort)).
			Configure(envoy_listeners.FilterChain(filterChainBuilder)).
			Configure(envoy_listeners.TransparentProxying(proxy.Dataplane.Spec.Networking.GetTransparentProxying())).
			Build()
		if err != nil {
			return nil, errors.Wrapf(err, "%s: could not generate listener %s", validators.RootedAt("dataplane").Field("networking").Field("outbound").Index(i), outboundListenerName)
		}
		resources.Add(&model.Resource{
			Name:     outboundListenerName,
			Resource: listener,
		})

		// generate RDS resources
		rdsResources, err := g.generateRds(protocol, outbound.Service, outboundRouteName, clusters, proxy.Dataplane.Spec.Tags())
		if err != nil {
			return nil, err
		}
		resources.Add(rdsResources...)
	}

	return resources.List(), nil
}

func (_ OutboundProxyGenerator) determineClusters(ctx xds_context.Context, proxy *model.Proxy, route *mesh_core.TrafficRouteResource) (clusters []envoy_common.ClusterInfo, err error) {
	for j, destination := range route.Spec.Conf {
		service, ok := destination.Destination[kuma_mesh.ServiceTag]
		if !ok {
			return nil, errors.Errorf("trafficroute{name=%q}.%s: mandatory tag %q is missing: %v", route.GetMeta().GetName(), validators.RootedAt("conf").Index(j).Field("destination"), kuma_mesh.ServiceTag, destination.Destination)
		}
		if destination.Weight == 0 {
			// Envoy doesn't support 0 weight
			continue
		}
		clusters = append(clusters, envoy_common.ClusterInfo{
			Name:   envoy_names.GetDestinationClusterName(service, destination.Destination),
			Weight: destination.Weight,
			Tags:   destination.Destination,
		})
	}
	return
}

func (_ OutboundProxyGenerator) generateEds(ctx xds_context.Context, proxy *model.Proxy, clusters []envoy_common.ClusterInfo) (resources []*model.Resource, allEndpoints []model.Endpoint, _ error) {
	for _, cluster := range clusters {
		serviceName := cluster.Tags[kuma_mesh.ServiceTag]
		healthCheck := proxy.HealthChecks[serviceName]
		edsCluster, err := envoy_clusters.CreateEdsCluster(ctx, cluster.Name, proxy.Metadata)
		if err != nil {
			return nil, nil, err
		}
		resources = append(resources, &model.Resource{
			Name:     cluster.Name,
			Resource: envoy_clusters.ClusterWithHealthChecks(edsCluster, healthCheck),
		})
		endpoints := model.EndpointList(proxy.OutboundTargets[serviceName]).Filter(kuma_mesh.MatchTags(cluster.Tags))
		resources = append(resources, &model.Resource{
			Name:     cluster.Name,
			Resource: envoy_endpoints.CreateClusterLoadAssignment(cluster.Name, endpoints),
		})
		allEndpoints = append(allEndpoints, endpoints...)
	}
	return
}

func (_ OutboundProxyGenerator) generateRds(protocol mesh_core.Protocol, service string, outboundRouteName string, clusters []envoy_common.ClusterInfo, tags kuma_mesh.MultiValueTagSet) ([]*model.Resource, error) {
	resources := &model.ResourceSet{}
	switch protocol {
	case mesh_core.ProtocolHTTP:
		// generate RDS resource
		routeConfiguration, err := envoy_routes.NewRouteConfigurationBuilder().
			Configure(envoy_routes.CommonRouteConfiguration(outboundRouteName)).
			Configure(envoy_routes.TagsHeader(tags)).
			Configure(envoy_routes.VirtualHost(envoy_routes.NewVirtualHostBuilder().
				Configure(envoy_routes.CommonVirtualHost(service)).
				Configure(envoy_routes.DefaultRoute(clusters...)))).
			Build()
		if err != nil {
			return nil, err
		}
		resources.Add(&model.Resource{
			Name:     outboundRouteName,
			Resource: routeConfiguration,
		})
	}
	return resources.List(), nil
}

type TransparentProxyGenerator struct {
}

func (_ TransparentProxyGenerator) Generate(ctx xds_context.Context, proxy *model.Proxy) ([]*model.Resource, error) {
	redirectPort := proxy.Dataplane.Spec.Networking.GetTransparentProxying().GetRedirectPort()
	if redirectPort == 0 {
		return nil, nil
	}
	sourceService := proxy.Dataplane.Spec.GetIdentifyingService()
	meshName := ctx.Mesh.Resource.GetMeta().GetName()
	listener, err := envoy_listeners.NewListenerBuilder().
		Configure(envoy_listeners.OutboundListener("catch_all", "0.0.0.0", redirectPort)).
		Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder().
			Configure(envoy_listeners.TcpProxy("pass_through", envoy_common.ClusterInfo{Name: "pass_through"})).
			Configure(envoy_listeners.NetworkAccessLog(meshName, sourceService, "external", proxy.Logs[mesh_core.PassThroughService], proxy)))).
		Configure(envoy_listeners.OriginalDstForwarder()).
		Build()
	if err != nil {
		return nil, errors.Wrapf(err, "could not generate listener %s", "catch_all")
	}
	return []*model.Resource{
		&model.Resource{
			Name:     "catch_all",
			Version:  proxy.Dataplane.Meta.GetVersion(),
			Resource: listener,
		},
		&model.Resource{
			Name:     "pass_through",
			Version:  proxy.Dataplane.Meta.GetVersion(),
			Resource: envoy_clusters.CreatePassThroughCluster("pass_through"),
		},
	}, nil
}
