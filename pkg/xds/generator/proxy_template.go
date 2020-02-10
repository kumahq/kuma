package generator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"

	kuma_mesh "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/validators"
	model "github.com/Kong/kuma/pkg/core/xds"
	util_envoy "github.com/Kong/kuma/pkg/util/envoy"
	xds_context "github.com/Kong/kuma/pkg/xds/context"

	envoy_clusters "github.com/Kong/kuma/pkg/xds/envoy/clusters"
	envoy_endpoints "github.com/Kong/kuma/pkg/xds/envoy/endpoints"
	envoy_listeners "github.com/Kong/kuma/pkg/xds/envoy/listeners"
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
		localClusterName := localClusterName(endpoint.WorkloadPort)
		resources.Add(&model.Resource{
			Name:     localClusterName,
			Version:  "",
			Resource: envoy_clusters.CreateLocalCluster(localClusterName, "127.0.0.1", endpoint.WorkloadPort),
		})

		// generate LDS resource
		protocol := proxy.Dataplane.GetProtocol(i)
		inboundListenerName := localListenerName(endpoint.DataplaneIP, endpoint.DataplanePort)
		inboundListener, err := envoy_listeners.NewListenerBuilder().
			Configure(envoy_listeners.InboundListener(inboundListenerName, endpoint.DataplaneIP, endpoint.DataplanePort)).
			Configure(envoy_listeners.ServerSideMTLS(ctx, proxy.Metadata)).
			Configure(g.protocolSpecificOpts(protocol, envoy_listeners.ClusterInfo{Name: localClusterName})...).
			Configure(envoy_listeners.NetworkRBAC(ctx.Mesh.Resource.Spec.GetMtls().GetEnabled(), proxy.TrafficPermissions.Get(endpoint.String()))).
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

func (_ InboundProxyGenerator) protocolSpecificOpts(protocol mesh_core.Protocol, localCluster envoy_listeners.ClusterInfo) []envoy_listeners.ListenerBuilderOpt {
	switch protocol {
	case mesh_core.ProtocolHTTP:
		return []envoy_listeners.ListenerBuilderOpt{
			envoy_listeners.HttpConnectionManager(localCluster.Name),
			envoy_listeners.HttpInboundRoute(localCluster),
		}
	case mesh_core.ProtocolTCP:
		fallthrough
	default:
		return []envoy_listeners.ListenerBuilderOpt{
			envoy_listeners.TcpProxy(localCluster.Name, localCluster),
		}
	}
}

type OutboundProxyGenerator struct {
}

func (g OutboundProxyGenerator) Generate(ctx xds_context.Context, proxy *model.Proxy) ([]*model.Resource, error) {
	ofaces := proxy.Dataplane.Spec.Networking.GetOutbound()
	if len(ofaces) == 0 {
		return nil, nil
	}
	resources := &model.ResourceSet{}
	sourceService := proxy.Dataplane.Spec.GetIdentifyingService()
	for i, oface := range ofaces {
		endpoint, err := kuma_mesh.ParseOutboundInterface(oface.Interface)
		if err != nil {
			return nil, errors.Wrapf(err, "%s: value is not valid: %q", validators.RootedAt("dataplane").Field("networking").Field("outbound").Index(i).Field("interface"), oface.Interface)
		}

		// pick a route
		route := proxy.TrafficRoutes[oface.Service]
		if route == nil {
			return nil, errors.Errorf("%s{service=%q}: has no TrafficRoute", validators.RootedAt("dataplane").Field("networking").Field("outbound").Index(i), oface.Service)
		}

		// determine the list of destination clusters
		clusters, err := g.determineClusters(ctx, proxy, route)
		if err != nil {
			return nil, err
		}

		// generate CDS and EDS resources
		resources.Add(g.generateEds(ctx, proxy, clusters)...)

		// generate LDS resource
		outboundListenerName := fmt.Sprintf("outbound:%s:%d", endpoint.DataplaneIP, endpoint.DataplanePort)
		destinationService := oface.Service

		listener, err := envoy_listeners.NewListenerBuilder().
			Configure(envoy_listeners.OutboundListener(outboundListenerName, endpoint.DataplaneIP, endpoint.DataplanePort)).
			Configure(envoy_listeners.TcpProxy(oface.Service, clusters...)).
			Configure(envoy_listeners.NetworkAccessLog(sourceService, destinationService, proxy.Logs[oface.Service], proxy)).
			Configure(envoy_listeners.TransparentProxying(proxy.Dataplane.Spec.Networking.GetTransparentProxying())).
			Build()
		if err != nil {
			return nil, errors.Wrapf(err, "%s: could not generate listener %s", validators.RootedAt("dataplane").Field("networking").Field("outbound").Index(i), outboundListenerName)
		}
		resources.Add(&model.Resource{
			Name:     outboundListenerName,
			Resource: listener,
		})
	}
	return resources.List(), nil
}

