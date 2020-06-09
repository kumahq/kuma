package generator

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/Kong/kuma/api/mesh/v1alpha1"
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
	IngressPort  = 10001
)

type IngressGenerator struct {
}

func (i IngressGenerator) Generate(ctx xds_context.Context, proxy *model.Proxy) ([]*model.Resource, error) {
	ingress := proxy.Dataplane
	resources := &model.ResourceSet{}

	inbound := ingress.Spec.Networking.Inbound[0]
	inboundListenerName := envoy_names.GetInboundListenerName(ingress.Spec.GetNetworking().GetAddress(), inbound.Port)
	inboundListenerBuilder := envoy_listeners.NewListenerBuilder().
		Configure(envoy_listeners.InboundListenerTLSInspector(inboundListenerName, ingress.Spec.GetNetworking().GetAddress(), inbound.Port))

	if len(ingress.Spec.Networking.Ingress) == 0 {
		inboundListenerBuilder = inboundListenerBuilder.
			Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder()))
	}
	for _, inbound := range ingress.Spec.Networking.Ingress {
		for _, perm := range TagPermutations(inbound.Service, inbound.GetTags()) {
			inboundListenerBuilder = inboundListenerBuilder.
				Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder().
					Configure(envoy_listeners.FilterChainMatch(perm)).
					Configure(envoy_listeners.TcpProxyWithMetaMatch(inbound.Service, envoy_common.ClusterInfo{
						Name: inbound.Service,
						Tags: TagsBySNI(perm),
					}))))
		}
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
		fmt.Println("generateEds ", service, endpoints)
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
			v1alpha1.ServiceTag: sni,
		}
	}
	service, tags := matches[1], matches[2]
	pairs := strings.Split(tags, ",")
	rv := map[string]string{
		v1alpha1.ServiceTag: service,
	}
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		rv[kv[0]] = kv[1]
	}
	return rv
}
