package generator

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	envoy_api_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	model "github.com/Kong/kuma/pkg/core/xds"
	xds_context "github.com/Kong/kuma/pkg/xds/context"
	envoy_common "github.com/Kong/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/Kong/kuma/pkg/xds/envoy/clusters"
	envoy_endpoints "github.com/Kong/kuma/pkg/xds/envoy/endpoints"
	envoy_listeners "github.com/Kong/kuma/pkg/xds/envoy/listeners"
	envoy_names "github.com/Kong/kuma/pkg/xds/envoy/names"
)

const (
	IngressProxy = "ingress-proxy"
)

type IngressGenerator struct {
}

func (i IngressGenerator) Generate(ctx xds_context.Context, proxy *model.Proxy) ([]*model.Resource, error) {
	resources := &model.ResourceSet{}

	listener, err := i.generateLDS(proxy.Dataplane)
	if err != nil {
		return nil, err
	}
	resources.AddNamed(listener)

	services := i.services(proxy)

	cdsResources, err := i.generateCDS(services)
	if err != nil {
		return nil, err
	}
	resources.Add(cdsResources...)

	edsResources := i.generateEDS(proxy, services)
	resources.Add(edsResources...)

	return resources.List(), nil
}

func (_ IngressGenerator) generateLDS(ingress *core_mesh.DataplaneResource) (*envoy_api_v2.Listener, error) {
	inbound := ingress.Spec.Networking.Inbound[0]
	inboundListenerName := envoy_names.GetInboundListenerName(ingress.Spec.GetNetworking().GetAddress(), inbound.Port)
	inboundListenerBuilder := envoy_listeners.NewListenerBuilder().
		Configure(envoy_listeners.InboundListener(inboundListenerName, ingress.Spec.GetNetworking().GetAddress(), inbound.Port)).
		Configure(envoy_listeners.TLSInspector())

	if !ingress.Spec.HasAvailableServices() {
		inboundListenerBuilder = inboundListenerBuilder.
			Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder()))
	}
	permUsed := map[string]bool{}
	for _, inbound := range ingress.Spec.GetNetworking().GetIngress().GetAvailableServices() {
		service := inbound.Tags[mesh_proto.ServiceTag]
		permutations := TagPermutations(service, mesh_proto.SingleValueTagSet(inbound.GetTags()).Exclude(mesh_proto.ServiceTag))
		for _, perm := range permutations {
			if permUsed[perm] {
				continue
			}
			permUsed[perm] = true
			inboundListenerBuilder = inboundListenerBuilder.
				Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder().
					Configure(envoy_listeners.FilterChainMatch(perm)).
					Configure(envoy_listeners.TcpProxy(service, envoy_common.ClusterSubset{
						ClusterName: service,
						Tags:        TagsBySNI(perm),
					}))))
		}
	}

	return inboundListenerBuilder.Build()
}

func (_ IngressGenerator) services(proxy *model.Proxy) []string {
	var services []string
	for service := range proxy.OutboundTargets {
		services = append(services, service)
	}
	sort.Strings(services)
	return services
}

func (_ IngressGenerator) generateCDS(services []string) (resources []*model.Resource, _ error) {
	for _, service := range services {
		edsCluster, err := envoy_clusters.NewClusterBuilder().
			Configure(envoy_clusters.EdsCluster(service)).
			Build()
		if err != nil {
			return nil, err
		}
		resources = append(resources, &model.Resource{
			Name:     service,
			Resource: edsCluster,
		})
	}
	return
}

func (_ IngressGenerator) generateEDS(proxy *model.Proxy, services []string) (resources []*model.Resource) {
	for _, service := range services {
		endpoints := proxy.OutboundTargets[service]
		resources = append(resources, &model.Resource{
			Name:     service,
			Resource: envoy_endpoints.CreateClusterLoadAssignment(service, endpoints),
		})
	}
	return
}

func TagPermutations(service string, tags map[string]string) []string {
	pairs := []string{}
	for k, v := range tags {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(pairs)
	rv := []string{}
	for _, tagSet := range Permutation(pairs) {
		rv = append(rv, fmt.Sprintf("%s{%s}", service, strings.Join(tagSet, ",")))
	}
	return append(rv, service)
}

func TagsBySNI(sni string) map[string]string {
	r := regexp.MustCompile(`(.*)\{(.*)\}`)
	matches := r.FindStringSubmatch(sni)
	if len(matches) == 0 {
		return map[string]string{
			mesh_proto.ServiceTag: sni,
		}
	}
	service, tags := matches[1], matches[2]
	pairs := strings.Split(tags, ",")
	rv := map[string]string{
		mesh_proto.ServiceTag: service,
	}
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		rv[kv[0]] = kv[1]
	}
	return rv
}
