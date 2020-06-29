package generator

import (
	"fmt"
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
	envoy_tls "github.com/Kong/kuma/pkg/xds/envoy/tls"
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

	cdsResources, err := i.generateCDS(services, proxy)
	if err != nil {
		return nil, err
	}
	resources.Add(cdsResources...)

	edsResources := i.generateEDS(proxy, services)
	resources.Add(edsResources...)

	return resources.List(), nil
}

// generateLDS generates one Ingress Listener
// Ingress Listener assumes that mTLS is on. Using TLSInspector we sniff SNI value.
// SNI value has service name and tag values specified with the following format: "backend{cluster=2,version=1}"
// Given every unique permutation of tags in available services in the cluster we generate filter chain match based on the SNI and then use subset load balancing to pick endpoints with tags from SNI
// Traffic is NOT decrypted here, therefore we don't need certificates and mTLS settings
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
						Tags:        envoy_tls.TagsFromSNI(perm),
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

func (i IngressGenerator) generateCDS(services []string, proxy *model.Proxy) (resources []*model.Resource, _ error) {
	for _, service := range services {
		edsCluster, err := envoy_clusters.NewClusterBuilder().
			Configure(envoy_clusters.EdsCluster(service)).
			Configure(envoy_clusters.LbSubset(i.lbSubsets(proxy.OutboundTargets[service]))).
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

// lbSubsets generate subsets given endpoints list.
// Example:
// given endpoints:
// - service: backend, cloud: aws, version: 1
// - service: backend, cloud: gcp
// we generate permutations: [[backend], [cloud], [version], [backend, cloud], [backend, version], [cloud, version], [backend, version, cloud]]
// because we don't know by which tags will be used in SNI by the client
func (_ IngressGenerator) lbSubsets(endpoints model.EndpointList) [][]string {
	uniqueKeys := map[string]bool{}
	for _, endpoint := range endpoints {
		for key := range envoy_common.Tags(endpoint.Tags).WithoutTag(mesh_proto.ServiceTag) {
			uniqueKeys[key] = true
		}
	}
	var keys []string
	for key := range uniqueKeys {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return Permutation(keys)
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
