package generator

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
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
	ingress := proxy.Dataplane
	resources := &model.ResourceSet{}

	inbound := ingress.Spec.Networking.Inbound[0]
	inboundListenerName := envoy_names.GetInboundListenerName(ingress.Spec.GetNetworking().GetAddress(), inbound.Port)
	inboundListenerBuilder := envoy_listeners.NewListenerBuilder().
		Configure(envoy_listeners.InboundListener(inboundListenerName, ingress.Spec.GetNetworking().GetAddress(), inbound.Port)).
		Configure(envoy_listeners.TLSInspector())

	if !ingress.Spec.HasAvailableServices() {
		inboundListenerBuilder = inboundListenerBuilder.
			Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder()))
	}
	for _, inbound := range ingress.Spec.GetNetworking().GetIngress().GetAvailableServices() {
		service := inbound.Tags[mesh_proto.ServiceTag]
		permutations := TagPermutations(service, mesh_proto.SingleValueTagSet(inbound.GetTags()).Exclude(mesh_proto.ServiceTag))
		for _, perm := range permutations {
			inboundListenerBuilder = inboundListenerBuilder.
				Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder().
					Configure(envoy_listeners.FilterChainMatch(perm)).
					Configure(envoy_listeners.TcpProxy(service, envoy_common.ClusterSubset{
						ClusterName: service,
						Tags:        TagsBySNI(perm),
					}))))
		}
	}
	listener, err := inboundListenerBuilder.Build()
	if err != nil {
		return nil, err
	}
	resources.AddNamed(listener)

	edsResources, err := i.generateEds(proxy)
	if err != nil {
		return nil, err
	}
	resources.Add(edsResources...)

	return resources.List(), nil
}

func (_ IngressGenerator) generateEds(proxy *model.Proxy) (resources []*model.Resource, _ error) {
	for service, endpoints := range proxy.OutboundTargets {
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
