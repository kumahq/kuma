package generator

import (
	"fmt"
	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	model "github.com/Kong/kuma/pkg/core/xds"
	xds_context "github.com/Kong/kuma/pkg/xds/context"
	envoy_common "github.com/Kong/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/Kong/kuma/pkg/xds/envoy/clusters"
	envoy_endpoints "github.com/Kong/kuma/pkg/xds/envoy/endpoints"
	envoy_listeners "github.com/Kong/kuma/pkg/xds/envoy/listeners"
	envoy_names "github.com/Kong/kuma/pkg/xds/envoy/names"
	"sort"
	"strings"
)

const (
	IngressProxy = "ingress-proxy"
	IngressPort  = 10001
)

type IngressGenerator struct {
}

func (i IngressGenerator) Generate(ctx xds_context.Context, proxy *model.Proxy) ([]*model.Resource, error) {
	ingress := proxy.Dataplane
	resources := &model.ResourceSet{}

	// todo: create listener from inbounds
	inboundListenerName := envoy_names.GetInboundListenerName(ingress.Spec.GetNetworking().GetAddress(), IngressPort)
	inboundListenerBuilder := envoy_listeners.NewListenerBuilder().
		Configure(envoy_listeners.InboundListener(inboundListenerName, ingress.Spec.GetNetworking().GetAddress(), IngressPort))

	for _, inbound := range ingress.Spec.Networking.Inbound {
		service := inbound.GetTags()[mesh_proto.ServiceTag]
		inboundListenerBuilder = inboundListenerBuilder.
			Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder().
				Configure(envoy_listeners.FilterChainMatch(tagPermutations(service, inbound.GetTags())...)).
				Configure(envoy_listeners.TcpProxyWithMetaMatch(service, envoy_common.ClusterInfo{Name: service, Tags: inbound.GetTags()})))) // todo: consider 'statsName'
	}

	listener, err := inboundListenerBuilder.Build()
	if err != nil {
		return nil, err
	}
	resources.AddNamed(listener)

	edsResources, err := generateEds(proxy)
	if err != nil {
		return nil, err
	}
	resources.Add(edsResources...)

	return resources.List(), nil
}

func generateEds(proxy *model.Proxy) (resources []*model.Resource, _ error) {
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

func tagPermutations(service string, tags map[string]string) []string {
	pairs := []string{}
	for k, v := range tags {
		if k == mesh_proto.ServiceTag {
			continue
		}
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(pairs)
	rv := []string{}
	for _, tagSet := range append(Permutation(pairs), []string{}) {
		rv = append(rv, fmt.Sprintf("%s{%s}", service, strings.Join(tagSet, ",")))
	}
	return rv
}