func (_ OutboundProxyGenerator) determineClusters(ctx xds_context.Context, proxy *model.Proxy, route *mesh_core.TrafficRouteResource) (clusters []envoy_listeners.ClusterInfo, err error) {
	for j, destination := range route.Spec.Conf {
		service, ok := destination.Destination[kuma_mesh.ServiceTag]
		if !ok {
			return nil, errors.Errorf("trafficroute{name=%q}.%s: mandatory tag %q is missing: %v", route.GetMeta().GetName(), validators.RootedAt("conf").Index(j).Field("destination"), kuma_mesh.ServiceTag, destination.Destination)
		}
		if destination.Weight == 0 {
			// Envoy doesn't support 0 weight
			continue
		}
		clusters = append(clusters, envoy_listeners.ClusterInfo{
			Name:   destinationClusterName(service, destination.Destination),
			Weight: destination.Weight,
			Tags:   destination.Destination,
		})
	}
	return
}

func (_ OutboundProxyGenerator) generateEds(ctx xds_context.Context, proxy *model.Proxy, clusters []envoy_listeners.ClusterInfo) (resources []*model.Resource) {
	for _, cluster := range clusters {
		serviceName := cluster.Tags[kuma_mesh.ServiceTag]
		healthCheck := proxy.HealthChecks[serviceName]
		resources = append(resources, &model.Resource{
			Name:     cluster.Name,
			Resource: envoy_clusters.ClusterWithHealthChecks(envoy_clusters.CreateEdsCluster(ctx, cluster.Name, proxy.Metadata), healthCheck),
		})
		endpoints := model.EndpointList(proxy.OutboundTargets[serviceName]).Filter(kuma_mesh.MatchTags(cluster.Tags))
		resources = append(resources, &model.Resource{
			Name:     cluster.Name,
			Resource: envoy_endpoints.CreateClusterLoadAssignment(cluster.Name, endpoints),
		})
	}
	return
}

type TransparentProxyGenerator struct {
}

func (_ TransparentProxyGenerator) Generate(ctx xds_context.Context, proxy *model.Proxy) ([]*model.Resource, error) {
	redirectPort := proxy.Dataplane.Spec.Networking.GetTransparentProxying().GetRedirectPort()
	if redirectPort == 0 {
		return nil, nil
	}
	listener, err := envoy_listeners.NewListenerBuilder().
		Configure(envoy_listeners.OutboundListener("catch_all", "0.0.0.0", redirectPort)).
		Configure(envoy_listeners.TcpProxy("pass_through", envoy_listeners.ClusterInfo{Name: "pass_through"})).
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

func destinationClusterName(service string, selector map[string]string) string {
	var pairs []string
	for key, value := range selector {
		if key == kuma_mesh.ServiceTag {
			continue
		}
		pairs = append(pairs, fmt.Sprintf("%s=%s", key, value))
	}
	if len(pairs) == 0 {
		return service
	}
	sort.Strings(pairs)
	return fmt.Sprintf("%s{%s}", service, strings.Join(pairs, ","))
}
